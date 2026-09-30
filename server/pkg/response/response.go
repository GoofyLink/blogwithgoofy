package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data})
}

func Page(c *gin.Context, list any, total int64) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: gin.H{"list": list, "total": total}})
}

func Fail(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Body{Code: 1, Msg: msg})
}

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, msg)
}

func ServerError(c *gin.Context, err error) {
	Fail(c, http.StatusInternalServerError, "服务器内部错误: "+err.Error())
}
