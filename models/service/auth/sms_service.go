package auth

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"agent/library/log"
	"agent/models/dao"
)

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// SMSService 验证码服务（本地模拟发送）
type SMSService struct {
	dao       *dao.SMSCodeDAO
	ttl       time.Duration
	cooldown  time.Duration
	appEnv    string
	fixedCode string
}

func NewSMSService(dao *dao.SMSCodeDAO) *SMSService {
	return &SMSService{
		dao:       dao,
		ttl:       5 * time.Minute,
		cooldown:  60 * time.Second,
		appEnv:    os.Getenv("APP_ENV"),
		fixedCode: os.Getenv("SMS_DEBUG_FIXED_CODE"),
	}
}

func (s *SMSService) ValidatePhone(phone string) bool {
	return phoneRegexp.MatchString(phone)
}

func (s *SMSService) SendCode(ctx context.Context, phone, scene string) error {
	if !s.ValidatePhone(phone) {
		return ErrInvalidPhone
	}
	if scene != "register" && scene != "reset_password" {
		return ErrInvalidSMScene
	}

	latest, err := s.dao.FindLatestValidCode(ctx, phone, scene)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < s.cooldown {
		return ErrTooFrequent
	}

	code := s.generateCode()
	expiredAt := time.Now().Add(s.ttl)
	if err := s.dao.CreateCode(ctx, phone, scene, code, expiredAt); err != nil {
		return err
	}

	log.Info(ctx, "SMSService.SendCode: 模拟发送验证码 phone=%s scene=%s code=%s", phone, scene, code)
	return nil
}

func (s *SMSService) VerifyCode(ctx context.Context, phone, scene, code string) error {
	record, err := s.dao.FindLatestValidCode(ctx, phone, scene)
	if err != nil {
		return err
	}
	if record == nil {
		return ErrCodeNotFound
	}
	if record.Used == 1 {
		return ErrCodeUsed
	}
	if time.Now().After(record.ExpiredAt) {
		return ErrCodeExpired
	}
	if record.Attempts >= 5 {
		return ErrTooManyAttempts
	}
	if record.Code != code {
		_ = s.dao.IncreaseAttempts(ctx, record.ID)
		return ErrInvalidCode
	}
	return s.dao.MarkUsed(ctx, record.ID)
}

func (s *SMSService) generateCode() string {
	if s.appEnv == "dev" && s.fixedCode != "" {
		return s.fixedCode
	}
	n := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000)
	return fmt.Sprintf("%06d", n)
}
