using Microsoft.Extensions.Options;
using Telegram.Bot;
using Telegram.Bot.Exceptions;
using Telegram.Bot.Types;
using Telegram.Bot.Types.Enums;
using TelegramFastGptBot.Configuration;

namespace TelegramFastGptBot.Services;

/// <summary>
/// Telegram机器人服务
/// </summary>
public class TelegramBotService
{
    private readonly ITelegramBotClient _botClient;
    private readonly FastGptService _fastGptService;
    private readonly ILogger<TelegramBotService> _logger;
    private readonly AppSettings _appSettings;

    public TelegramBotService(
        ITelegramBotClient botClient,
        FastGptService fastGptService,
        IOptions<AppSettings> appSettings,
        ILogger<TelegramBotService> logger)
    {
        _botClient = botClient;
        _fastGptService = fastGptService;
        _logger = logger;
        _appSettings = appSettings.Value;
    }

    /// <summary>
    /// 设置Webhook
    /// </summary>
    public async Task SetupWebhookAsync()
    {
        var webhookUrl = _appSettings.WebhookUrl + _appSettings.WebhookPath;
        
        _logger.LogInformation("正在设置Webhook: {WebhookUrl}", webhookUrl);
        
        await _botClient.SetWebhookAsync(webhookUrl);
        _logger.LogInformation("Webhook已设置为: {WebhookUrl}", webhookUrl);
    }

    /// <summary>
    /// 处理Telegram更新
    /// </summary>
    public async Task HandleUpdateAsync(Update update)
    {
        try
        {
            // 只处理接收到的文本消息
            if (update.Message is not { } message)
                return;
            if (message.Text is not { } messageText)
                return;
            if (message.Chat.Id == 0)
                return;

            var chatId = message.Chat.Id;
            
            // 忽略命令或空消息
            if (string.IsNullOrEmpty(messageText) || messageText.StartsWith("/"))
                return;

            _logger.LogInformation("收到来自 {ChatId} 的消息: {Message}", chatId, messageText);

            // 发送"正在思考..."消息
            var sentMessage = await _botClient.SendTextMessageAsync(
                chatId: chatId,
                text: "🤔 正在思考中，请稍候...");

            // 调用FastGPT获取回答
            var (answer, success) = await _fastGptService.QueryKnowledgeBaseAsync(
                messageText, 
                chatId.ToString());

            // 准备编辑之前的消息，显示答案
            if (success)
            {
                await _botClient.EditMessageTextAsync(
                    chatId: chatId,
                    messageId: sentMessage.MessageId,
                    text: answer);
                _logger.LogInformation("已发送回答给 {ChatId}: {Answer}", chatId, answer);
            }
            else
            {
                await _botClient.EditMessageTextAsync(
                    chatId: chatId,
                    messageId: sentMessage.MessageId,
                    text: "😥 抱歉，处理您的问题时发生了错误。");
                _logger.LogError("处理消息 {ChatId} 时出错", chatId);
            }
        }
        catch (Exception exception)
        {
            await HandleErrorAsync(exception);
        }
    }

    /// <summary>
    /// 处理错误
    /// </summary>
    private Task HandleErrorAsync(Exception exception)
    {
        var errorMessage = exception switch
        {
            ApiRequestException apiRequestException =>
                $"Telegram API 错误:\n{apiRequestException.ErrorCode}\n{apiRequestException.Message}",
            _ => exception.ToString()
        };

        _logger.LogError("处理消息时发生异常: {ErrorMessage}", errorMessage);
        return Task.CompletedTask;
    }
}
