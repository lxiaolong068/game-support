package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 保存应用程序的配置
type Config struct {
	// Telegram 配置
	TelegramBotToken string

	// FastGPT API 配置
	FastGPTAPIEndpoint string
	FastGPTAPIKey      string
	FastGPTKBID        string

	// 服务器配置
	Port        int
	WebhookURL  string
	WebhookPath string
}

// 全局配置实例
var AppConfig Config

// 从环境变量加载配置
func LoadConfig() error {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 未找到 .env 文件，将尝试从环境变量加载配置")
	}

	// 加载 Telegram 相关配置
	AppConfig.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if AppConfig.TelegramBotToken == "" {
		return nil
	}

	// 加载 FastGPT 相关配置
	AppConfig.FastGPTAPIEndpoint = os.Getenv("FASTGPT_API_ENDPOINT")
	AppConfig.FastGPTAPIKey = os.Getenv("FASTGPT_API_KEY")
	AppConfig.FastGPTKBID = os.Getenv("FASTGPT_KB_ID")

	// 加载服务器相关配置
	portStr := os.Getenv("PORT")
	if portStr == "" {
		AppConfig.Port = 3000 // 默认端口
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("警告: 无法解析端口号 '%s'，使用默认端口 3000", portStr)
			AppConfig.Port = 3000
		} else {
			AppConfig.Port = port
		}
	}

	AppConfig.WebhookURL = os.Getenv("WEBHOOK_URL")
	AppConfig.WebhookPath = "/webhook/" + AppConfig.TelegramBotToken

	return nil
}

// 检查配置是否有效
func (c *Config) Validate() []string {
	var missingConfigs []string

	if c.TelegramBotToken == "" {
		missingConfigs = append(missingConfigs, "TELEGRAM_BOT_TOKEN")
	}
	if c.FastGPTAPIEndpoint == "" {
		missingConfigs = append(missingConfigs, "FASTGPT_API_ENDPOINT")
	}
	if c.FastGPTAPIKey == "" {
		missingConfigs = append(missingConfigs, "FASTGPT_API_KEY")
	}
	if c.FastGPTKBID == "" {
		missingConfigs = append(missingConfigs, "FASTGPT_KB_ID")
	}
	if c.WebhookURL == "" {
		missingConfigs = append(missingConfigs, "WEBHOOK_URL")
	}

	return missingConfigs
}
