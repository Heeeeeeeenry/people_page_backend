package middleware

import (
	"net/http"
	"strings"
	"people-page-backend/internal/model"
	"people-page-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthRequired 需要登录认证的中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录或token已过期"})
			return
		}

		// 支持 "Bearer <token>" 格式
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		user := service.ValidateToken(token)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token无效或已过期"})
			return
		}

		// 将用户信息存入上下文
		c.Set("citizen_user", user)
		c.Set("citizen_token", token)
		c.Next()
	}
}

// GetCitizenUser 从上下文获取当前登录的市民用户
func GetCitizenUser(c *gin.Context) *model.CitizenUser {
	user, _ := c.Get("citizen_user")
	if user == nil {
		return nil
	}
	return user.(*model.CitizenUser)
}

// GetCitizenToken 从上下文获取当前token
func GetCitizenToken(c *gin.Context) string {
	token, _ := c.Get("citizen_token")
	if token == nil {
		return ""
	}
	return token.(string)
}
