package agent

import (
	"context"
	"fmt"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) string
}

// ToolManager 工具管理器
type ToolManager struct {
	tools map[string]Tool
}

func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]Tool),
	}
}

func (m *ToolManager) Register(t Tool) {
	m.tools[t.Name()] = t
}

func (m *ToolManager) Execute(ctx context.Context, toolName, input string) string {
	if t, ok := m.tools[toolName]; ok {
		return t.Execute(ctx, input)
	}
	return "未知工具：" + toolName
}

func (m *ToolManager) GetToolsDesc() string {
	desc := ""
	for _, t := range m.tools {
		desc += fmt.Sprintf("%s: %s\n", t.Name(), t.Description())
	}
	return desc
}

// GetToolList 获取所有工具
func (m *ToolManager) GetToolList() []Tool {
	tools := make([]Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}
