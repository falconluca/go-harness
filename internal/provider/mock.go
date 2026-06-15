package provider

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

// ==========================================
// 1. 伪造的大模型 Provider
// ==========================================
type MockProvider struct {
	turn int
}

// 模拟大模型的响应：第一轮请求执行 bash，第二轮输出最终结果
func (m *MockProvider) Generate(ctx context.Context, msgs []schema.Message,
	_ []schema.ToolDefinition) (*schema.Message, error) {

	m.turn++
	if m.turn == 1 {
		return &schema.Message{
			Role:    schema.RoleAssistant,
			Content: "让我来看看当前目录下有什么文件。",
			ToolCalls: []schema.ToolCall{
				{ID: "call_123", Name: "bash", Arguments: []byte(`{"command": "ls -la"}`)},
			},
		}, nil
	}

	return &schema.Message{
		Role:    schema.RoleAssistant,
		Content: "我看到了文件列表，里面包含 main.go，任务完成！",
	}, nil
}
