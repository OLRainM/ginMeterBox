package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// R 统一响应结构
type R struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// OK 成功响应（带数据）
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, R{Success: true, Data: data})
}

// Created 创建成功
func Created(c *gin.Context, data any, msg string) {
	c.JSON(http.StatusCreated, R{Success: true, Data: data, Message: msg})
}

// OKMsg 成功响应（带消息）
func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, R{Success: true, Message: msg})
}

// OKData 成功响应（带数据和额外字段）
func OKData(c *gin.Context, data gin.H) {
	data["success"] = true
	c.JSON(http.StatusOK, data)
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, R{Success: false, Error: err})
}

// NotFound 资源不存在
func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, R{Success: false, Error: err})
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, R{Success: false, Error: err})
}
