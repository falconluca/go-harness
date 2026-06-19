package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/falconluca/go-harness/internal/provider"
	"github.com/falconluca/go-harness/internal/schema"
	"github.com/falconluca/go-harness/internal/tools"

	ctxpkg "github.com/falconluca/go-harness/internal/context"
)

type AgentEngine struct {
	provider       provider.LLMProvider
	registry       tools.Registry
	EnableThinking bool
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry,
	workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		EnableThinking: enableThinking,
	}
}

func (e *AgentEngine) Run(c context.Context, session *Session, reporter Reporter) error {
	if reporter == nil {
		return fmt.Errorf("[Engine] Reporter 不能为空")
	}

	log.Printf("[Engine] 引擎启动，锁定工作区：%s\n", session.WorkDir)
	log.Printf("[Engine] 慢思考模型：%v\n", e.EnableThinking)

	composer := ctxpkg.NewPromptComposer(session.WorkDir)
	systemMsg := composer.Build()

	turnCount := 0

	for {
		turnCount++
		log.Printf("[Engine] 第 %d 轮对话开始\n", turnCount)

		var contextHistory []schema.Message
		contextHistory = append(contextHistory, systemMsg)
		workingMemory := session.GetWorkingMemory(6)
		contextHistory = append(contextHistory, workingMemory...)

		if e.EnableThinking {
			thinkResp, err := e.provider.Generate(c, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("[Engine] 思考阶段生成失败：%w", err)
			}

			if thinkResp.Content != "" {
				log.Printf("[Engine] 🧠 Thinking... %s\n", thinkResp.Content)
				
				session.Append(*thinkResp)
				contextHistory = append(contextHistory, *thinkResp)

				reporter.OnThinking(c, thinkResp.Content)
			}
		}

		availableTools := e.registry.GetAvailableTools()
		actionResp, err := e.provider.Generate(c, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("[Engine] 行动阶段生成失败: %w", err)
		}

		session.Append(*actionResp)
		contextHistory = append(contextHistory, *actionResp)

		if actionResp.Content != "" {
			reporter.OnMessage(c, actionResp.Content)
			log.Printf("[Engine] 🤖 Speaking...: %s\n", actionResp.Content)
		}

		if len(actionResp.ToolCalls) == 0 {
			log.Println("[Engine] DONE")
			break
		}

		var wg sync.WaitGroup
		observationMsgs := make([]schema.Message, len(actionResp.ToolCalls))

		for i, toolCall := range actionResp.ToolCalls {
			wg.Add(1)

			go func(idx int, call schema.ToolCall) {
				defer wg.Done()

				reporter.OnToolCall(c, call.Name, string(call.Arguments))

				log.Printf("[Engine] 🔧 Acting...: %s, 参数: %s\n", toolCall.Name, string(toolCall.Arguments))

				result := e.registry.Execute(c, toolCall)

				reporter.OnToolResult(c, call.Name, result.DisplayOutput(), result.IsError)

				observationMsg := schema.Message{
					Role:       schema.RoleUser,
					Content:    result.Output,
					ToolCallID: toolCall.ID,
				}
				observationMsgs[idx] = observationMsg
			}(i, toolCall)
		}

		wg.Wait()

		session.Append(observationMsgs...)
	}
	return nil
}
