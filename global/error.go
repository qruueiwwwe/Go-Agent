package global

// ErrorCode 错误码定义
type ErrorCode int

const (
	// 通用错误 1000-1999
	ErrCodeSuccess       ErrorCode = 1000 // 成功
	ErrCodeUnknown       ErrorCode = 1001 // 未知错误
	ErrCodeInvalidParams ErrorCode = 1002 // 参数错误

	// 服务错误 2000-2999
	ErrCodeServiceInit  ErrorCode = 2001 // 服务初始化失败
	ErrCodeServiceStart ErrorCode = 2002 // 服务启动失败

	// Ollama 相关错误 3000-3999
	ErrCodeOllamaConnect    ErrorCode = 3001 // Ollama 连接失败
	ErrCodeOllamaTimeout    ErrorCode = 3002 // Ollama 请求超时
	ErrCodeOllamaResponse   ErrorCode = 3003 // Ollama 响应错误
	ErrCodeOllamaModelError ErrorCode = 3004 // 模型错误

	// 工具相关错误 4000-4999
	ErrCodeToolNotFound ErrorCode = 4001 // 工具未找到
	ErrCodeToolExecute  ErrorCode = 4002 // 工具执行失败
	ErrCodeToolParams   ErrorCode = 4003 // 工具参数错误

	// 天气API相关错误 5000-5999
	ErrCodeWeatherAPI  ErrorCode = 5001 // 天气API请求失败
	ErrCodeWeatherCity ErrorCode = 5002 // 城市不支持
	ErrCodeWeatherData ErrorCode = 5003 // 天气数据解析失败
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrTypeSystem     ErrorType = "system"      // 系统错误
	ErrTypeBusiness   ErrorType = "business"    // 业务错误
	ErrTypeThirdParty ErrorType = "third_party" // 第三方服务错误
	ErrTypeValidate   ErrorType = "validate"    // 验证错误
)

// Error 错误结构
type Error struct {
	Code    ErrorCode `json:"code"`
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// NewError 创建新错误
func NewError(code ErrorCode, errType ErrorType, message string) *Error {
	return &Error{
		Code:    code,
		Type:    errType,
		Message: message,
	}
}

// NewErrorWithDetails 创建带详情的错误
func NewErrorWithDetails(code ErrorCode, errType ErrorType, message, details string) *Error {
	return &Error{
		Code:    code,
		Type:    errType,
		Message: message,
		Details: details,
	}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Details != "" {
		return string(e.Type) + " error: " + e.Message + " (" + e.Details + ")"
	}
	return string(e.Type) + " error: " + e.Message
}

// GetErrorMsg 获取错误信息
func GetErrorMsg(code ErrorCode) string {
	switch code {
	case ErrCodeSuccess:
		return "成功"
	case ErrCodeUnknown:
		return "未知错误"
	case ErrCodeInvalidParams:
		return "参数错误"
	case ErrCodeServiceInit:
		return "服务初始化失败"
	case ErrCodeServiceStart:
		return "服务启动失败"
	case ErrCodeOllamaConnect:
		return "Ollama 连接失败"
	case ErrCodeOllamaTimeout:
		return "Ollama 请求超时"
	case ErrCodeOllamaResponse:
		return "Ollama 响应错误"
	case ErrCodeOllamaModelError:
		return "模型错误"
	case ErrCodeToolNotFound:
		return "工具未找到"
	case ErrCodeToolExecute:
		return "工具执行失败"
	case ErrCodeToolParams:
		return "工具参数错误"
	case ErrCodeWeatherAPI:
		return "天气API请求失败"
	case ErrCodeWeatherCity:
		return "不支持的城市"
	case ErrCodeWeatherData:
		return "天气数据解析失败"
	default:
		return "未知错误"
	}
}
