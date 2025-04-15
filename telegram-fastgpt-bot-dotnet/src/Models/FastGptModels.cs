using System.Text.Json.Serialization;

namespace TelegramFastGptBot.Models;

/// <summary>
/// FastGPT API请求和响应模型类
/// </summary>
public class FastGptMessage
{
    [JsonPropertyName("role")]
    public string Role { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;
}

public class FastGptChatRequest
{
    [JsonPropertyName("chatId")]
    public string ChatId { get; set; } = string.Empty;

    [JsonPropertyName("stream")]
    public bool Stream { get; set; } = false;

    [JsonPropertyName("detail")]
    public bool Detail { get; set; } = false;

    [JsonPropertyName("messages")]
    public List<FastGptMessage> Messages { get; set; } = new List<FastGptMessage>();

    [JsonPropertyName("model")]
    public string Model { get; set; } = string.Empty;
}

public class FastGptChoice
{
    [JsonPropertyName("message")]
    public FastGptChoiceMessage Message { get; set; } = new FastGptChoiceMessage();
}

public class FastGptChoiceMessage
{
    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;
}

public class FastGptChatResponse
{
    [JsonPropertyName("choices")]
    public List<FastGptChoice> Choices { get; set; } = new List<FastGptChoice>();
}
