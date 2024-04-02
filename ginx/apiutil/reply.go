package apiutil

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type Error struct {
	Code    int                    `json:"code,omitempty"`
	Reason  string                 `json:"reason,omitempty"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type Reply[T any] struct {
	Error *Error            `json:"error,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
	Data  T                 `json:"data,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func JSON(c *gin.Context, reply any) {
	c.JSON(http.StatusOK, reply)
}

func Data[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Reply[T]{Data: data})
}

func ERROR(c *gin.Context, errcode int, errmsg string) {
	c.AbortWithStatusJSON(http.StatusOK, &Reply[any]{Error: &Error{Code: errcode, Message: errmsg}})
}

func NotFound(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotFound)
}

func Unauthorized(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func BadRequest(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, &Reply[any]{Error: &Error{Code: -1, Message: err.Error()}})
}

func FATAL(c *gin.Context, err error, codes ...int) {
	logger := slog.Default()
	logger.Error("fatal response", "error", err.Error())
	code := -1
	if len(codes) > 0 {
		code = codes[0]
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, &Reply[any]{Error: &Error{Code: code, Message: err.Error()}})
}
