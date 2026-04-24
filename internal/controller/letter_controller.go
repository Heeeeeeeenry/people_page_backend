package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"people-page-backend/internal/model"
	"people-page-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// GetPrompt 获取系统提示词
func GetPrompt(c *gin.Context) {
	prompt, err := service.GetSystemPrompt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"prompt": prompt})
}

// ChatStream 流式对话
func ChatStream(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	var req struct {
		Messages []map[string]interface{} `json:"messages"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求体失败"})
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持SSE"})
		return
	}

	if err := service.ChatStream(req.Messages, c.Writer, flusher); err != nil {
		// If we already started streaming, the error message won't work well
		// But try to signal error anyway
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(c.Writer, "data: %s\n\n", errData)
		flusher.Flush()
	}
}

// SubmitLetter 提交信件（需登录，自动填入手机号）
func SubmitLetter(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求体失败"})
		return
	}

	// 获取当前登录用户
	user, _ := c.Get("citizen_user")
	if user != nil {
		if u, ok := user.(*model.CitizenUser); ok {
			data["登录手机号"] = u.Phone
		}
	}

	result, err := service.SubmitLetterCitizen(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
