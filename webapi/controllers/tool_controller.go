package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"agent/models/service/agent"
)

// ToolController 工具控制器
type ToolController struct {
	toolManager *agent.ToolManager
}

// NewToolController 创建工具控制器
func NewToolController(tm *agent.ToolManager) *ToolController {
	return &ToolController{
		toolManager: tm,
	}
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToolListResponse 工具列表响应
type ToolListResponse struct {
	Tools []ToolInfo `json:"tools"`
}

// ToolExecuteRequest 工具执行请求
type ToolExecuteRequest struct {
	Input string `json:"input"`
}

// ToolExecuteResponse 工具执行响应
type ToolExecuteResponse struct {
	Tool   string `json:"tool"`
	Input  string `json:"input"`
	Result string `json:"result"`
}

// ListTools 获取工具列表
func (c *ToolController) ListTools(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := c.toolManager.GetToolList()
	result := ToolListResponse{Tools: make([]ToolInfo, 0, len(tools))}

	for _, t := range tools {
		result.Tools = append(result.Tools, ToolInfo{
			Name:        t.Name(),
			Description: t.Description(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ExecuteTool 执行工具（统一入口）
func (c *ToolController) ExecuteTool(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 获取工具名
	toolName := r.PathValue("tool")
	if toolName == "" {
		// 兼容旧版本 Go
		path := r.URL.Path
		parts := strings.Split(path, "/")
		// /internal/tools/execute/nbnhhsh -> ["", "internal", "tools", "execute", "nbnhhsh"]
		if len(parts) >= 5 {
			toolName = parts[4]
		}
	}

	if toolName == "" {
		http.Error(w, "Tool name required", http.StatusBadRequest)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := c.toolManager.Execute(ctx, toolName, req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolExecuteResponse{
		Tool:   toolName,
		Input:  req.Input,
		Result: result,
	})
}

// Weather 天气工具接口
func (c *ToolController) Weather(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := c.toolManager.Execute(ctx, "weather", req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolExecuteResponse{
		Tool:   "weather",
		Input:  req.Input,
		Result: result,
	})
}

// Calculator 计算器工具接口
func (c *ToolController) Calculator(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := c.toolManager.Execute(ctx, "calculator", req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolExecuteResponse{
		Tool:   "calculator",
		Input:  req.Input,
		Result: result,
	})
}

// File 文件工具接口
func (c *ToolController) File(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := c.toolManager.Execute(ctx, "file", req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolExecuteResponse{
		Tool:   "file",
		Input:  req.Input,
		Result: result,
	})
}

// Nbnhhsh 缩写词猜测工具接口
func (c *ToolController) Nbnhhsh(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := c.toolManager.Execute(ctx, "nbnhhsh", req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolExecuteResponse{
		Tool:   "nbnhhsh",
		Input:  req.Input,
		Result: result,
	})
}
