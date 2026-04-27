package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"people-page-backend/internal/config"
	"people-page-backend/internal/dao"
	"people-page-backend/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// SendLoginCode 发送登录/注册验证码
func SendLoginCode(phone string) (string, error) {
	if err := dao.EnsureCitizenTable(); err != nil {
		return "", fmt.Errorf("初始化用户表失败: %w", err)
	}

	code, err := generateRandomCode(6)
	if err != nil {
		return "", err
	}

	key := "sms:" + phone
	if err := dao.RDB.Set(dao.Ctx, key, code, 5*time.Minute).Err(); err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}

	return code, nil
}

// RegisterByCode 验证码注册（设置密码）
func RegisterByCode(phone, code, password string) (*model.CitizenUser, string, error) {
	// 验证验证码
	if err := verifySMSCode(phone, code); err != nil {
		return nil, "", err
	}

	// 检查用户是否存在
	existing, err := dao.GetCitizenByPhone(phone)
	if err == nil && existing != nil {
		return nil, "", fmt.Errorf("该手机号已注册，请直接登录")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user, err := dao.CreateCitizenWithPassword(phone, string(hash))
	if err != nil {
		return nil, "", fmt.Errorf("注册失败: %w", err)
	}

	// 生成 token
	token, err := generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// LoginByCode 验证码登录（首次自动创建用户）
func LoginByCode(phone, code string) (*model.CitizenUser, string, error) {
	if err := verifySMSCode(phone, code); err != nil {
		return nil, "", err
	}

	user, err := dao.GetCitizenByPhone(phone)
	if err != nil {
		// 首次验证码登录 → 自动创建用户（无需单独注册步骤）
		user, err = dao.CreateCitizen(phone)
		if err != nil {
			return nil, "", fmt.Errorf("创建用户失败: %w", err)
		}
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// LoginByPassword 密码登录
func LoginByPassword(phone, password string) (*model.CitizenUser, string, error) {
	user, err := dao.GetCitizenByPhone(phone)
	if err != nil {
		return nil, "", fmt.Errorf("该手机号未注册")
	}

	if user.PasswordHash == "" {
		return nil, "", fmt.Errorf("该账号未设置密码，请使用验证码登录")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("密码错误")
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// SetPassword 为已登录用户设置/修改密码（需验证码验证）
func SetPassword(phone, code, newPassword string) error {
	if err := verifySMSCode(phone, code); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	return dao.UpdatePassword(phone, string(hash))
}

// WechatMiniProgramLogin 微信小程序登录
// 流程：
//
//	1. 前端传 wx.login() 获得的 code
//	2. 后端调微信接口换 session_key + openid
//	3. 查数据库：
//	   a. openid 已绑定手机号 → 直接登录
//	   b. openid 未绑定（新用户）→ 返回 openid，前端引导绑定手机号
func WechatMiniProgramLogin(wxCode string) (*model.CitizenUser, string, bool, error) {
	// 调微信接口
	session := &wechatSession{}
	if err := getWechatSession(wxCode, session); err != nil {
		return nil, "", false, fmt.Errorf("微信登录失败: %w", err)
	}

	// 登录/创建
	user, isReturning, err := dao.WxLoginOrCreate(session.OpenID, session.SessionKey)
	if err != nil {
		return nil, "", false, fmt.Errorf("微信登录失败: %w", err)
	}

	if !isReturning || user.Phone == "" {
		// 新用户 or 老用户但没绑定手机号 → 需要绑定
		return user, "", false, nil
	}

	// 已有完整用户（有手机号）→ 生成 token 直接登录
	token, err := generateToken(user)
	if err != nil {
		return nil, "", false, fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, true, nil
}

// BindPhoneToWechat 将手机号绑定到微信用户
// 场景：微信登录后，用户（未绑定手机号）输入手机号+验证码完成绑定
func BindPhoneToWechat(wxOpenid, phone, code string) (*model.CitizenUser, string, error) {
	// 验证验证码
	if err := verifySMSCode(phone, code); err != nil {
		return nil, "", err
	}

	// 查找该微信用户
	_, err := dao.GetCitizenByWxOpenid(wxOpenid)
	if err != nil {
		return nil, "", fmt.Errorf("微信用户不存在，请先微信登录")
	}

	// 检查该手机号是否已被其他账号绑定
	existing, err := dao.GetCitizenByPhone(phone)
	if err == nil && existing != nil {
		// 手机号已存在 → 把微信 openid 绑定到已有账号
		if err := dao.UpdateCitizenWxOpenid(existing.ID, wxOpenid); err != nil {
			return nil, "", fmt.Errorf("绑定失败: %w", err)
		}
		// 删除旧的微信临时用户（可选，保留以避免数据丢失）
		token, err := generateToken(existing)
		if err != nil {
			return nil, "", fmt.Errorf("生成token失败: %w", err)
		}
		return existing, token, nil
	}

	// 手机号不存在 → 更新微信用户的手机号
	if err := dao.BindPhoneToWx(phone, wxOpenid); err != nil {
		return nil, "", fmt.Errorf("绑定手机号失败: %w", err)
	}

	// 重新获取完整用户
	user, err := dao.GetCitizenByWxOpenid(wxOpenid)
	if err != nil {
		return nil, "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// BindWechatToCurrentUser 已登录用户绑定微信
func BindWechatToCurrentUser(userID int, wxOpenid string) error {
	return dao.UpdateCitizenWxOpenid(userID, wxOpenid)
}

// ValidateToken 验证 token，返回用户信息
func ValidateToken(token string) *model.CitizenUser {
	key := "auth_token:" + token
	data, err := dao.RDB.Get(dao.Ctx, key).Bytes()
	if err != nil {
		return nil
	}

	var user model.CitizenUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil
	}
	return &user
}

// Logout 登出，删除 token
func Logout(token string) {
	key := "auth_token:" + token
	dao.RDB.Del(dao.Ctx, key)
}

// GetCitizenProfile 获取用户信息
func GetCitizenProfile(userID int) (*model.CitizenUser, error) {
	return nil, fmt.Errorf("功能待实现")
}

// CheckPhoneRegistered 检查手机号是否已注册
func CheckPhoneRegistered(phone string) (bool, error) {
	_, err := dao.GetCitizenByPhone(phone)
	if err != nil {
		return false, nil // 不存在也算"正常"，返回 false
	}
	return true, nil
}

// CheckHasPassword 检查用户是否设置了密码
func CheckHasPassword(phone string) (bool, error) {
	return dao.HasPassword(phone)
}

// WechatWebLogin 微信开放平台扫码登录
// 流程：微信扫码后回调 → 带 code 参数 → 后端换 access_token + openid → 登录
func WechatWebLogin(wxCode string) (*model.CitizenUser, string, bool, error) {
	cfg := config.AppConfig.Wechat
	if cfg.Open.AppID == "" || cfg.Open.AppSecret == "" {
		return nil, "", false, fmt.Errorf("微信开放平台未配置（请设置 app_id 和 app_secret）")
	}

	// 1. 调微信开放平台接口，用 code 换 access_token 和 openid
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(cfg.Open.AppID),
		url.QueryEscape(cfg.Open.AppSecret),
		url.QueryEscape(wxCode),
	)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, "", false, fmt.Errorf("调用微信开放平台接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", false, fmt.Errorf("读取微信响应失败: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, "", false, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return nil, "", false, fmt.Errorf("微信开放平台错误(%d): %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// 2. 可选：获取用户基本信息（昵称、头像）
	nickname := ""
	avatar := ""
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(tokenResp.AccessToken),
		url.QueryEscape(tokenResp.OpenID),
	)
	userInfoResp, err := http.Get(userInfoURL)
	if err == nil {
		defer userInfoResp.Body.Close()
		userInfoBody, _ := io.ReadAll(userInfoResp.Body)
		var userInfo struct {
			Nickname  string `json:"nickname"`
			Headimgurl string `json:"headimgurl"`
			ErrCode   int    `json:"errcode"`
		}
		if json.Unmarshal(userInfoBody, &userInfo) == nil && userInfo.ErrCode == 0 {
			nickname = userInfo.Nickname
			avatar = userInfo.Headimgurl
		}
	}

	// 3. 查找或创建用户
	user, isReturning, err := dao.WxLoginOrCreate(tokenResp.OpenID, "")
	if err != nil {
		return nil, "", false, fmt.Errorf("微信登录失败: %w", err)
	}

	// 更新昵称和头像（如果有）
	if nickname != "" && user.Nickname == "" {
		dao.UpdateCitizenProfile(user.ID, nickname, avatar)
		user.Nickname = nickname
		user.Avatar = avatar
	}

	if !isReturning || user.Phone == "" {
		// 新用户 or 老用户没绑定手机 → 需要绑定
		return user, "", false, nil
	}

	// 已有完整用户 → 生成 token 直接登录
	token, err := generateToken(user)
	if err != nil {
		return nil, "", false, fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, true, nil
}

// ==== 辅助函数 ====

// wechatSession 微信小程序登录返回值
type wechatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// getWechatSession 调微信接口获取 session
func getWechatSession(code string, session *wechatSession) error {
	cfg := config.AppConfig.Wechat
	if cfg.Miniprogram.AppID == "" || cfg.Miniprogram.AppSecret == "" {
		return fmt.Errorf("微信小程序未配置（请设置 app_id 和 app_secret）")
	}

	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(cfg.Miniprogram.AppID),
		url.QueryEscape(cfg.Miniprogram.AppSecret),
		url.QueryEscape(code),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("调用微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取微信响应失败: %w", err)
	}

	if err := json.Unmarshal(body, session); err != nil {
		return fmt.Errorf("解析微信响应失败: %w", err)
	}

	if session.ErrCode != 0 {
		return fmt.Errorf("微信接口错误(%d): %s", session.ErrCode, session.ErrMsg)
	}

	return nil
}

// verifySMSCode 验证并消耗验证码
func verifySMSCode(phone, code string) error {
	key := "sms:" + phone
	storedCode, err := dao.RDB.Get(dao.Ctx, key).Result()
	if err != nil {
		return fmt.Errorf("验证码错误或已过期")
	}
	if storedCode != code {
		return fmt.Errorf("验证码错误")
	}
	dao.RDB.Del(dao.Ctx, key)
	return nil
}

// generateToken 生成登录 token 并存入 Redis（7天有效）
func generateToken(user *model.CitizenUser) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("序列化用户信息失败: %w", err)
	}

	key := "auth_token:" + token
	if err := dao.RDB.Set(dao.Ctx, key, userJSON, 7*24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("存储token失败: %w", err)
	}

	return token, nil
}

func generateRandomCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
