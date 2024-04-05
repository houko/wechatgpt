package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

var config *Config

type Config struct {
	ChatGpt ChatGptConfig `json:"chatgpt" mapstructure:"chatgpt" yaml:"chatgpt"`
}

type ChatGptConfig struct {
	Token               string `json:"token,omitempty"  mapstructure:"token,omitempty"  yaml:"token,omitempty"`
	OpenAITextModel     string `json:"openai_text_model,omitempty"  mapstructure:"openai_text_model,omitempty"  yaml:"openai_text_model,omitempty"`
	OpenAIImageGenModel string `json:"openai_image_gen_model,omitempty"  mapstructure:"openai_image_gen_model,omitempty"  yaml:"openai_image_gen_model,omitempty"`
	OpenAIVisionModel   string `json:"openai_vision_model,omitempty"  mapstructure:"openai_vision_model,omitempty"  yaml:"openai_vision_model,omitempty"`
	Wechat              string `json:"wechat,omitempty" mapstructure:"wechat,omitempty" yaml:"wechat,omitempty"`
	WechatKeyword       string `json:"wechat_keyword"   mapstructure:"wechat_keyword"   yaml:"wechat_keyword"`
	Telegram            string `json:"telegram"         mapstructure:"telegram"         yaml:"telegram"`
	TgWhitelist         string `json:"tg_whitelist"     mapstructure:"tg_whitelist"     yaml:"tg_whitelist"`
	TgKeyword           string `json:"tg_keyword"       mapstructure:"tg_keyword"       yaml:"tg_keyword"`
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

func GetWechat() string {
	wechat := getEnv("wechat")
	if wechat != "" {
		return wechat
	}

	if config == nil {
		return ""
	}

	if wechat == "" {
		wechat = config.ChatGpt.Wechat
	}
	return wechat
}

func GetWechatKeyword() string {
	keyword := getEnv("wechat_keyword")

	if keyword != "" {
		return keyword
	}

	if config == nil {
		return ""
	}

	if keyword == "" {
		keyword = config.ChatGpt.WechatKeyword
	}
	return keyword
}

func GetTelegram() string {
	tg := getEnv("telegram")
	if tg != "" {
		return tg
	}

	if config == nil {
		return ""
	}

	if tg == "" {
		tg = config.ChatGpt.Telegram
	}
	return tg
}

func GetTelegramKeyword() string {
	tgKeyword := getEnv("tg_keyword")

	if tgKeyword != "" {
		return tgKeyword
	}

	if config == nil {
		return ""
	}

	if tgKeyword == "" {
		tgKeyword = config.ChatGpt.TgKeyword
	}
	return tgKeyword
}

func GetTelegramWhitelist() string {
	tgWhitelist := getEnv("tg_whitelist")

	if tgWhitelist != "" {
		return tgWhitelist
	}

	if config == nil {
		return ""
	}

	if tgWhitelist == "" {
		tgWhitelist = config.ChatGpt.TgWhitelist
	}
	return tgWhitelist
}

func GetOpenAiApiKey() string {
	apiKey := getEnv("api_key")
	if apiKey != "" {
		return apiKey
	}

	if config == nil {
		return ""
	}

	if apiKey == "" {
		apiKey = config.ChatGpt.Token
	}
	return apiKey
}

func GetOpenAiTextModel() (model string) {
	defer func() {
		if model == "" {
			model = "gpt-3.5-turbo"
		}
	}()
	model = getEnv("openai_text_model")
	if model != "" {
		return model
	}

	if config == nil {
		return ""
	}

	if model == "" {
		model = config.ChatGpt.OpenAITextModel
	}
	return model
}

func GetOpenAiImageGenModel() (model string) {
	defer func() {
		if model == "" {
			model = "dall-e-2"
		}
	}()
	model = getEnv("openai_image_gen_model")
	if model != "" {
		return model
	}

	if config == nil {
		return ""
	}

	if model == "" {
		model = config.ChatGpt.OpenAIImageGenModel
	}
	return model
}

func GetOpenAiVisionModel() (model string) {
	defer func() {
		if model == "" {
			model = "gpt-4-1106-vision-preview"
		}
	}()
	model = getEnv("openai_vision_model")
	if model != "" {
		return model
	}

	if config == nil {
		return ""
	}

	if model == "" {
		model = config.ChatGpt.OpenAIVisionModel
	}
	return model
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = os.Getenv(strings.ToUpper(key))
	}

	if len(value) > 0 {
		return value
	}

	if config == nil {
		return ""
	}

	if len(value) > 0 {
		return value
	}

	if config.ChatGpt.WechatKeyword != "" {
		value = config.ChatGpt.WechatKeyword
	}
	return ""
}
