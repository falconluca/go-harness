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

	// 1. 初始化会话上下文
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
			log.Println("[Engine] 正在思考...")

			thinkResp, err := e.provider.Generate(c, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("思考失败：%w", err)
			}

			if thinkResp.Content != "" {
				fmt.Printf("🧠 [内部思考 Trace]: %s\n", thinkResp.Content)
				contextHistory = append(contextHistory, *thinkResp)
			}
		}

		log.Println("[Engine][Phase 2] 恢复工具挂载，等待模型采取行动...")

		actionResp, err := e.provider.Generate(c, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("Action 阶段生成失败: %w", err)
		}

		contextHistory = append(contextHistory, *actionResp)

		if actionResp.Content != "" {
			log.Printf("[Agent] 模型回复: %s\n", actionResp.Content)
		}

		if len(actionResp.ToolCalls) == 0 {
			log.Println("[Engine] 没有工具调用，结束对话")
			break
		}

		log.Printf("[Engine] 模型请求调用 %d 个工具...\n", len(actionResp.ToolCalls))

		for _, toolCall := range actionResp.ToolCalls {
			log.Printf("=> 🛠️ 执行工具: %s, 参数: %s\n", toolCall.Name, string(toolCall.Arguments))

			result := e.registry.Execute(c, toolCall)

			if result.IsError {
				log.Printf("=> ❌ 工具执行报错：%s\n", result.Output)
			} else {
				log.Println("=> ✅ 工具执行成功")
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
