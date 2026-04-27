# 民意智感中心 - 市民端后端服务 (Golang)

## 项目概述

市民端后端服务，提供市民注册、登录、提交信件等功能。

- 框架: Gin
- 数据库: MySQL + Redis
- 端口: 8081
- 认证: Bearer Token（Authorization header）
- 缓存: Redis（验证码、Token）

---

## 目录结构

```
people_page_backend/
├── cmd/server/main.go        # 入口
├── config/
│   └── config.yaml           # 配置（DB, Redis, LLM, CORS等）
├── internal/
│   ├── config/config.go      # 配置加载（含 WORK_ENV 环境覆盖）
│   ├── controller/           # API 控制器
│   ├── dao/                  # 数据访问层（MySQL + Redis）
│   ├── middleware/auth.go    # Token 认证中间件
│   ├── model/                # 数据模型
│   └── service/              # 业务逻辑层
├── go.mod
└── README.md
```

---

## 快速开始

### 1. 环境要求

- Go 1.25+
- MySQL 5.7+ / 8.0
- Redis 6+

### 2. 修改配置

编辑 `config/config.yaml`，主要修改项：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: "000000"
  name: letter_manage_db

redis:
  host: 127.0.0.1
  port: 6379

llm:
  api_key: "your_deepseek_api_key"
```

### 3. 启动 Redis

```bash
brew services start redis
```

### 4. 启动服务

```bash
# 构建
go build -o server ./cmd/server

# 运行
WORK_ENV=home ./server
```

服务启动后监听 `http://localhost:8081`

---

## API 路由

| 路径 | 说明 | 认证 |
|------|------|------|
| POST /api/auth/send-code | 发送验证码 | 否 |
| POST /api/auth/register | 注册（验证码+密码） | 否 |
| POST /api/auth/login | 验证码登录 | 否 |
| POST /api/auth/login/password | 密码登录 | 否 |
| POST /api/auth/logout | 登出 | 是 |
| GET  /api/auth/me | 获取当前用户 | 是 |
| POST /api/letter/submit | 提交信件 | 是 |
| POST /api/letter/classify | AI 智能分类 | 是 |
| GET  /api/letter/categories | 获取分类树 | 是 |
| GET  /api/amap/poi/search | POI 搜索 | 否 |
| GET  /api/prompt | 获取系统提示词 | 否 |

---

## WORK_ENV 环境切换

系统支持通过 `WORK_ENV` 环境变量切换配置，避免手动修改 `config.yaml`。

### 可用环境

| 环境值 | 用途 | 数据库配置 |
|--------|------|-----------|
| `home` | 本地开发 | 127.0.0.1:3306 |
| `company` | 公司/服务器 | 10.25.65.177:8306 |

### 使用方式

```bash
export WORK_ENV=home
./server
```

### 支持环境覆盖的字段

| 配置段 | 支持覆盖的字段 |
|--------|---------------|
| `database` | `host`, `port`, `user`, `password`, `name` |
| `redis` | `host`, `port`, `password`, `db` |
| `server` | `port`, `mode` |

---

## LLM_API_KEY 环境变量

与后端其他项目共用同一个 `LLM_API_KEY` 环境变量，优先级高于 `config.yaml`。

```bash
export LLM_API_KEY="sk-your-key-here"
./server
```

启动日志显示 `LLM_API_KEY: applied environment override` 表示已生效。

---

## 本地域名配置（Cookie 隔离）

市民端使用 Token 认证（`Authorization: Bearer`），不依赖 cookie，但建议与管理端使用不同的本地域名以规范部署。

### 配置 hosts

```bash
sudo tee -a /etc/hosts << EOF

# letter-manage 平台本地域名（多平台 cookie 隔离）
127.0.0.1	admin.letter.local
127.0.0.1	citizen.letter.local
EOF
```

### 访问方式

| 平台 | 前端地址 | 后端 API 代理 |
|------|---------|--------------|
| 管理端 | http://admin.letter.local:5173 | → localhost:8080 |
| 市民端 | http://citizen.letter.local:5174 | → localhost:8081 |

---

## 注意事项

1. 必须先启动 Redis，否则服务无法启动
2. Token 存储在 Redis 中，有效期 7 天
3. 验证码存储在 Redis 中，有效期 5 分钟
4. 市民用户表（`citizen_users`）与管理端用户表（`police_users`）相互独立
