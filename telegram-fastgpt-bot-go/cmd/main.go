package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/lxiaolong068/game-support/telegram-fastgpt-bot-go/internal/bot"
	"github.com/lxiaolong068/game-support/telegram-fastgpt-bot-go/internal/config"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志器
	if err := config.InitLogger(); err != nil {
		panic(fmt.Sprintf("初始化日志器失败: %v", err))
	}
	defer config.Logger.Sync()

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		config.Logger.Fatal("加载配置失败", zap.Error(err))
	}

	// 验证配置
	missingConfigs := config.AppConfig.Validate()
	if len(missingConfigs) > 0 {
		config.Logger.Fatal("缺少必要的配置", zap.Any("missing_configs", missingConfigs))
	}

	// 初始化Telegram机器人
	if err := bot.InitBot(); err != nil {
		config.Logger.Fatal("初始化Telegram机器人失败", zap.Error(err))
	}

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		// 请求体限制为10MB
		BodyLimit: 10 * 1024 * 1024,
	})

	// 使用中间件
	app.Use(recover.New())
	app.Use(logger.New())

	// 定义根路由
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Telegram FastGPT Bot 服务正在运行")
	})

	// Telegram Webhook路由
	app.Post(config.AppConfig.WebhookPath, func(c *fiber.Ctx) error {
		// 处理Telegram更新
		bot.ProcessWebhookUpdate(c.Body())
		return c.SendStatus(fiber.StatusOK)
	})

	// 健康检查路由
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"bot":    bot.Bot.Self.UserName,
		})
	})

	// 设置Webhook
	if err := bot.SetupWebhook(); err != nil {
		config.Logger.Fatal("设置Webhook失败", zap.Error(err))
	}

	// 获取端口
	port := config.AppConfig.Port

	// 启动服务器
	config.Logger.Info("服务器开始监听", zap.Int("port", port))
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		config.Logger.Error("启动服务器失败", zap.Error(err))
		os.Exit(1)
	}
}
