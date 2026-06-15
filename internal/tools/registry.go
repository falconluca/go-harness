package tools

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

type Registry interface {
	GetAvailableTools() []schema.ToolDefinition
	Execute(c context.Context, call schema.ToolCall) schema.ToolResult
}
