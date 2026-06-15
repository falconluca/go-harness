package tools

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

type MockRegistry struct{}

func (m *MockRegistry) GetAvailableTools() []schema.ToolDefinition {
	// 为了让 Phase 2 能检测到工具，这里返回一个伪造的工具定义数组
	return []schema.ToolDefinition{{Name: "bash"}}
}

func (m *MockRegistry) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     "-rw-r--r--  1 user group  234 Oct 24 10:00 main.go\n",
		IsError:    false,
	}
}
