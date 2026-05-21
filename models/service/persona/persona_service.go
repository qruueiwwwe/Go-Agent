package persona

import (
	"context"
	"errors"

	"agent/library/log"
	"agent/models/dao"
	"agent/models/entity"
	"agent/models/service/blocker"
)

var (
	// ErrPersonaNotFound 角色卡不存在
	ErrPersonaNotFound = errors.New("角色卡不存在")
	// ErrNoPermission 无权限
	ErrNoPermission = errors.New("无权限操作此角色卡")
	// ErrInvalidInput 输入无效
	ErrInvalidInput = errors.New("输入参数无效")
	// ErrBlockedWord 包含敏感词
	ErrBlockedWord = errors.New("内容包含敏感词，请修改后重试")
)

// PersonaService 角色卡服务
type PersonaService struct {
	personaDAO *dao.PersonaDAO
	exampleDAO *dao.PersonaExampleDAO
	blocker    *blocker.Service
}

// NewPersonaService 创建角色卡服务
func NewPersonaService(personaDAO *dao.PersonaDAO, exampleDAO *dao.PersonaExampleDAO, blockerSvc *blocker.Service) *PersonaService {
	return &PersonaService{
		personaDAO: personaDAO,
		exampleDAO: exampleDAO,
		blocker:    blockerSvc,
	}
}

// CreatePersona 创建角色卡（含示例对话）
func (s *PersonaService) CreatePersona(ctx context.Context, creatorID int64, req *entity.PersonaCreateRequest) (*entity.Persona, error) {
	if req.Name == "" || req.SystemPrompt == "" {
		return nil, ErrInvalidInput
	}

	// 屏蔽词检查
	if s.blocker != nil {
		fields := map[string]string{
			"名称":  req.Name,
			"简介":  req.Tagline,
			"性格":  req.Personality,
			"提示词": req.SystemPrompt,
		}
		if found, word, field := s.blocker.CheckMultiple(fields); found {
			log.Warn(ctx, "PersonaService.CreatePersona: 包含敏感词 field=%s, word=%s", field, word)
			return nil, ErrBlockedWord
		}
		// 检查示例对话
		for i, ex := range req.Examples {
			exFields := map[string]string{
				"示例用户输入": ex.UserInput,
				"示例AI回复": ex.AIResponse,
			}
			if found, word, field := s.blocker.CheckMultiple(exFields); found {
				log.Warn(ctx, "PersonaService.CreatePersona: 示例对话包含敏感词 index=%d, field=%s, word=%s", i, field, word)
				return nil, ErrBlockedWord
			}
		}
	}

	// 限制示例数量
	if len(req.Examples) > 5 {
		req.Examples = req.Examples[:5]
	}

	persona := &entity.Persona{
		Name:         req.Name,
		Avatar:       req.Avatar,
		Tagline:      req.Tagline,
		Personality:  req.Personality,
		SystemPrompt: req.SystemPrompt,
		CreatorID:    &creatorID,
		IsPublic:     req.IsPublic,
		UsageCount:   0,
		Status:       1,
	}

	// 创建角色卡
	if err := s.personaDAO.Create(ctx, persona); err != nil {
		return nil, err
	}

	// 创建示例对话
	if len(req.Examples) > 0 {
		if err := s.exampleDAO.BatchCreate(ctx, persona.ID, req.Examples); err != nil {
			// 示例创建失败，记录日志但不回滚角色卡
			log.Error(ctx, "PersonaService.CreatePersona: 创建示例失败 personaID=%d, err=%v", persona.ID, err)
		} else {
			persona.Examples = req.Examples
		}
	}

	log.Info(ctx, "PersonaService.CreatePersona: 创建成功 id=%d, name=%s", persona.ID, persona.Name)
	return persona, nil
}

// GetPersona 获取角色卡详情（含示例对话）
func (s *PersonaService) GetPersona(ctx context.Context, id int64) (*entity.Persona, error) {
	persona, err := s.personaDAO.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if persona == nil {
		return nil, ErrPersonaNotFound
	}

	// 加载示例对话
	examples, err := s.exampleDAO.FindByPersonaID(ctx, id)
	if err != nil {
		log.Error(ctx, "PersonaService.GetPersona: 加载示例失败 id=%d, err=%v", id, err)
	} else {
		persona.Examples = examples
	}

	return persona, nil
}

// UpdatePersona 更新角色卡（权限校验：创建者或管理员）
func (s *PersonaService) UpdatePersona(ctx context.Context, userID int64, userRole string, id int64, req *entity.PersonaCreateRequest) error {
	if req.Name == "" || req.SystemPrompt == "" {
		return ErrInvalidInput
	}

	// 查询角色卡
	persona, err := s.personaDAO.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if persona == nil {
		return ErrPersonaNotFound
	}

	// 权限校验：创建者或管理员
	if !s.canModify(persona, userID, userRole) {
		return ErrNoPermission
	}

	// 限制示例数量
	if len(req.Examples) > 5 {
		req.Examples = req.Examples[:5]
	}

	// 更新角色卡
	persona.Name = req.Name
	persona.Avatar = req.Avatar
	persona.Tagline = req.Tagline
	persona.Personality = req.Personality
	persona.SystemPrompt = req.SystemPrompt
	persona.IsPublic = req.IsPublic

	if err := s.personaDAO.Update(ctx, persona); err != nil {
		return err
	}

	// 删除旧示例，创建新示例
	if err := s.exampleDAO.DeleteByPersonaID(ctx, id); err != nil {
		log.Error(ctx, "PersonaService.UpdatePersona: 删除旧示例失败 id=%d, err=%v", id, err)
	}

	if len(req.Examples) > 0 {
		if err := s.exampleDAO.BatchCreate(ctx, id, req.Examples); err != nil {
			log.Error(ctx, "PersonaService.UpdatePersona: 创建新示例失败 id=%d, err=%v", id, err)
		}
	}

	log.Info(ctx, "PersonaService.UpdatePersona: 更新成功 id=%d", id)
	return nil
}

// DeletePersona 删除角色卡（权限校验：创建者或管理员）
func (s *PersonaService) DeletePersona(ctx context.Context, userID int64, userRole string, id int64) error {
	// 查询角色卡
	persona, err := s.personaDAO.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if persona == nil {
		return ErrPersonaNotFound
	}

	// 权限校验：创建者或管理员
	if !s.canModify(persona, userID, userRole) {
		return ErrNoPermission
	}

	// 软删除角色卡
	if err := s.personaDAO.Delete(ctx, id); err != nil {
		return err
	}

	log.Info(ctx, "PersonaService.DeletePersona: 删除成功 id=%d", id)
	return nil
}

// ListPersonas 列表查询
func (s *PersonaService) ListPersonas(ctx context.Context, queryType string, userID int64, page, size int) (*entity.PersonaListResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	var list []*entity.Persona
	var total int64
	var err error

	if queryType == "mine" {
		// 查询我的角色卡
		list, total, err = s.personaDAO.FindByCreator(ctx, userID, page, size)
	} else {
		// 查询公开角色卡
		list, total, err = s.personaDAO.FindPublic(ctx, page, size)
	}

	if err != nil {
		return nil, err
	}

	return &entity.PersonaListResponse{
		Total: total,
		Page:  page,
		Size:  size,
		List:  s.convertToList(list),
	}, nil
}

// canModify 检查是否有修改权限
func (s *PersonaService) canModify(persona *entity.Persona, userID int64, userRole string) bool {
	// 管理员有所有权限
	if userRole == "admin" {
		return true
	}
	// 创建者有权限
	if persona.CreatorID != nil && *persona.CreatorID == userID {
		return true
	}
	return false
}

// convertToList 转换为列表格式（去除详细信息）
func (s *PersonaService) convertToList(list []*entity.Persona) []entity.Persona {
	result := make([]entity.Persona, len(list))
	for i, p := range list {
		result[i] = entity.Persona{
			ID:         p.ID,
			Name:       p.Name,
			Avatar:     p.Avatar,
			Tagline:    p.Tagline,
			UsageCount: p.UsageCount,
			CreatedAt:  p.CreatedAt,
		}
	}
	return result
}
