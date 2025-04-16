package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config 保存应用程序的配置
type Config struct {
	// Telegram 配置
	TelegramBotToken string

	// FastGPT API 配置
	FastGPTAPIEndpoint string
	FastGPTAPIKey      string
	FastGPTKBID        string
	FastGPTLimit       int
	FastGPTSimilarity  float64
	FastGPTSearchMode  string
	FastGPTUsingReRank bool
	FastGPTDatasetSearchExtensionModel string

	// 缓存配置
	EnableCache      bool
	CacheExpiration  int    // 单位秒
	CacheMaxEntries  int

	// 服务器配置
	Port        int
	WebhookURL  string
	WebhookPath string
}

// 全局配置实例
var AppConfig Config

// 全局日志实例
var Logger *zap.Logger

// 初始化日志器
func InitLogger() error {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		return fmt.Errorf("初始化日志器失败: %w", err)
	}
	return nil
}

// 从环境变量加载配置
func LoadConfig() error {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		Logger.Warn("未找到 .env 文件，将尝试从环境变量加载配置")
	}

	// 加载 Telegram 相关配置
	AppConfig.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if AppConfig.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN 不能为空")
	}

	// 加载 FastGPT 相关配置
	AppConfig.FastGPTAPIEndpoint = os.Getenv("FASTGPT_API_ENDPOINT")
	AppConfig.FastGPTAPIKey = os.Getenv("FASTGPT_API_KEY")
	AppConfig.FastGPTKBID = os.Getenv("FASTGPT_KB_ID")
	// limit
	if v := os.Getenv("FASTGPT_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			AppConfig.FastGPTLimit = n
		} else {
			AppConfig.FastGPTLimit = 5000
		}
	} else {
		AppConfig.FastGPTLimit = 5000
	}
	// similarity
	if v := os.Getenv("FASTGPT_SIMILARITY"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			AppConfig.FastGPTSimilarity = f
		} else {
			AppConfig.FastGPTSimilarity = 0
		}
	} else {
		AppConfig.FastGPTSimilarity = 0
	}
	// search mode
	if v := os.Getenv("FASTGPT_SEARCH_MODE"); v != "" {
		AppConfig.FastGPTSearchMode = v
	} else {
		AppConfig.FastGPTSearchMode = "embedding"
	}
	// usingReRank
	if v := os.Getenv("FASTGPT_USING_RERANK"); v != "" {
		AppConfig.FastGPTUsingReRank = v == "true" || v == "1"
	} else {
		AppConfig.FastGPTUsingReRank = false
	}
	// datasetSearchExtensionModel
	if v := os.Getenv("FASTGPT_DATASET_SEARCH_EXTENSION_MODEL"); v != "" {
		AppConfig.FastGPTDatasetSearchExtensionModel = v
	} else {
		AppConfig.FastGPTDatasetSearchExtensionModel = "deepseek-v3"
	}

	// 缓存配置
	if v := os.Getenv("ENABLE_CACHE"); v != "" {
		AppConfig.EnableCache = v == "true" || v == "1"
	} else {
		AppConfig.EnableCache = true
	}
	if v := os.Getenv("CACHE_EXPIRATION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			AppConfig.CacheExpiration = n
		} else {
			AppConfig.CacheExpiration = 120
		}
	} else {
		AppConfig.CacheExpiration = 120
	}
	if v := os.Getenv("CACHE_MAX_ENTRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			AppConfig.CacheMaxEntries = n
		} else {
			AppConfig.CacheMaxEntries = 1000
		}
	} else {
		AppConfig.CacheMaxEntries = 1000
	}

	// 加载服务器相关配置
	portStr := os.Getenv("PORT")
	if portStr == "" {
		AppConfig.Port = 3000 // 默认端口
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			Logger.Warn("无法解析端口号，使用默认端口", zap.String("port_str", portStr))
			AppConfig.Port = 3000
		} else {
			AppConfig.Port = port
		}
	}

	AppConfig.WebhookURL = os.Getenv("WEBHOOK_URL")
	// 使用 Token 的 SHA256 哈希作为 WebhookPath，避免直接暴露 Token
	hash := sha256.Sum256([]byte(AppConfig.TelegramBotToken))
	AppConfig.WebhookPath = "/webhook/" + hex.EncodeToString(hash[:8]) // 取前8字节，足够唯一且不易反推

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
