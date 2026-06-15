package tools

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

type MockRegistry struct{}

func (m *MockRegistry) GetAvailableTools() []schema.ToolDefinition { return nil }

func (m *MockRegistry) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	// 直接返回一段伪造的终端输出
	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     "-rw-r--r--  1 user group  234 Oct 24 10:00 main.go\n",
		IsError:    false,
	}
}
