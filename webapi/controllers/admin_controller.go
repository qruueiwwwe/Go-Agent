package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"agent/library/log"
	"agent/models/dao"
)

// AdminController 后台管理控制器
type AdminController struct {
	userDAO *dao.UserDAO
}

// NewAdminController 创建后台管理控制器
func NewAdminController(userDAO *dao.UserDAO) *AdminController {
	return &AdminController{userDAO: userDAO}
}

// GetUsers 获取所有用户列表
func (c *AdminController) GetUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GenerateLogIDWithUser(claims)

	users, err := c.userDAO.ListAll(ctx)
	if err != nil {
		Reply(w, 500, "获取用户列表失败", nil, logid)
		return
	}

	// 清除密码字段
	for _, u := range users {
		u.Password = ""
	}

	Reply(w, 0, "", users, logid)
}

// UpdateUserRoleRequest 更新角色请求
type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserRole 更新用户角色
func (c *AdminController) UpdateUserRole(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GenerateLogIDWithUser(claims)

	if claims == nil {
		Reply(w, 401, "未登录", nil, logid)
		return
	}

	// 解析路径参数 /api/admin/users/{id}/role
	pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		Reply(w, 404, "无效的请求路径", nil, logid)
		return
	}
	userIDStr := pathParts[4]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		Reply(w, 404, "无效的用户ID", nil, logid)
		return
	}

	// 解析请求体
	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}

	// 验证角色值
	if req.Role != "user" && req.Role != "vip" && req.Role != "admin" {
		Reply(w, 404, "无效的角色", nil, logid)
		return
	}

	// 管理员不能取消自己的管理员权限
	if userID == claims.UserID && req.Role != "admin" {
		Reply(w, 403, "不能取消自己的管理员权限", nil, logid)
		return
	}

	// 更新角色
	if err := c.userDAO.UpdateRole(ctx, userID, req.Role); err != nil {
		Reply(w, 500, "更新角色失败", nil, logid)
		return
	}

	Reply(w, 0, "", nil, logid)
}

// UpdateUserStatusRequest 更新状态请求
type UpdateUserStatusRequest struct {
	Status int `json:"status"`
}

// UpdateUserStatus 更新用户状态
func (c *AdminController) UpdateUserStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GenerateLogIDWithUser(claims)

	if claims == nil {
		Reply(w, 401, "未登录", nil, logid)
		return
	}

	// 解析路径参数 /api/admin/users/{id}/status
	pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		Reply(w, 404, "无效的请求路径", nil, logid)
		return
	}
	userIDStr := pathParts[4]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		Reply(w, 404, "无效的用户ID", nil, logid)
		return
	}

	// 解析请求体
	var req UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}

	// 验证状态值
	if req.Status != -1 && req.Status != 1 {
		Reply(w, 404, "无效的状态", nil, logid)
		return
	}

	// 管理员不能删除自己
	if userID == claims.UserID && req.Status == -1 {
		Reply(w, 403, "不能删除自己的账号", nil, logid)
		return
	}

	// 更新状态
	if err := c.userDAO.UpdateStatus(ctx, userID, req.Status); err != nil {
		Reply(w, 500, "更新状态失败", nil, logid)
		return
	}

	Reply(w, 0, "", nil, logid)
}
