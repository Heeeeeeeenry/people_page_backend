package main

import (
	"fmt"
	"log"
	"people-page-backend/internal/config"
	"people-page-backend/internal/dao"
	"people-page-backend/internal/router"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 连接数据库
	if err := dao.InitDB(); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer dao.CloseDB()

	// 连接 Redis（验证码/Token 缓存）
	if err := dao.InitRedis(); err != nil {
		log.Fatalf("连接Redis失败: %v", err)
	}
	defer dao.CloseRedis()

	// 确保市民用户表存在
	if err := dao.EnsureCitizenTable(); err != nil {
		log.Fatalf("创建市民用户表失败: %v", err)
	}

	// 设置路由
	r := router.SetupRouter()

	// 启动服务
	addr := fmt.Sprintf(":%s", config.AppConfig.Server.Port)
	log.Printf("服务器启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
