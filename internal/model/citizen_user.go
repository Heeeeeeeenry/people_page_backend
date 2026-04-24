package model

// CitizenUser 市民用户表
type CitizenUser struct {
	ID           int    `json:"id" db:"id"`
	Phone        string `json:"phone" db:"phone"`
	Nickname     string `json:"nickname" db:"nickname"`
	Avatar       string `json:"avatar" db:"avatar"`
	PasswordHash string `json:"-" db:"password_hash"`
	WxOpenid     string `json:"wx_openid" db:"wx_openid"`
	WxSessionKey string `json:"-" db:"wx_session_key"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}
