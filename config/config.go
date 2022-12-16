package config

import (
	"github.com/spf13/viper"
	"os"
)

var config *Config

type Config struct {
	ChatGpt ChatGptConfig `json:"chatgpt"`
}

type ChatGptConfig struct {
	Wechat   *string `json:"wechat,omitempty"`
	Token    string  `json:"token,omitempty" json:"token,omitempty"`
	Telegram *string `json:"telegram"`
}

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./local")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	return config
}

func GetWechatEnv() *string {
	return getEnv("wechat")
}

func GetWechatKeywordEnv() *string {
	return getEnv("wechat_keyword")
}

func GetTelegram() *string {
	return getEnv("telegram")
}

func GetTelegramKeyword() *string {
	return getEnv("tg_keyword")
}

func GetTelegramWhitelist() *string {
	return getEnv("tg_whitelist")
}

func GetOpenAiApiKey() *string {
	return getEnv("api_key")
}

func getEnv(key string) *string {
	value := os.Getenv(key)
	if len(value) > 0 {
		return &value
	}
	return nil
}
