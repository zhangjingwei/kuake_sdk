package validation

import "fmt"

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeBusiness
	ErrorTypeSystem
)

// ValidationError 校验错误
type ValidationError struct {
	Code    string
	Message string
	Type    ErrorType
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ErrInvalidArgument 创建校验错误
func ErrInvalidArgument(msg string) *ValidationError {
	return &ValidationError{
		Code:    "INVALID_ARG",
		Message: msg,
		Type:    ErrorTypeValidation,
	}
}

// ErrInvalidPath 创建路径格式/内容非法错误（非穿越类）
func ErrInvalidPath(msg string) *ValidationError {
	return &ValidationError{
		Code:    "FILE_INVALID_PATH",
		Message: msg,
		Type:    ErrorTypeValidation,
	}
}

// ErrUserInvalidInput 创建用户输入非法错误
func ErrUserInvalidInput(msg string) *ValidationError {
	return &ValidationError{
		Code:    "USER_INVALID_INPUT",
		Message: msg,
		Type:    ErrorTypeValidation,
	}
}

// ErrPathTraversal 创建路径穿越错误
func ErrPathTraversal(path string) *ValidationError {
	return &ValidationError{
		Code:    "FILE_PATH_TRAVERSAL",
		Message: fmt.Sprintf("路径包含非法段（如路径穿越）：%s", path),
		Type:    ErrorTypeValidation,
	}
}
