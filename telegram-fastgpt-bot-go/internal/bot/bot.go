package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lxiaolong068/game-support/telegram-fastgpt-bot-go/internal/config"
	"github.com/lxiaolong068/game-support/telegram-fastgpt-bot-go/internal/fastgpt"
	"go.uber.org/zap"
)

var Bot *tgbotapi.BotAPI

// 初始化Telegram机器人
func InitBot() error {
	var err error
	
	// 创建机器人实例
	Bot, err = tgbotapi.NewBotAPI(config.AppConfig.TelegramBotToken)
	if err != nil {
		config.Logger.Error("创建Telegram机器人失败", zap.Error(err))
		return fmt.Errorf("创建Telegram机器人失败: %w", err)
	}

	// 设置调试模式
	Bot.Debug = false
	config.Logger.Info("已授权账号", zap.String("username", Bot.Self.UserName))
	
	return nil
}

// 设置Webhook
func SetupWebhook() error {
	webhookURL := config.AppConfig.WebhookURL + config.AppConfig.WebhookPath
	
	// 配置Webhook
	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		config.Logger.Error("创建Webhook配置失败", zap.Error(err))
		return fmt.Errorf("创建Webhook配置失败: %w", err)
	}
	
	// 设置Webhook
	_, err = Bot.Request(webhook)
	if err != nil {
		config.Logger.Error("设置Webhook失败", zap.Error(err))
		return fmt.Errorf("设置Webhook失败: %w", err)
	}
	
	config.Logger.Info("Webhook已设置", zap.String("webhook_url", webhookURL))
	return nil
}

// 处理Telegram消息
func HandleUpdate(update tgbotapi.Update) {
	// 只处理接收到的消息
	if update.Message == nil {
		return
	}
	
	message := update.Message
	chatID := message.Chat.ID
	text := message.Text

	// 忽略命令或空消息
	if text == "" || strings.HasPrefix(text, "/") {
		return
	}

	config.Logger.Info("收到消息", zap.Int64("chat_id", chatID), zap.String("text", text))

	// 发送"正在思考..."消息
	thinkingMsg := tgbotapi.NewMessage(chatID, "🤔 正在思考中，请稍候...")
	sentMsg, err := Bot.Send(thinkingMsg)
	if err != nil {
		config.Logger.Error("发送思考消息失败", zap.Error(err))
		return
	}
	
	// 调用FastGPT获取回答
	answer, err := fastgpt.QueryKnowledgeBase(text, fmt.Sprintf("%d", chatID))
	
	// 准备编辑之前的消息，显示答案
	var editMsg tgbotapi.EditMessageTextConfig
	
	if err != nil {
		config.Logger.Error("处理消息时出错", zap.Int64("chat_id", chatID), zap.Error(err))
		editMsg = tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, "😥 抱歉，处理您的问题时发生了错误。")
	} else {
		editMsg = tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, answer)
		config.Logger.Info("已发送回答", zap.Int64("chat_id", chatID), zap.String("answer", answer))
	}
	
	// 编辑之前的消息
	_, err = Bot.Send(editMsg)
	if err != nil {
		config.Logger.Warn("编辑消息失败，尝试发送新消息", zap.Error(err))
		// 编辑失败时，直接发送新消息，确保用户能收到回复
		fallbackMsg := tgbotapi.NewMessage(chatID, editMsg.Text)
		_, sendErr := Bot.Send(fallbackMsg)
		if sendErr != nil {
			config.Logger.Error("补发新消息也失败", zap.Error(sendErr))
		}
	}
}

// 处理Webhook更新
// 支持高并发：每个 update 启动一个 goroutine 处理消息。
// go-telegram-bot-api v5 的 BotAPI.Send 方法是并发安全的。
func ProcessWebhookUpdate(updateBytes []byte) {
	var update tgbotapi.Update
	err := json.Unmarshal(updateBytes, &update)
	if err != nil {
		config.Logger.Error("解析更新失败", zap.Error(err))
		return
	}
	// 并发处理每个 update，提升高并发下的响应能力
	go HandleUpdate(update)
}
