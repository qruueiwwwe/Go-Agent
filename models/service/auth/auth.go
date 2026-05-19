package auth

import (
	"context"
	"errors"
	"time"

	"agent/models/dao"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("用户名已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserDisabled       = errors.New("用户已被禁用")
	ErrUserDeleted        = errors.New("账号已被删除")
	ErrPasswordTooShort   = errors.New("密码长度至少6位")
	ErrInvalidPhone       = errors.New("手机号格式不正确")
	ErrInvalidSMScene     = errors.New("验证码场景不正确")
	ErrTooFrequent        = errors.New("请求过于频繁，请稍后再试")
	ErrCodeNotFound       = errors.New("验证码不存在")
	ErrCodeUsed           = errors.New("验证码已使用")
	ErrCodeExpired        = errors.New("验证码已过期")
	ErrTooManyAttempts    = errors.New("验证码尝试次数过多")
	ErrInvalidCode        = errors.New("验证码错误")
	ErrPhoneAlreadyUsed   = errors.New("手机号已被使用")
	ErrPhoneNotBound      = errors.New("该手机号未绑定账号")
	ErrInvalidEmail       = errors.New("邮箱格式不正确")
	ErrEmailAlreadyUsed   = errors.New("邮箱已被使用")
	ErrEmailNotBound      = errors.New("该邮箱未绑定账号")
)

// AuthService 认证服务
type AuthService struct {
	userDAO      *dao.UserDAO
	smsCodeDAO   *dao.SMSCodeDAO
	smsService   *SMSService
	emailCodeDAO *dao.EmailCodeDAO
	emailService *EmailService
	jwtSecret    []byte
	expireTime   time.Duration
}

// Claims JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthService 创建认证服务
func NewAuthService(
	userDAO *dao.UserDAO,
	smsCodeDAO *dao.SMSCodeDAO,
	smsService *SMSService,
	emailCodeDAO *dao.EmailCodeDAO,
	emailService *EmailService,
	jwtSecret string,
	expireTime time.Duration,
) *AuthService {
	return &AuthService{
		userDAO:      userDAO,
		smsCodeDAO:   smsCodeDAO,
		smsService:   smsService,
		emailCodeDAO: emailCodeDAO,
		emailService: emailService,
		jwtSecret:    []byte(jwtSecret),
		expireTime:   expireTime,
	}
}

// Register 基础注册（用户名+密码）。
// 该方法不处理验证码，仅用于创建账号主体。
func (s *AuthService) Register(ctx context.Context, username, password, nickname string) (*dao.User, error) {
	// 验证密码长度
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}

	// 检查用户名是否存在
	existingUser, err := s.userDAO.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &dao.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: nickname,
		Role:     "user",
		Status:   1,
	}

	if err := s.userDAO.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 账号密码登录，校验通过后返回 JWT。
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *dao.User, error) {
	// 查找用户
	user, err := s.userDAO.FindByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status == -1 {
		return "", nil, ErrUserDeleted
	}
	if user.Status == 0 {
		return "", nil, ErrUserDisabled
	}

	// 生成 JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// GenerateToken 生成 JWT Token
func (s *AuthService) GenerateToken(user *dao.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "agent",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ParseToken 解析 JWT Token
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecret, nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateToken 验证 Token 并返回用户信息
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*dao.User, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 从数据库获取最新用户信息
	user, err := s.userDAO.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// RegisterByPhoneCode 手机号+验证码注册
func (s *AuthService) RegisterByPhoneCode(ctx context.Context, phone, code, username, password, nickname string) (*dao.User, error) {
	if !s.smsService.ValidatePhone(phone) {
		return nil, ErrInvalidPhone
	}
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}
	if err := s.smsService.VerifyCode(ctx, phone, "register", code); err != nil {
		return nil, err
	}

	existingByPhone, err := s.userDAO.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existingByPhone != nil {
		return nil, ErrPhoneAlreadyUsed
	}

	user, err := s.Register(ctx, username, password, nickname)
	if err != nil {
		return nil, err
	}
	if err := s.userDAO.BindPhone(ctx, user.ID, phone); err != nil {
		return nil, err
	}
	user.Phone = phone
	return user, nil
}

// ResetPasswordByPhoneCode 手机号+验证码重置密码
func (s *AuthService) ResetPasswordByPhoneCode(ctx context.Context, phone, code, newPassword string) error {
	if !s.smsService.ValidatePhone(phone) {
		return ErrInvalidPhone
	}
	if len(newPassword) < 6 {
		return ErrPasswordTooShort
	}
	if err := s.smsService.VerifyCode(ctx, phone, "reset_password", code); err != nil {
		return err
	}

	user, err := s.userDAO.FindByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrPhoneNotBound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userDAO.UpdatePasswordByPhone(ctx, phone, string(hashedPassword))
}

// RegisterByEmailCode 邮箱+验证码注册（phone可选）
func (s *AuthService) RegisterByEmailCode(ctx context.Context, email, code, username, password, nickname, phone string) (*dao.User, error) {
	if !s.emailService.ValidateEmail(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}
	if err := s.emailService.VerifyCode(ctx, email, "register", code); err != nil {
		return nil, err
	}

	existingByEmail, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingByEmail != nil {
		return nil, ErrEmailAlreadyUsed
	}
	if phone != "" {
		existingByPhone, err := s.userDAO.FindByPhone(ctx, phone)
		if err != nil {
			return nil, err
		}
		if existingByPhone != nil {
			return nil, ErrPhoneAlreadyUsed
		}
	}

	user, err := s.Register(ctx, username, password, nickname)
	if err != nil {
		return nil, err
	}
	if err := s.userDAO.BindEmail(ctx, user.ID, email); err != nil {
		return nil, err
	}
	if phone != "" {
		if err := s.userDAO.BindPhone(ctx, user.ID, phone); err != nil {
			return nil, err
		}
	}
	user.Email = email
	user.Phone = phone
	return user, nil
}

// ResetPasswordByEmailCode 邮箱+验证码重置密码
func (s *AuthService) ResetPasswordByEmailCode(ctx context.Context, email, code, newPassword string) error {
	if !s.emailService.ValidateEmail(email) {
		return ErrInvalidEmail
	}
	if len(newPassword) < 6 {
		return ErrPasswordTooShort
	}
	if err := s.emailService.VerifyCode(ctx, email, "reset_password", code); err != nil {
		return err
	}

	user, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrEmailNotBound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userDAO.UpdatePasswordByEmail(ctx, email, string(hashedPassword))
}
