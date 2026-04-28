package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type WechatMiniprogramConfig struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

type WechatOpenConfig struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

type WechatConfig struct {
	Miniprogram WechatMiniprogramConfig `yaml:"miniprogram"`
	Open        WechatOpenConfig        `yaml:"open"`
}

type Config struct {
	Server       ServerConfig            `yaml:"server"`
	Database     DatabaseConfig          `yaml:"database"`
	Redis        RedisConfig             `yaml:"redis"`
	Wechat       WechatConfig            `yaml:"wechat"`
	LLM          LLMConfig               `yaml:"llm"`
	Map          MapConfig               `yaml:"map"`
	CORS         CORSConfig              `yaml:"cors"`
	Environments map[string]EnvOverride  `yaml:"environments"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Name         string `yaml:"name"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type LLMConfig struct {
	APIKey      string  `yaml:"api_key"`
	APIURL      string  `yaml:"api_url"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

type MapConfig struct {
	AmapKey          string `yaml:"amap_key"`
	AmapGeocodeURL   string `yaml:"amap_geocode_url"`
	AmapRegeocodeURL string `yaml:"amap_regeocode_url"`
	AmapPoiURL       string `yaml:"amap_poi_url"`
	AmapAroundURL    string `yaml:"amap_around_url"`
	AmapInputTipsURL string `yaml:"amap_input_tips_url"`
}

// EnvOverride 环境覆盖配置，只包含需要按环境区分的字段
type EnvOverride struct {
	Database *DatabaseConfig `yaml:"database"`
	Redis    *RedisConfig    `yaml:"redis"`
	Server   *ServerConfig   `yaml:"server"`
}

var AppConfig *Config

// LoadConfig 加载配置：先加载基础配置，再根据 WORK_ENV 合并环境覆盖
func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// 根据 WORK_ENV 合并环境覆盖
	workEnv := os.Getenv("WORK_ENV")
	if workEnv != "" {
		if override, ok := cfg.Environments[workEnv]; ok {
			mergeOverride(cfg, &override)
			log.Printf("WORK_ENV=%s: applied environment overrides", workEnv)
		} else {
			log.Printf("WORK_ENV=%q set but no matching config in environments section", workEnv)
		}
	} else {
		log.Println("WORK_ENV not set, using base config only")
	}

	AppConfig = cfg

	// LLM_API_KEY 环境变量覆盖 API Key（优先级最高）
	if envKey := os.Getenv("LLM_API_KEY"); envKey != "" {
		AppConfig.LLM.APIKey = envKey
		log.Println("LLM_API_KEY: applied environment override")
	}

	return nil
}

// mergeOverride 将环境覆盖配置合并到主配置
func mergeOverride(dst *Config, src *EnvOverride) {
	if src.Database != nil {
		o := src.Database
		if o.Host != "" {
			dst.Database.Host = o.Host
		}
		if o.Port != 0 {
			dst.Database.Port = o.Port
		}
		if o.Name != "" {
			dst.Database.Name = o.Name
		}
		if o.User != "" {
			dst.Database.User = o.User
		}
		if o.Password != "" {
			dst.Database.Password = o.Password
		}
		if o.Charset != "" {
			dst.Database.Charset = o.Charset
		}
		if o.MaxOpenConns != 0 {
			dst.Database.MaxOpenConns = o.MaxOpenConns
		}
		if o.MaxIdleConns != 0 {
			dst.Database.MaxIdleConns = o.MaxIdleConns
		}
	}
	if src.Redis != nil {
		o := src.Redis
		if o.Host != "" {
			dst.Redis.Host = o.Host
		}
		if o.Port != 0 {
			dst.Redis.Port = o.Port
		}
		if o.Password != "" {
			dst.Redis.Password = o.Password
		}
		if o.DB != 0 {
			dst.Redis.DB = o.DB
		}
	}
	if src.Server != nil {
		if src.Server.Port != "" {
			dst.Server.Port = src.Server.Port
		}
		if src.Server.Mode != "" {
			dst.Server.Mode = src.Server.Mode
		}
	}
}

// StartAutoReload 启动定时重载配置，每1分钟重新读取配置文件
func StartAutoReload(path string) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := LoadConfig(path); err != nil {
				log.Printf("auto-reload config failed: %v", err)
			} else {
				log.Println("config auto-reloaded successfully")
			}
		}
	}()
}
