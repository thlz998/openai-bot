package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var config *Config

type Config struct {
	ChatGpt ChatGptConfig `json:"chatgpt"`
}

type ChatGptConfig struct {
	Token         string  `json:"token,omitempty" json:"token,omitempty"`
	Wechat        *string `json:"wechat,omitempty"`
	WechatKeyword *string `json:"wechat_keyword"`
	Telegram      *string `json:"telegram"`
	TgWhitelist   *string `json:"tg_whitelist"`
	TgKeyword     *string `json:"tg_keyword"`
	WeworkToken   *string `json:"wework_token"`
	WeworkAes     *string `json:"wework_aes"`
}

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env文件不存在,读取环境变量", err)
		// os.Exit(0)
	}
	return nil
}

func GetOpenAiApiKey() *string {
	apiKey := getEnv("api_key")

	if apiKey != nil {
		return apiKey
	}

	if config == nil {
		return nil
	}
	if apiKey == nil {
		apiKey = &config.ChatGpt.Token
	}
	return apiKey
}

func getEnv(key string) *string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = os.Getenv(strings.ToUpper(key))
	}

	if len(value) > 0 {
		return &value
	}

	if config == nil {
		return nil
	}

	if len(value) > 0 {
		return &value
	} else if config.ChatGpt.WechatKeyword != nil {
		value = *config.ChatGpt.WechatKeyword
	}
	return nil
}

func GetConfig() *Config {
	return config
}
