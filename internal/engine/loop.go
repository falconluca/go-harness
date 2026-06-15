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

	WorkDir string
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string) *AgentEngine {
	return &AgentEngine{
		provider: p,
		registry: r,
		WorkDir:  workDir,
	}
}

func (e *AgentEngine) Run(c context.Context, userPrompt string) error {
	// TODO: 实现 AgentEngine 的核心逻辑
	log.Printf("[Engine] 引擎启动，锁定工作区：%s\n", e.WorkDir)

	// 1. 初始化会话上下文
	contextHistory := []schema.Message{
		{Role: "system", Content: "你是一个智能助手，协助用户完成任务。"},
		{Role: "user", Content: userPrompt},
	}

	turnCount := 0
	for {
		turnCount++
		log.Printf("[Engine] 第 %d 轮对话开始\n", turnCount)

		availableTools := e.registry.GetAvailableTools()

		log.Println("[Engine] 正在思考...")
		responseMsg, err := e.provider.Generate(c, contextHistory, availableTools)
		if err != nil {
			// log.Printf("[Engine] 生成响应时出错: %v\n", err)
			return fmt.Errorf("模型生成失败：%w", err)
		}

		contextHistory = append(contextHistory, *responseMsg)

		if responseMsg.Content != "" {
			log.Printf("[Agent] 模型回复: %s\n", responseMsg.Content)
		}

		if len(responseMsg.ToolCalls) == 0 {
			log.Println("[Engine] 没有工具调用，结束对话")
			break
		}

		log.Printf("[Engine] 模型请求调用 %d 个工具...\n", len(responseMsg.ToolCalls))

		for _, toolCall := range responseMsg.ToolCalls {
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
