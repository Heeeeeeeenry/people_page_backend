package controller

import (
	"net/http"
	"net/url"
	"strings"

	"people-page-backend/internal/middleware"
	"people-page-backend/internal/model"
	"people-page-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// 包级变量实例，供路由引用
var (
	authCtrl = &AuthController{}

	// 包级函数（兼容路由直接引用）
	SendCode       = authCtrl.sendCode
	Login          = authCtrl.loginByCode
	Logout         = authCtrl.logout
	GetCurrentUser = authCtrl.profile
	Register       = authCtrl.registerByCode
	LoginByPassword = authCtrl.loginByPassword
	SetPassword    = authCtrl.setPassword
	WechatLogin    = authCtrl.wechatLogin
	BindPhone      = authCtrl.bindPhone
	BindWechat     = authCtrl.bindWechat
	CheckPhone     = authCtrl.checkPhone
	WechatWebLogin = authCtrl.wechatWebLogin
	WechatWebCallback = authCtrl.wechatWebCallback
)

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// RegisterByCodeRequest 验证码注册请求
type RegisterByCodeRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginByCodeRequest 验证码登录请求
type LoginByCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// LoginByPasswordRequest 密码登录请求
type LoginByPasswordRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SetPasswordRequest 设置密码请求
type SetPasswordRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// BindPhoneRequest 微信绑定手机号请求
type BindPhoneRequest struct {
	WxOpenid string `json:"wx_openid" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// BindWechatRequest 绑定微信到当前用户
type BindWechatRequest struct {
	WxOpenid string `json:"wx_openid" binding:"required"`
}

// WechatWebLoginRequest 微信开放平台扫码登录请求
type WechatWebLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// CheckPhoneRequest 检查手机号是否注册
type CheckPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// LoginResponse 登录返回
type LoginResponse struct {
	Token    string           `json:"token"`
	User     *model.CitizenUser `json:"user"`
	Message  string           `json:"message,omitempty"`
}

// sendCode 发送验证码
func (ac *AuthController) sendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供手机号"})
		return
	}

	code, err := service.SendLoginCode(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": "验证码已发送",
	})
}

// registerByCode 验证码注册（设置密码）
func (ac *AuthController) registerByCode(c *gin.Context) {
	var req RegisterByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := "请提供手机号、验证码和密码（至少6位）"
		c.JSON(http.StatusBadRequest, gin.H{"message": msg})
		return
	}

	user, token, err := service.RegisterByCode(req.Phone, req.Code, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// loginByCode 验证码登录
func (ac *AuthController) loginByCode(c *gin.Context) {
	var req LoginByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供手机号和验证码"})
		return
	}

	user, token, err := service.LoginByCode(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// loginByPassword 密码登录
func (ac *AuthController) loginByPassword(c *gin.Context) {
	var req LoginByPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供手机号和密码"})
		return
	}

	user, token, err := service.LoginByPassword(req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// setPassword 设置/修改密码（需验证码验证）
func (ac *AuthController) setPassword(c *gin.Context) {
	var req SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := "请提供手机号、验证码和新密码（至少6位）"
		c.JSON(http.StatusBadRequest, gin.H{"message": msg})
		return
	}

	if err := service.SetPassword(req.Phone, req.Code, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码设置成功"})
}

// wechatWebCallback 微信开放平台扫码回调（供 WeChat 重定向用）
// 注册到微信开放平台的 redirect_uri：https://YOUR_DOMAIN/api/auth/wechat/callback
func (ac *AuthController) wechatWebCallback(c *gin.Context) {
	code := c.Query("code")

	frontendBase := "http://localhost:5174"

	if code == "" {
		c.Redirect(http.StatusFound, frontendBase+"/auth/wechat/callback?error="+url.QueryEscape("缺少code参数"))
		return
	}

	user, token, loggedIn, err := service.WechatWebLogin(code)
	if err != nil {
		c.Redirect(http.StatusFound, frontendBase+"/auth/wechat/callback?error="+url.QueryEscape(err.Error()))
		return
	}

	if !loggedIn {
		// 新用户需要绑定手机号
		c.Redirect(http.StatusFound, frontendBase+"/auth/wechat/callback?need_bind_phone=true&wx_openid="+user.WxOpenid)
		return
	}

	// 登录成功，重定向到前端回调页面
	c.Redirect(http.StatusFound, frontendBase+"/auth/wechat/callback?token="+token)
}

// wechatWebLogin 微信开放平台扫码登录（Web端 JSON API，供前端回调页调用）
func (ac *AuthController) wechatWebLogin(c *gin.Context) {
	var req WechatWebLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供微信登录code"})
		return
	}

	user, token, loggedIn, err := service.WechatWebLogin(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !loggedIn {
		// 需要绑定手机号
		c.JSON(http.StatusOK, gin.H{
			"need_bind_phone": true,
			"wx_openid":       user.WxOpenid,
			"message":         "请绑定手机号完成注册",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// wechatLogin 微信小程序登录
func (ac *AuthController) wechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供微信登录code"})
		return
	}

	user, token, loggedIn, err := service.WechatMiniProgramLogin(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !loggedIn {
		// 需要绑定手机号
		c.JSON(http.StatusOK, gin.H{
			"need_bind_phone": true,
			"wx_openid":       user.WxOpenid,
			"message":         "请绑定手机号完成注册",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// bindPhone 微信绑定手机号
func (ac *AuthController) bindPhone(c *gin.Context) {
	var req BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供微信openid、手机号和验证码"})
		return
	}

	user, token, err := service.BindPhoneToWechat(req.WxOpenid, req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// bindWechat 已登录用户绑定微信
func (ac *AuthController) bindWechat(c *gin.Context) {
	// 从 token 获取当前用户
	tokenStr := c.GetHeader("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	user := middleware.GetCitizenUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "请先登录"})
		return
	}

	// 验证 token 一致性（安全校验）
	if tokenStr != "" {
		u := service.ValidateToken(tokenStr)
		if u == nil || u.ID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "登录已过期，请重新登录"})
			return
		}
	}

	var req BindWechatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供微信openid"})
		return
	}

	if err := service.BindWechatToCurrentUser(user.ID, req.WxOpenid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "微信绑定成功"})
}

// checkPhone 检查手机号是否已注册
func (ac *AuthController) checkPhone(c *gin.Context) {
	var req CheckPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请提供手机号"})
		return
	}

	registered, err := service.CheckPhoneRegistered(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败"})
		return
	}

	hasPassword, _ := service.CheckHasPassword(req.Phone)

	c.JSON(http.StatusOK, gin.H{
		"registered":   registered,
		"has_password": hasPassword,
	})
}

// logout 登出
func (ac *AuthController) logout(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	if tokenStr != "" {
		service.Logout(tokenStr)
	}

	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

// profile 获取当前用户信息
func (ac *AuthController) profile(c *gin.Context) {
	user := middleware.GetCitizenUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}
	c.JSON(http.StatusOK, user)
}
