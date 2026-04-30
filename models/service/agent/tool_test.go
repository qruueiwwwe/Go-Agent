package agent

import (
	"context"
	"testing"
)

// MockTool 用于测试的模拟工具
type MockTool struct {
	name        string
	description string
	result      string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Execute(ctx context.Context, input string) string {
	return m.result
}

func TestToolManager_NewToolManager(t *testing.T) {
	tm := NewToolManager()
	if tm == nil {
		t.Error("NewToolManager() should not return nil")
	}
	if tm.tools == nil {
		t.Error("ToolManager.tools should be initialized")
	}
}

func TestToolManager_Register(t *testing.T) {
	tm := NewToolManager()
	tool := &MockTool{name: "test", description: "test tool", result: "ok"}

	tm.Register(tool)

	if len(tm.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tm.tools))
	}
	if tm.tools["test"] == nil {
		t.Error("Tool should be registered with key 'test'")
	}
}

func TestToolManager_Execute_ExistingTool(t *testing.T) {
	tm := NewToolManager()
	tool := &MockTool{name: "test", description: "test tool", result: "executed"}
	tm.Register(tool)

	ctx := context.Background()
	result := tm.Execute(ctx, "test", "input")

	if result != "executed" {
		t.Errorf("Expected 'executed', got '%s'", result)
	}
}

func TestToolManager_Execute_UnknownTool(t *testing.T) {
	tm := NewToolManager()

	ctx := context.Background()
	result := tm.Execute(ctx, "unknown", "input")

	expected := "未知工具：unknown"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestToolManager_GetToolsDesc(t *testing.T) {
	tm := NewToolManager()
	tm.Register(&MockTool{name: "tool1", description: "First tool", result: ""})
	tm.Register(&MockTool{name: "tool2", description: "Second tool", result: ""})

	desc := tm.GetToolsDesc()

	if desc == "" {
		t.Error("GetToolsDesc() should not return empty string")
	}
	if !contains(desc, "tool1") || !contains(desc, "tool2") {
		t.Errorf("GetToolsDesc() should contain tool names, got: %s", desc)
	}
}

func TestToolManager_GetToolList(t *testing.T) {
	tm := NewToolManager()
	tm.Register(&MockTool{name: "tool1", description: "First tool", result: ""})
	tm.Register(&MockTool{name: "tool2", description: "Second tool", result: ""})

	tools := tm.GetToolList()

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestToolManager_GetToolList_Empty(t *testing.T) {
	tm := NewToolManager()

	tools := tm.GetToolList()

	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}