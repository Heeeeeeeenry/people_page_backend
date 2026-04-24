package service

import "people-page-backend/internal/model"

// CitizenUserContext 传递给后端服务的用户上下文
type CitizenUserContext struct {
	User  *model.CitizenUser
	Token string
}

// LoginResponse 登录返回结果
type LoginResponse struct {
	Token    string            `json:"token"`
	User     *model.CitizenUser `json:"user"`
}
