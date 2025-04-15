namespace TelegramFastGptBot.Configuration;

/// <summary>
/// 应用程序配置类
/// </summary>
public class AppSettings
{
    // Telegram 配置
    public string TelegramBotToken { get; set; } = string.Empty;

    // FastGPT API 配置
    public string FastGptApiEndpoint { get; set; } = string.Empty;
    public string FastGptApiKey { get; set; } = string.Empty;
    public string FastGptKbId { get; set; } = string.Empty;

    // 服务器配置
    public int Port { get; set; } = 5000;
    public string WebhookUrl { get; set; } = string.Empty;
    public string WebhookPath => $"/webhook/{TelegramBotToken}";

    /// <summary>
    /// 检查配置是否有效
    /// </summary>
    /// <returns>缺失的配置项列表</returns>
    public List<string> Validate()
    {
        var missingConfigs = new List<string>();

        if (string.IsNullOrEmpty(TelegramBotToken))
            missingConfigs.Add("TELEGRAM_BOT_TOKEN");
        
        if (string.IsNullOrEmpty(FastGptApiEndpoint))
            missingConfigs.Add("FASTGPT_API_ENDPOINT");
        
        if (string.IsNullOrEmpty(FastGptApiKey))
            missingConfigs.Add("FASTGPT_API_KEY");
        
        if (string.IsNullOrEmpty(FastGptKbId))
            missingConfigs.Add("FASTGPT_KB_ID");
        
        if (string.IsNullOrEmpty(WebhookUrl))
            missingConfigs.Add("WEBHOOK_URL");

        return missingConfigs;
    }
}
