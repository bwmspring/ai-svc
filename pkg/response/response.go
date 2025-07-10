package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// 常用响应码
const (
	SUCCESS              = 200 // 成功
	ERROR                = 500 // 系统错误
	INVALID_PARAMS       = 400 // 参数错误
	UNAUTHORIZED         = 401 // 未授权
	FORBIDDEN            = 403 // 禁止访问
	NOT_FOUND            = 404 // 资源不存在
	METHOD_NOT_ALLOWED   = 405 // 方法不允许
	CONFLICT             = 409 // 冲突
	UNPROCESSABLE_ENTITY = 422 // 无法处理的实体
	TOO_MANY_REQUESTS    = 429 // 请求过多
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Page 分页响应
func Page(c *gin.Context, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    SUCCESS,
		Message: "查询成功",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// PageWithMessage 带消息的分页响应
func PageWithMessage(c *gin.Context, message string, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    SUCCESS,
		Message: message,
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// getHTTPStatus 根据业务状态码获取HTTP状态码
func getHTTPStatus(code int) int {
	switch code {
	case SUCCESS:
		return http.StatusOK
	case INVALID_PARAMS:
		return http.StatusBadRequest
	case UNAUTHORIZED:
		return http.StatusUnauthorized
	case FORBIDDEN:
		return http.StatusForbidden
	case NOT_FOUND:
		return http.StatusNotFound
	case METHOD_NOT_ALLOWED:
		return http.StatusMethodNotAllowed
	case CONFLICT:
		return http.StatusConflict
	case UNPROCESSABLE_ENTITY:
		return http.StatusUnprocessableEntity
	case TOO_MANY_REQUESTS:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
