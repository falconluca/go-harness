package provider

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

type MockProvider struct {
	turn int
}

func (m *MockProvider) Generate(ctx context.Context, msgs []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error) {
	// 如果工具列表为空，说明这是引擎发起的 Phase 1: Thinking 阶段
	if len(tools) == 0 {
		return &schema.Message{
			Role:    schema.RoleAssistant,
			Content: "【推理中】目标是检查文件。我不能直接盲猜，我需要先调用 bash 工具执行 ls 命令，看看当前目录下有什么，然后再做定夺。",
		}, nil
	}

	// 如果工具列表不为空，说明这是 Phase 2: Action 阶段
	m.turn++
	if m.turn == 1 {
		// 第一轮 Action：顺着刚才的 Thinking，精准调用工具
		return &schema.Message{
			Role:    schema.RoleAssistant,
			Content: "我要执行我刚才计划的步骤了。",
			ToolCalls: []schema.ToolCall{
				{ID: "call_123", Name: "bash", Arguments: []byte(`{"command": "ls -la"}`)},
			},
		}, nil
	}

	// 第二轮 Action：直接总结退出
	return &schema.Message{
		Role:    schema.RoleAssistant,
		Content: "根据工具返回的结果，我看到了 main.go，任务圆满完成！",
	}, nil
}
