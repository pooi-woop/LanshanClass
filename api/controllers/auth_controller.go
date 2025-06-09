// FilePath: C:/LanshanClass1.3/api/controllers\auth_controller.go
package controllers

import (
	"LanshanClass1.3/proto"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthServiceClient 是 gRPC 客户端
var AuthServiceClient proto.AuthServiceClient

// Register 注册接口
func Register(c *gin.Context) {
	var req proto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := AuthServiceClient.Register(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": resp.Token, "message": resp.Message})
}

// Login 登录接口
func Login(c *gin.Context) {
	var req proto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := AuthServiceClient.Login(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": resp.Token, "message": resp.Message})
}
