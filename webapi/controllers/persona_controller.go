package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"agent/library/log"
	"agent/models/entity"
	"agent/models/service/persona"
)

// PersonaController 角色卡控制器
type PersonaController struct {
	personaSvc *persona.PersonaService
	chatSvc    *persona.ChatService
}

// NewPersonaController 创建角色卡控制器
func NewPersonaController(personaSvc *persona.PersonaService, chatSvc *persona.ChatService) *PersonaController {
	return &PersonaController{
		personaSvc: personaSvc,
		chatSvc:    chatSvc,
	}
}

// PersonaList 获取角色卡列表
func (c *PersonaController) PersonaList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodGet {
		ReplyError(w, 404, "只支持 GET 方法", logid)
		return
	}

	queryType := r.URL.Query().Get("type")
	if queryType == "" {
		queryType = "public"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 50 {
		size = 10
	}

	resp, err := c.personaSvc.ListPersonas(ctx, queryType, claims.UserID, page, size)
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaList: 查询失败 err=%v", err)
		ReplyError(w, 500, "查询失败", logid)
		return
	}

	ReplySuccess(w, resp, logid)
}

// PersonaDetail 获取角色卡详情
func (c *PersonaController) PersonaDetail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)

	if r.Method != http.MethodGet {
		ReplyError(w, 404, "只支持 GET 方法", logid)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		ReplyError(w, 400, "缺少 id 参数", logid)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ReplyError(w, 400, "id 参数无效", logid)
		return
	}

	p, err := c.personaSvc.GetPersona(ctx, id)
	if err == persona.ErrPersonaNotFound {
		ReplyError(w, 404, "角色卡不存在", logid)
		return
	}
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaDetail: 查询失败 id=%d, err=%v", id, err)
		ReplyError(w, 500, "查询失败", logid)
		return
	}

	ReplySuccess(w, p, logid)
}

// PersonaCreate 创建角色卡
func (c *PersonaController) PersonaCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodPost {
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	var req entity.PersonaCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ReplyError(w, 400, "解析请求失败", logid)
		return
	}

	if req.Name == "" || req.SystemPrompt == "" {
		ReplyError(w, 400, "名称和系统提示词不能为空", logid)
		return
	}

	p, err := c.personaSvc.CreatePersona(ctx, claims.UserID, &req)
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaCreate: 创建失败 err=%v", err)
		ReplyError(w, 500, "创建失败", logid)
		return
	}

	ReplySuccess(w, map[string]interface{}{
		"id":   p.ID,
		"name": p.Name,
	}, logid)
}

// PersonaUpdate 更新角色卡
func (c *PersonaController) PersonaUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodPost {
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	var req struct {
		ID           int64               `json:"id"`
		Name         string              `json:"name"`
		Avatar       string              `json:"avatar,omitempty"`
		Tagline      string              `json:"tagline,omitempty"`
		Personality  string              `json:"personality,omitempty"`
		SystemPrompt string              `json:"system_prompt"`
		IsPublic     bool                `json:"is_public"`
		Examples     []entity.Example    `json:"examples,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ReplyError(w, 400, "解析请求失败", logid)
		return
	}

	if req.ID == 0 {
		ReplyError(w, 400, "缺少 id 参数", logid)
		return
	}

	createReq := &entity.PersonaCreateRequest{
		Name:         req.Name,
		Avatar:       req.Avatar,
		Tagline:      req.Tagline,
		Personality:  req.Personality,
		SystemPrompt: req.SystemPrompt,
		IsPublic:     req.IsPublic,
		Examples:     req.Examples,
	}

	err := c.personaSvc.UpdatePersona(ctx, claims.UserID, claims.Role, req.ID, createReq)
	if err == persona.ErrPersonaNotFound {
		ReplyError(w, 404, "角色卡不存在", logid)
		return
	}
	if err == persona.ErrNoPermission {
		ReplyError(w, 403, "无权限操作此角色卡", logid)
		return
	}
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaUpdate: 更新失败 id=%d, err=%v", req.ID, err)
		ReplyError(w, 500, "更新失败", logid)
		return
	}

	ReplySuccess(w, map[string]interface{}{"id": req.ID}, logid)
}

// PersonaDelete 删除角色卡
func (c *PersonaController) PersonaDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodPost {
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ReplyError(w, 400, "解析请求失败", logid)
		return
	}

	if req.ID == 0 {
		ReplyError(w, 400, "缺少 id 参数", logid)
		return
	}

	err := c.personaSvc.DeletePersona(ctx, claims.UserID, claims.Role, req.ID)
	if err == persona.ErrPersonaNotFound {
		ReplyError(w, 404, "角色卡不存在", logid)
		return
	}
	if err == persona.ErrNoPermission {
		ReplyError(w, 403, "无权限操作此角色卡", logid)
		return
	}
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaDelete: 删除失败 id=%d, err=%v", req.ID, err)
		ReplyError(w, 500, "删除失败", logid)
		return
	}

	ReplySuccess(w, nil, logid)
}

// PersonaChat 角色卡对话
func (c *PersonaController) PersonaChat(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodPost {
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	var req entity.PersonaChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ReplyError(w, 400, "解析请求失败", logid)
		return
	}

	if req.PersonaID == 0 {
		ReplyError(w, 400, "缺少 persona_id 参数", logid)
		return
	}

	if req.Message == "" {
		ReplyError(w, 400, "消息不能为空", logid)
		return
	}

	resp, err := c.chatSvc.Chat(ctx, claims.UserID, &req)
	if err == persona.ErrPersonaNotFound {
		ReplyError(w, 404, "角色卡不存在", logid)
		return
	}
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaChat: 对话失败 personaID=%d, err=%v", req.PersonaID, err)
		ReplyError(w, 500, "对话失败", logid)
		return
	}

	ReplySuccess(w, resp, logid)
}

// PersonaSessions 获取用户的会话列表
func (c *PersonaController) PersonaSessions(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodGet {
		ReplyError(w, 404, "只支持 GET 方法", logid)
		return
	}

	personaID, _ := strconv.ParseInt(r.URL.Query().Get("persona_id"), 10, 64)

	sessions, err := c.chatSvc.GetSessions(ctx, claims.UserID, personaID)
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaSessions: 查询失败 err=%v", err)
		ReplyError(w, 500, "查询失败", logid)
		return
	}

	ReplySuccess(w, map[string]interface{}{
		"sessions": sessions,
	}, logid)
}

// PersonaHistory 获取会话历史
func (c *PersonaController) PersonaHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	claims := GetUserFromContext(ctx)

	if r.Method != http.MethodGet {
		ReplyError(w, 404, "只支持 GET 方法", logid)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		ReplyError(w, 400, "缺少 session_id 参数", logid)
		return
	}

	history, err := c.chatSvc.GetHistory(ctx, claims.UserID, sessionID)
	if err == persona.ErrSessionNotFound {
		ReplyError(w, 404, "会话不存在", logid)
		return
	}
	if err != nil {
		log.Error(ctx, "PersonaController.PersonaHistory: 查询失败 sessionID=%s, err=%v", sessionID, err)
		ReplyError(w, 500, "查询失败", logid)
		return
	}

	ReplySuccess(w, history, logid)
}
