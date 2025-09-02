package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// AppError 应用错误类型
type AppError struct {
	Code    int
	Message string
	Details string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// 预定义错误类型
var (
	ErrBadRequest          = &AppError{Code: http.StatusBadRequest, Message: "请求参数错误"}
	ErrUnauthorized        = &AppError{Code: http.StatusUnauthorized, Message: "未授权访问"}
	ErrForbidden           = &AppError{Code: http.StatusForbidden, Message: "禁止访问"}
	ErrNotFound            = &AppError{Code: http.StatusNotFound, Message: "资源不存在"}
	ErrConflict            = &AppError{Code: http.StatusConflict, Message: "资源冲突"}
	ErrInternalServerError = &AppError{Code: http.StatusInternalServerError, Message: "服务器内部错误"}
)

// NewAppError 创建应用错误
func NewAppError(code int, message string, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewAppErrorWithErr 创建包含原始错误的应用错误
func NewAppErrorWithErr(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrorHandler 统一错误处理中间件（用于panic恢复）
func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			handleError(c, err, logger)
		} else {
			handleError(c, ErrInternalServerError, logger)
		}
	})
}

// ErrorMiddleware 错误处理中间件（用于手动错误处理）
func ErrorMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err, logger)
			c.Abort()
		}
	}
}

// HandleError 手动处理错误的辅助函数
func HandleError(c *gin.Context, err error) {
	c.Error(err)
}

func handleError(c *gin.Context, err error, logger *logrus.Logger) {
	// 记录错误日志
	logFields := logrus.Fields{
		"method":     c.Request.Method,
		"uri":        c.Request.RequestURI,
		"user_agent": c.Request.UserAgent(),
		"ip":         c.ClientIP(),
		"error":      err.Error(),
	}

	// 检查是否是自定义应用错误
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= 500 {
			// 服务器错误，使用 ERROR 级别
			logger.WithFields(logFields).Error("Server error")
		} else {
			// 客户端错误，使用 WARN 级别
			logger.WithFields(logFields).Warn("Client error")
		}
		c.JSON(appErr.Code, ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// 检查是否是 Gin 绑定错误
	if bindErr, ok := err.(*gin.Error); ok {
		code := http.StatusBadRequest
		message := "请求参数错误"

		logger.WithFields(logFields).Warn("Binding error")

		c.JSON(code, ErrorResponse{
			Code:    code,
			Message: message,
			Details: bindErr.Error(),
		})
		return
	}

	// 未知错误，返回 500
	logger.WithFields(logFields).Error("Unknown error")
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "服务器内部错误",
	})
}

// getHTTPErrorMessage 获取 HTTP 错误消息
func getHTTPErrorMessage(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "请求参数错误"
	case http.StatusUnauthorized:
		return "未授权访问"
	case http.StatusForbidden:
		return "禁止访问"
	case http.StatusNotFound:
		return "资源不存在"
	case http.StatusMethodNotAllowed:
		return "请求方法不允许"
	case http.StatusConflict:
		return "资源冲突"
	case http.StatusUnprocessableEntity:
		return "请求数据格式错误"
	case http.StatusTooManyRequests:
		return "请求过于频繁"
	case http.StatusInternalServerError:
		return "服务器内部错误"
	case http.StatusBadGateway:
		return "网关错误"
	case http.StatusServiceUnavailable:
		return "服务暂不可用"
	case http.StatusGatewayTimeout:
		return "网关超时"
	default:
		return "未知错误"
	}
}
