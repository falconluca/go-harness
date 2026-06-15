package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/falconluca/go-harness/internal/schema"
)

type BaseTool interface {
	Name() string
	Definition() schema.ToolDefinition
	Execute(c context.Context, args json.RawMessage) (string, error)
}

type Registry interface {
	Register(tool BaseTool)
	GetAvailableTools() []schema.ToolDefinition
	Execute(c context.Context, call schema.ToolCall) schema.ToolResult
}

type registryImpl struct {
	tools map[string]BaseTool
}

func NewRegistry() Registry {
	return &registryImpl{
		tools: make(map[string]BaseTool),
	}
}

func (r *registryImpl) Register(tool BaseTool) {
	toolName := tool.Name()
	if _, exists := r.tools[toolName]; exists {
		log.Printf("[Warning] 工具 '%s' 已经被注册，将被覆盖。\n", toolName)
	}
	r.tools[toolName] = tool
	log.Printf("[Registry] 成功挂载工具: %s\n", toolName)
}

func (r *registryImpl) GetAvailableTools() []schema.ToolDefinition {
	definitions := make([]schema.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, tool.Definition())
	}
	return definitions
}

func (r *registryImpl) Execute(c context.Context, call schema.ToolCall) schema.ToolResult {
	tool, exists := r.tools[call.Name]
	if !exists {
		errMsg := fmt.Sprintf("Error: 系统中不存在名为 '%s' 的工具。", call.Name)
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     errMsg,
			IsError:    true,
		}
	}

	output, err := tool.Execute(c, call.Arguments)
	if err != nil {
		errMsg := fmt.Sprintf("Error executing %s: %v", call.Name, err)
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     errMsg,
			IsError:    true,
		}
	}

	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     output,
		IsError:    false,
	}
}
