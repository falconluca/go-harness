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

	composer  *ctxpkg.PromptComposer
	compactor *ctxpkg.Compactor
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry,
	workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		EnableThinking: enableThinking,

		// 【初始化压缩器】：为了便于今天的极端测试，我们将水位线阈值设积极（例如 3000 字符），
		// 并保护最近的 6 条消息（大约两轮 Turn 的交互）
		compactor: ctxpkg.NewCompactor(3000, 6),
	}
}

func (e *AgentEngine) Run(c context.Context, session *Session, reporter Reporter) error {
	if reporter == nil {
		return fmt.Errorf("[Engine] Reporter 不能为空")
	}

	log.Printf("[Engine] 引擎启动，锁定工作区：%s\n", session.WorkDir)
	log.Printf("[Engine] 慢思考模型：%v\n", e.EnableThinking)

	e.composer = ctxpkg.NewPromptComposer(session.WorkDir)
	systemMsg := e.composer.Build()

	for {
		var contextHistory []schema.Message
		contextHistory = append(contextHistory, systemMsg)
		workingMemory := session.GetWorkingMemory(6)
		contextHistory = append(contextHistory, workingMemory...)

		compactedContext := e.compactor.Compact(contextHistory)

		if e.EnableThinking {
			thinkResp, err := e.provider.Generate(c, compactedContext, nil)
			if err != nil {
				return fmt.Errorf("[Engine] 思考阶段生成失败：%w", err)
			}

			if thinkResp.Content != "" {
				log.Printf("[Engine] 🧠 Thinking... %s\n", thinkResp.Content)

				session.Append(*thinkResp)
				compactedContext = append(compactedContext, *thinkResp)

				reporter.OnThinking(c, thinkResp.Content)
			}
		}

		availableTools := e.registry.GetAvailableTools()
		actionResp, err := e.provider.Generate(c, compactedContext, availableTools)
		if err != nil {
			return fmt.Errorf("[Engine] 行动阶段生成失败: %w", err)
		}

		session.Append(*actionResp)
		compactedContext = append(compactedContext, *actionResp)

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
