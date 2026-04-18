package response

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess            = 0
	CodeBadRequest         = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeConflict           = 409
	CodeInternalError      = 500
	CodeServiceUnavailable = 503
)

var codeMessages = map[int]string{
	CodeSuccess:            "success",
	CodeBadRequest:         "bad request",
	CodeUnauthorized:       "unauthorized",
	CodeForbidden:          "forbidden",
	CodeNotFound:           "not found",
	CodeConflict:           "conflict",
	CodeInternalError:      "internal server error",
	CodeServiceUnavailable: "service unavailable",
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string, detail ...string) {
	errDetail := ""
	if len(detail) > 0 {
		errDetail = detail[0]
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Detail:  errDetail,
	})
}

func BadRequest(c *gin.Context, message string, detail ...string) {
	Error(c, http.StatusOK, CodeBadRequest, message, detail...)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = codeMessages[CodeUnauthorized]
	}
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = codeMessages[CodeForbidden]
	}
	Error(c, http.StatusOK, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = codeMessages[CodeNotFound]
	}
	Error(c, http.StatusOK, CodeNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = codeMessages[CodeConflict]
	}
	Error(c, http.StatusOK, CodeConflict, message)
}

func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = codeMessages[CodeInternalError]
	}
	Error(c, http.StatusOK, CodeInternalError, message)
}

func SerializeJSON(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func DeserializeJSON(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}
