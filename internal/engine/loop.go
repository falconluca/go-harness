package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/falconluca/go-harness/internal/provider"
	"github.com/falconluca/go-harness/internal/schema"
	"github.com/falconluca/go-harness/internal/tools"
)

type AgentEngine struct {
	provider provider.LLMProvider
	registry tools.Registry

	WorkDir        string
	EnableThinking bool
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry,
	workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		WorkDir:        workDir,
		EnableThinking: enableThinking,
	}
}

func (e *AgentEngine) Run(c context.Context, userPrompt string) error {
	log.Printf("[Engine] 引擎启动，锁定工作区：%s\n", e.WorkDir)
	log.Printf("[Engine] 慢思考模型：%v\n", e.EnableThinking)

	contextHistory := []schema.Message{
		{
			Role:    schema.RoleSystem,
			Content: "你是一个智能助手，协助用户完成任务。",
		},
		{
			Role:    schema.RoleUser,
			Content: userPrompt,
		},
	}

	turnCount := 0

	for {
		turnCount++
		log.Printf("[Engine] 第 %d 轮对话开始\n", turnCount)

		availableTools := e.registry.GetAvailableTools()

		if e.EnableThinking {
			thinkResp, err := e.provider.Generate(c, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("[Engine] 思考阶段生成失败：%w", err)
			}

			if thinkResp.Content != "" {
				log.Printf("[Engine] 🧠 Thinking... %s\n", thinkResp.Content)
				contextHistory = append(contextHistory, *thinkResp)
			}
		}

		actionResp, err := e.provider.Generate(c, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("[Engine] 行动阶段生成失败: %w", err)
		}

		contextHistory = append(contextHistory, *actionResp)

		if actionResp.Content != "" {
			log.Printf("[Engine] 🤖 Speaking...: %s\n", actionResp.Content)
		}

		if len(actionResp.ToolCalls) == 0 {
			log.Println("[Engine] DONE")
			break
		}

		for _, toolCall := range actionResp.ToolCalls {
			log.Printf("[Engine] 🔧 Acting: %s, 参数: %s\n", toolCall.Name, string(toolCall.Arguments))

			result := e.registry.Execute(c, toolCall)

			if result.IsError {
				log.Printf("[Engine] ❌ 工具执行报错：%s\n", result.Output)
			}

			observationMsg := schema.Message{
				Role:       schema.RoleUser,
				Content:    result.Output,
				ToolCallID: toolCall.ID,
			}
			contextHistory = append(contextHistory, observationMsg)
		}
	}
	return nil
}
