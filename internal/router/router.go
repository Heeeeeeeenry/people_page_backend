package router

import (
	"github.com/gin-gonic/gin"
	"people-page-backend/internal/controller"
	"people-page-backend/internal/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS中间件
	r.Use(corsMiddleware())

	api := r.Group("/api")
	{
		// 认证相关（无需登录）
		api.POST("/auth/send-code", controller.SendCode)
		api.POST("/auth/register", controller.Register)           // 验证码注册
		api.POST("/auth/login", controller.Login)                 // 验证码登录
		api.POST("/auth/login/password", controller.LoginByPassword) // 密码登录
		api.POST("/auth/check-phone", controller.CheckPhone)      // 检查手机号

		// 微信相关（无需登录）
		api.POST("/auth/wechat/login", controller.WechatLogin)    // 微信扫码登录
		api.POST("/auth/wechat/web-login", controller.WechatWebLogin) // 微信开放平台扫码登录（Web端）
		api.GET("/auth/wechat/callback", controller.WechatWebCallback) // 微信开放平台回调（OAuth2 redirect_uri）
		api.POST("/auth/wechat/bind-phone", controller.BindPhone) // 微信绑定手机号

		// 提示词
		api.GET("/prompt", controller.GetPrompt)

		// 对话流
		api.POST("/chat/stream", controller.ChatStream)

		// 需要登录的接口
		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		{
			// 提交信件
			auth.POST("/letter/submit", controller.SubmitLetter)
			// 获取分类
			auth.GET("/letter/categories", controller.GetCategories)
			// AI智能分类
			auth.POST("/letter/classify", controller.ClassifyLetter)

			// 用户相关
			auth.POST("/auth/logout", controller.Logout)
			auth.GET("/auth/me", controller.GetCurrentUser)
			auth.POST("/auth/set-password", controller.SetPassword)   // 设置密码
			auth.POST("/auth/wechat/bind", controller.BindWechat)     // 绑定微信到当前用户
		}

		// 高德地图（公开接口）
		api.GET("/amap/poi/search", controller.SearchPOI)
		api.GET("/amap/geocode", controller.Geocode)
		api.GET("/amap/regeocode", controller.Regeocode)
		api.GET("/amap/poi/around", controller.SearchPOIAround)
		api.GET("/amap/input/tips", controller.GetInputTips)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
