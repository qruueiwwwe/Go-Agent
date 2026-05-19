package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"time"

	"agent/library/log"
	"agent/models/dao"
)

// emailRegexp 用于基础邮箱格式校验（不做 MX 解析，仅做语法层过滤）。
var emailRegexp = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

// EmailService 提供邮箱验证码发送与校验能力。
// - 发送：生成验证码、持久化、调用 SMTP 发信
// - 校验：检查存在性、过期、尝试次数、使用状态
// - 开发模式：支持固定验证码（仅 APP_ENV=dev）
type EmailService struct {
	dao       *dao.EmailCodeDAO
	ttl       time.Duration
	cooldown  time.Duration
	appEnv    string
	fixedCode string

	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
	fromName string
	fromAddr string
}

// NewEmailService 从环境变量加载 SMTP 与调试配置，构造邮箱验证码服务。
func NewEmailService(dao *dao.EmailCodeDAO) *EmailService {
	port, _ := strconv.Atoi(os.Getenv("MAIL_SMTP_PORT"))
	if port == 0 {
		port = 465
	}
	fromAddr := os.Getenv("MAIL_FROM_ADDR")
	if fromAddr == "" {
		fromAddr = os.Getenv("MAIL_USERNAME")
	}
	return &EmailService{
		dao:       dao,
		ttl:       5 * time.Minute,
		cooldown:  60 * time.Second,
		appEnv:    os.Getenv("APP_ENV"),
		fixedCode: os.Getenv("EMAIL_DEBUG_FIXED_CODE"),
		smtpHost:  os.Getenv("MAIL_SMTP_HOST"),
		smtpPort:  port,
		smtpUser:  os.Getenv("MAIL_USERNAME"),
		smtpPass:  os.Getenv("MAIL_PASSWORD"),
		fromName:  os.Getenv("MAIL_FROM_NAME"),
		fromAddr:  fromAddr,
	}
}

// ValidateEmail 校验邮箱格式是否合法。
func (s *EmailService) ValidateEmail(email string) bool {
	return emailRegexp.MatchString(email)
}

// SendCode 发送验证码：校验邮箱与场景、频控、入库并发信。
func (s *EmailService) SendCode(ctx context.Context, email, scene string) error {
	if !s.ValidateEmail(email) {
		return ErrInvalidEmail
	}
	if scene != "register" && scene != "reset_password" {
		return ErrInvalidSMScene
	}

	latest, err := s.dao.FindLatestValidCode(ctx, email, scene)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < s.cooldown {
		return ErrTooFrequent
	}

	code := s.generateCode()
	expiredAt := time.Now().Add(s.ttl)
	if err := s.dao.CreateCode(ctx, email, scene, code, expiredAt); err != nil {
		return err
	}

	if err := s.sendEmail(email, code, scene); err != nil {
		log.Error(ctx, "EmailService.SendCode: 发信失败 email=%s err=%v", email, err)
		if s.appEnv == "dev" {
			log.Info(ctx, "EmailService.SendCode: dev回退日志验证码 email=%s scene=%s code=%s", email, scene, code)
			return nil
		}
		return err
	}

	log.Info(ctx, "EmailService.SendCode: 验证码发送成功 email=%s scene=%s", email, scene)
	return nil
}

// VerifyCode 校验验证码有效性，成功后标记为已使用。
func (s *EmailService) VerifyCode(ctx context.Context, email, scene, code string) error {
	record, err := s.dao.FindLatestValidCode(ctx, email, scene)
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

// generateCode 生成验证码。
// 规则：
// - APP_ENV=dev 且 EMAIL_DEBUG_FIXED_CODE 非空时，返回固定值（可含英文）
// - 其他情况返回 4 位纯数字验证码
func (s *EmailService) generateCode() string {
	if s.appEnv == "dev" && s.fixedCode != "" {
		return s.fixedCode
	}
	n := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10000)
	return fmt.Sprintf("%04d", n)
}

// sendEmail 通过 SMTP 发送验证码邮件。
// 端口 465 走 TLS 直连，其他端口使用 SendMail。
func (s *EmailService) sendEmail(toEmail, code, scene string) error {
	if s.smtpHost == "" || s.smtpUser == "" || s.smtpPass == "" || s.fromAddr == "" {
		return fmt.Errorf("SMTP配置不完整")
	}
	subject := "验证码"
	if scene == "reset_password" {
		subject = "重置密码验证码"
	}
	body := fmt.Sprintf("您的验证码是：%s，有效期5分钟。", code)
	fromName := s.fromName
	if fromName == "" {
		fromName = "Agent"
	}
	msg := []byte("From: " + fromName + " <" + s.fromAddr + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body)

	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	if s.smtpPort == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.smtpHost})
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, s.smtpHost)
		if err != nil {
			return err
		}
		defer c.Quit()
		if err = c.Auth(auth); err != nil {
			return err
		}
		if err = c.Mail(s.fromAddr); err != nil {
			return err
		}
		if err = c.Rcpt(toEmail); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err = w.Write(msg); err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, s.fromAddr, []string{toEmail}, msg)
}
