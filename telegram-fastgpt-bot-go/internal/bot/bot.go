package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/telegram-fastgpt-bot-go/internal/config"
	"github.com/yourusername/telegram-fastgpt-bot-go/internal/fastgpt"
)

var Bot *tgbotapi.BotAPI

// 初始化Telegram机器人
func InitBot() error {
	var err error
	
	// 创建机器人实例
	Bot, err = tgbotapi.NewBotAPI(config.AppConfig.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("创建Telegram机器人失败: %w", err)
	}

	// 设置调试模式
	Bot.Debug = false
	log.Printf("已授权账号 %s", Bot.Self.UserName)
	
	return nil
}

// 设置Webhook
func SetupWebhook() error {
	webhookURL := config.AppConfig.WebhookURL + config.AppConfig.WebhookPath
	
	// 配置Webhook
	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("创建Webhook配置失败: %w", err)
	}
	
	// 设置Webhook
	_, err = Bot.Request(webhook)
	if err != nil {
		return fmt.Errorf("设置Webhook失败: %w", err)
	}
	
	log.Printf("Webhook已设置为: %s", webhookURL)
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

	log.Printf("收到来自 %d 的消息: %s", chatID, text)

	// 发送"正在思考..."消息
	thinkingMsg := tgbotapi.NewMessage(chatID, "🤔 正在思考中，请稍候...")
	sentMsg, err := Bot.Send(thinkingMsg)
	if err != nil {
		log.Printf("发送思考消息失败: %v", err)
		return
	}
	
	// 调用FastGPT获取回答
	answer, err := fastgpt.QueryKnowledgeBase(text, fmt.Sprintf("%d", chatID))
	
	// 准备编辑之前的消息，显示答案
	var editMsg tgbotapi.EditMessageTextConfig
	
	if err != nil {
		log.Printf("处理消息 %d 时出错: %v", chatID, err)
		editMsg = tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, "😥 抱歉，处理您的问题时发生了错误。")
	} else {
		editMsg = tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, answer)
		log.Printf("已发送回答给 %d: %s", chatID, answer)
	}
	
	// 编辑之前的消息
	_, err = Bot.Send(editMsg)
	if err != nil {
		log.Printf("编辑消息失败: %v", err)
	}
}

// 处理Webhook更新
func ProcessWebhookUpdate(updateBytes []byte) {
	update, err := tgbotapi.NewUpdateFromJSON(updateBytes)
	if err != nil {
		log.Printf("解析更新失败: %v", err)
		return
	}
	
	HandleUpdate(update)
}
