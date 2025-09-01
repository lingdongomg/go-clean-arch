package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
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

// NewAppErrorWithErr 创建带原始错误的应用错误
func NewAppErrorWithErr(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrorHandler 统一错误处理中间件
func ErrorHandler(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return handleError(c, err, logger)
			}
			return nil
		}
	}
}

// handleError 处理错误
func handleError(c echo.Context, err error, logger *logrus.Logger) error {
	var appErr *AppError
	var httpErr *echo.HTTPError

	// 判断错误类型
	switch {
	case errors.As(err, &appErr):
		// 应用自定义错误
		logError(logger, c, appErr, appErr.Code >= 500)
		return c.JSON(appErr.Code, ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		})

	case errors.As(err, &httpErr):
		// Echo HTTP 错误
		code := httpErr.Code
		message := getHTTPErrorMessage(code)

		logError(logger, c, err, code >= 500)
		return c.JSON(code, ErrorResponse{
			Code:    code,
			Message: message,
		})

	default:
		// 未知错误，统一返回 500
		logError(logger, c, err, true)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "服务器内部错误",
		})
	}
}

// logError 记录错误日志
func logError(logger *logrus.Logger, c echo.Context, err error, isServerError bool) {
	fields := logrus.Fields{
		"method":     c.Request().Method,
		"uri":        c.Request().RequestURI,
		"user_agent": c.Request().UserAgent(),
		"remote_ip":  c.RealIP(),
	}

	if isServerError {
		logger.WithFields(fields).WithError(err).Error("Server error occurred")
	} else {
		logger.WithFields(fields).WithError(err).Warn("Client error occurred")
	}
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
