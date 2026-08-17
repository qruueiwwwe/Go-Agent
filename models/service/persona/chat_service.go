package persona

import (
	"context"
	"errors"
	"strings"

	"agent/library/log"
	"agent/models/dao"
	"agent/models/entity"

	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
)

// ChatService 角色卡对话服务
type ChatService struct {
	personaDAO     *dao.PersonaDAO
	chatDAO        *dao.PersonaChatDAO
	exampleDAO     *dao.PersonaExampleDAO
	contextBuilder *ContextBuilder
	ollamaSvc      LLMService
	zhipuChat      func(ctx context.Context, msgs []api.Message) (string, error)
	zhipuStream    func(ctx context.Context, msgs []api.Message, tokenCh chan<- string) error
}

// NewChatService 创建角色卡对话服务
func NewChatService(
	personaDAO *dao.PersonaDAO,
	chatDAO *dao.PersonaChatDAO,
	exampleDAO *dao.PersonaExampleDAO,
	ollamaSvc LLMService,
	zhipuChat func(ctx context.Context, msgs []api.Message) (string, error),
	zhipuStream func(ctx context.Context, msgs []api.Message, tokenCh chan<- string) error,
) *ChatService {
	contextBuilder := NewContextBuilder(ollamaSvc)
	return &ChatService{
		personaDAO:     personaDAO,
		chatDAO:        chatDAO,
		exampleDAO:     exampleDAO,
		contextBuilder: contextBuilder,
		ollamaSvc:      ollamaSvc,
		zhipuChat:      zhipuChat,
		zhipuStream:    zhipuStream,
	}
}

// Chat 角色卡对话
func (s *ChatService) Chat(ctx context.Context, userID int64, req *entity.PersonaChatRequest) (*entity.PersonaChatResponse, error) {
	// 1. 获取角色卡
	persona, err := s.personaDAO.FindByID(ctx, req.PersonaID)
	if err != nil {
		return nil, err
	}
	if persona == nil {
		return nil, ErrPersonaNotFound
	}

	// 加载示例对话
	examples, err := s.exampleDAO.FindByPersonaID(ctx, req.PersonaID)
	if err != nil {
		log.Error(ctx, "ChatService.Chat: 加载示例失败 personaID=%d, err=%v", req.PersonaID, err)
	} else {
		persona.Examples = examples
	}

	// 2. 获取或创建会话
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
		log.Info(ctx, "ChatService.Chat: 创建新会话 sessionID=%s, personaID=%d, personaName=%s", sessionID, req.PersonaID, persona.Name)
	} else {
		log.Info(ctx, "ChatService.Chat: 继续会话 sessionID=%s, personaID=%d, personaName=%s", sessionID, req.PersonaID, persona.Name)
	}

	// 3. 加载对话历史
	history, err := s.chatDAO.FindBySession(ctx, userID, sessionID)
	if err != nil {
		log.Error(ctx, "ChatService.Chat: 加载历史失败 sessionID=%s, err=%v", sessionID, err)
		history = nil
	}
	log.Info(ctx, "ChatService.Chat: 加载历史消息 count=%d", len(history))

	// 4. 构建上下文
	msgs := s.contextBuilder.Build(persona, history, req.Message)
	log.Info(ctx, "ChatService.Chat: 构建上下文完成 messageCount=%d", len(msgs))

	// 记录用户消息
	log.Info(ctx, "ChatService.Chat: 用户消息: %s", req.Message)

	// 5. 调用 LLM
	var response string
	var usedBackend string
	if s.ollamaSvc != nil {
		response, err = s.ollamaSvc.Chat(ctx, msgs)
		if err != nil {
			log.Warn(ctx, "ChatService.Chat: Ollama调用失败 err=%v", err)
		} else if response != "" {
			usedBackend = "Ollama"
		}
	}

	// 如果 Ollama 失败或不可用，尝试智谱后备
	if response == "" && s.zhipuChat != nil {
		response, err = s.zhipuChat(ctx, msgs)
		if err != nil {
			log.Error(ctx, "ChatService.Chat: 智谱调用失败 err=%v", err)
			return nil, err
		}
		usedBackend = "智谱"
	}

	if response == "" {
		return nil, ErrLLMUnavailable
	}

	// 记录 AI 响应（截取前200字符）
	responsePreview := response
	if len(responsePreview) > 200 {
		responsePreview = responsePreview[:200] + "..."
	}
	log.Info(ctx, "ChatService.Chat: AI响应[%s]: %s", usedBackend, responsePreview)

	// 6. 保存对话记录
	userMsg := &entity.PersonaChat{
		UserID:    userID,
		PersonaID: req.PersonaID,
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
	}
	if err := s.chatDAO.Create(ctx, userMsg); err != nil {
		log.Error(ctx, "ChatService.Chat: 保存用户消息失败 err=%v", err)
	}

	assistantMsg := &entity.PersonaChat{
		UserID:    userID,
		PersonaID: req.PersonaID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   response,
	}
	if err := s.chatDAO.Create(ctx, assistantMsg); err != nil {
		log.Error(ctx, "ChatService.Chat: 保存AI消息失败 err=%v", err)
	}

	// 7. 更新使用次数
	if err := s.personaDAO.IncrementUsage(ctx, req.PersonaID); err != nil {
		log.Error(ctx, "ChatService.Chat: 更新使用次数失败 err=%v", err)
	}

	log.Info(ctx, "ChatService.Chat: 对话成功 sessionID=%s, personaID=%d", sessionID, req.PersonaID)

	return &entity.PersonaChatResponse{
		SessionID:   sessionID,
		Response:    response,
		PersonaID:   persona.ID,
		PersonaName: persona.Name,
	}, nil
}

// PersonaStreamChunk 流式返回单元
type PersonaStreamChunk struct {
	Type    string // "answer" | "error"
	Content string
}

// ChatStream 角色卡流式对话
// chunkCh 由调用方创建，本方法负责在结束前 close。
// 返回 sessionID / personaName / 完整拼接文本 / err。
func (s *ChatService) ChatStream(
	ctx context.Context,
	userID int64,
	req *entity.PersonaChatRequest,
	chunkCh chan<- PersonaStreamChunk,
) (sessionID string, personaName string, fullResponse string, err error) {
	defer func() {
		// 由本方法关闭 chunkCh，简化调用方
		close(chunkCh)
	}()

	// 1. 获取角色卡
	persona, err := s.personaDAO.FindByID(ctx, req.PersonaID)
	if err != nil {
		return "", "", "", err
	}
	if persona == nil {
		return "", "", "", ErrPersonaNotFound
	}
	personaName = persona.Name

	// 加载示例对话
	examples, exErr := s.exampleDAO.FindByPersonaID(ctx, req.PersonaID)
	if exErr != nil {
		log.Error(ctx, "ChatService.ChatStream: 加载示例失败 personaID=%d, err=%v", req.PersonaID, exErr)
	} else {
		persona.Examples = examples
	}

	// 2. 会话
	sessionID = req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
		log.Info(ctx, "ChatService.ChatStream: 创建新会话 sessionID=%s, personaID=%d", sessionID, req.PersonaID)
	} else {
		log.Info(ctx, "ChatService.ChatStream: 继续会话 sessionID=%s, personaID=%d", sessionID, req.PersonaID)
	}

	// 3. 历史
	history, hErr := s.chatDAO.FindBySession(ctx, userID, sessionID)
	if hErr != nil {
		log.Error(ctx, "ChatService.ChatStream: 加载历史失败 sessionID=%s, err=%v", sessionID, hErr)
		history = nil
	}
	log.Info(ctx, "ChatService.ChatStream: 加载历史消息 count=%d", len(history))

	// 4. 构建上下文
	msgs := s.contextBuilder.Build(persona, history, req.Message)
	log.Info(ctx, "ChatService.ChatStream: 构建上下文完成 messageCount=%d", len(msgs))

	// 5. 流式调用 LLM
	tokenCh := make(chan string, 32)
	doneCh := make(chan error, 1)
	var used string
	go func() {
		var e error
		if s.ollamaSvc != nil {
			e = s.ollamaSvc.ChatStream(ctx, msgs, tokenCh)
			if e == nil {
				used = "Ollama"
			} else {
				log.Warn(ctx, "ChatService.ChatStream: Ollama失败 err=%v", e)
			}
		}
		if used == "" && s.zhipuStream != nil {
			e = s.zhipuStream(ctx, msgs, tokenCh)
			if e == nil {
				used = "智谱"
			} else {
				log.Error(ctx, "ChatService.ChatStream: 智谱流式失败 err=%v", e)
			}
		}
		close(tokenCh)
		doneCh <- e
	}()

	var sb strings.Builder
	for tk := range tokenCh {
		sb.WriteString(tk)
		select {
		case <-ctx.Done():
			// 消费者不再关注：耗尽后端 goroutine 后返回
			for range tokenCh {
			}
			<-doneCh
			return sessionID, personaName, sb.String(), ctx.Err()
		default:
			chunkCh <- PersonaStreamChunk{Type: "answer", Content: tk}
		}
	}
	backendErr := <-doneCh

	fullResponse = sb.String()
	if fullResponse == "" {
		if backendErr != nil {
			return sessionID, personaName, "", backendErr
		}
		return sessionID, personaName, "", ErrLLMUnavailable
	}
	log.Info(ctx, "ChatService.ChatStream: AI响应[%s] len=%d", used, len(fullResponse))

	// 6. 落库
	if e := s.chatDAO.Create(ctx, &entity.PersonaChat{
		UserID: userID, PersonaID: req.PersonaID, SessionID: sessionID,
		Role: "user", Content: req.Message,
	}); e != nil {
		log.Error(ctx, "ChatService.ChatStream: 保存用户消息失败 err=%v", e)
	}
	if e := s.chatDAO.Create(ctx, &entity.PersonaChat{
		UserID: userID, PersonaID: req.PersonaID, SessionID: sessionID,
		Role: "assistant", Content: fullResponse,
	}); e != nil {
		log.Error(ctx, "ChatService.ChatStream: 保存AI消息失败 err=%v", e)
	}

	// 7. 使用次数
	if e := s.personaDAO.IncrementUsage(ctx, req.PersonaID); e != nil {
		log.Error(ctx, "ChatService.ChatStream: 更新使用次数失败 err=%v", e)
	}

	return sessionID, personaName, fullResponse, nil
}

// GetSessions 获取用户的会话列表
func (s *ChatService) GetSessions(ctx context.Context, userID int64, personaID int64) ([]entity.SessionInfo, error) {
	sessions, err := s.chatDAO.FindSessions(ctx, userID, personaID)
	if err != nil {
		return nil, err
	}
	if sessions == nil {
		return []entity.SessionInfo{}, nil
	}
	return sessions, nil
}

// GetHistory 获取会话历史
func (s *ChatService) GetHistory(ctx context.Context, userID int64, sessionID string) (*entity.SessionHistory, error) {
	messages, err := s.chatDAO.FindBySession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, ErrSessionNotFound
	}

	// 获取角色卡名称
	personaID := messages[0].PersonaID
	persona, err := s.personaDAO.FindByID(ctx, personaID)
	personaName := ""
	if err == nil && persona != nil {
		personaName = persona.Name
	}

	return &entity.SessionHistory{
		SessionID:   sessionID,
		PersonaID:   personaID,
		PersonaName: personaName,
		Messages:    messages,
	}, nil
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(ctx context.Context, userID int64, sessionID string) error {
	return s.chatDAO.DeleteSession(ctx, userID, sessionID)
}

// ErrLLMUnavailable LLM 服务不可用
var ErrLLMUnavailable = errors.New("大模型服务暂时不可用")

// ErrSessionNotFound 会话不存在
var ErrSessionNotFound = errors.New("会话不存在")
