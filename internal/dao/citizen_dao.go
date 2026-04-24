package dao

import (
	"fmt"
	"time"
	"people-page-backend/internal/model"
)

// EnsureCitizenTable 确保市民用户表存在（含 password_hash 和 wx_session_key 列）
func EnsureCitizenTable() error {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS citizen_users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		phone VARCHAR(20) NOT NULL DEFAULT '',
		nickname VARCHAR(50) DEFAULT '',
		avatar VARCHAR(255) DEFAULT '',
		password_hash VARCHAR(255) DEFAULT '',
		wx_openid VARCHAR(100) DEFAULT '',
		wx_session_key VARCHAR(100) DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uk_phone (phone),
		UNIQUE KEY uk_wx_openid (wx_openid)
	)`)
	return err
}

// CreateCitizen 创建市民用户（仅手机号）
func CreateCitizen(phone string) (*model.CitizenUser, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		"INSERT INTO citizen_users (phone, created_at, updated_at) VALUES (?, ?, ?)",
		phone, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return GetCitizenByPhone(phone)
}

// CreateCitizenWithPassword 创建市民用户（手机号 + 密码）
func CreateCitizenWithPassword(phone, passwordHash string) (*model.CitizenUser, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		"INSERT INTO citizen_users (phone, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
		phone, passwordHash, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("注册用户失败: %w", err)
	}
	return GetCitizenByPhone(phone)
}

// GetCitizenByPhone 根据手机号查询市民用户
func GetCitizenByPhone(phone string) (*model.CitizenUser, error) {
	user := &model.CitizenUser{}
	err := DB.QueryRow(
		"SELECT id, phone, nickname, avatar, password_hash, wx_openid, wx_session_key, created_at, updated_at FROM citizen_users WHERE phone = ?",
		phone,
	).Scan(&user.ID, &user.Phone, &user.Nickname, &user.Avatar, &user.PasswordHash, &user.WxOpenid, &user.WxSessionKey, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetCitizenByWxOpenid 根据微信 openid 查询用户
func GetCitizenByWxOpenid(wxOpenid string) (*model.CitizenUser, error) {
	user := &model.CitizenUser{}
	err := DB.QueryRow(
		"SELECT id, phone, nickname, avatar, password_hash, wx_openid, wx_session_key, created_at, updated_at FROM citizen_users WHERE wx_openid = ?",
		wxOpenid,
	).Scan(&user.ID, &user.Phone, &user.Nickname, &user.Avatar, &user.PasswordHash, &user.WxOpenid, &user.WxSessionKey, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateCitizenWxOpenid 更新微信 openid（绑定/解绑）
func UpdateCitizenWxOpenid(userID int, wxOpenid string) error {
	_, err := DB.Exec(
		"UPDATE citizen_users SET wx_openid = ? WHERE id = ?",
		wxOpenid, userID,
	)
	return err
}

// UpdateCitizenWxSession 更新微信 session_key（用于小程序会话维持）
func UpdateCitizenWxSession(phone, sessionKey string) error {
	_, err := DB.Exec(
		"UPDATE citizen_users SET wx_session_key = ? WHERE phone = ?",
		sessionKey, phone,
	)
	return err
}

// UpdateCitizenProfile 更新用户资料
func UpdateCitizenProfile(userID int, nickname, avatar string) error {
	_, err := DB.Exec(
		"UPDATE citizen_users SET nickname = ?, avatar = ? WHERE id = ?",
		nickname, avatar, userID,
	)
	return err
}

// UpdatePassword 更新密码
func UpdatePassword(phone, passwordHash string) error {
	_, err := DB.Exec(
		"UPDATE citizen_users SET password_hash = ? WHERE phone = ?",
		passwordHash, phone,
	)
	return err
}

// BindPhoneToWx 将手机号绑定到微信 openid（如果 openid 已存在则更新 phone）
func BindPhoneToWx(phone, wxOpenid string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		"UPDATE citizen_users SET phone = ?, updated_at = ? WHERE wx_openid = ?",
		phone, now, wxOpenid,
	)
	return err
}

// WxLoginOrCreate 微信登录：如果 openid 已绑定用户则返回，否则创建新用户（无手机号）
func WxLoginOrCreate(wxOpenid, sessionKey string) (*model.CitizenUser, bool, error) {
	// 先查是否已绑定
	user, err := GetCitizenByWxOpenid(wxOpenid)
	if err == nil {
		// 已存在用户，更新 session_key
		UpdateCitizenWxSession(user.Phone, sessionKey)
		return user, true, nil // isReturning = true
	}

	// 不存在，创建临时用户（无手机号，以后绑定）
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := DB.Exec(
		"INSERT INTO citizen_users (wx_openid, wx_session_key, created_at, updated_at) VALUES (?, ?, ?, ?)",
		wxOpenid, sessionKey, now, now,
	)
	if err != nil {
		return nil, false, fmt.Errorf("创建微信用户失败: %w", err)
	}
	id, _ := result.LastInsertId()
	user = &model.CitizenUser{
		ID:           int(id),
		WxOpenid:     wxOpenid,
		WxSessionKey: sessionKey,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return user, false, nil // isReturning = false（新用户，还没手机号）
}

// HasPassword 检查用户是否设置了密码
func HasPassword(phone string) (bool, error) {
	var hash string
	err := DB.QueryRow(
		"SELECT password_hash FROM citizen_users WHERE phone = ?",
		phone,
	).Scan(&hash)
	if err != nil {
		return false, err
	}
	return hash != "", nil
}
