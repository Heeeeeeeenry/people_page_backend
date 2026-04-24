package config

import (
	"os"
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
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Wechat   WechatConfig   `yaml:"wechat"`
	LLM      LLMConfig      `yaml:"llm"`
	Map      MapConfig      `yaml:"map"`
	CORS     CORSConfig     `yaml:"cors"`
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
	AmapKey         string `yaml:"amap_key"`
	AmapGeocodeURL  string `yaml:"amap_geocode_url"`
	AmapRegeocodeURL string `yaml:"amap_regeocode_url"`
	AmapPoiURL      string `yaml:"amap_poi_url"`
	AmapAroundURL   string `yaml:"amap_around_url"`
	AmapInputTipsURL string `yaml:"amap_input_tips_url"`
}

var AppConfig *Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	AppConfig = &Config{}
	return yaml.Unmarshal(data, AppConfig)
}
