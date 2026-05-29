package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API response codes
const (
	CodeSuccess       = 200
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeInternalError = 500
	CodeServiceUnavailable = 503
)

// Standard error codes
const (
	ErrBadRequest          = 400
	ErrUnauthorized        = 401
	ErrForbidden           = 403
	ErrNotFound            = 404
	ErrInternalServer      = 500
	ErrServiceUnavailable  = 503
)

// Response is the standard API response envelope
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse is the standard paginated response envelope
type PaginatedResponse struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success returns a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// Paginated returns a paginated success response
func Paginated(c *gin.Context, data interface{}, total int, page int, pageSize int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:     CodeSuccess,
		Message:  "success",
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
