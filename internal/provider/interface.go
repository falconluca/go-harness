package provider

import (
	"context"

	"github.com/falconluca/go-harness/internal/schema"
)

type LLMProvider interface {
	Generate(c context.Context, messages []schema.Message,
		availableTools []schema.ToolDefinition) (*schema.Message, error)
}
