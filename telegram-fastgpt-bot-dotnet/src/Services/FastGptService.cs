using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Options;
using TelegramFastGptBot.Configuration;
using TelegramFastGptBot.Models;

namespace TelegramFastGptBot.Services;

/// <summary>
/// FastGPT 知识库查询服务
/// </summary>
public class FastGptService
{
    private readonly HttpClient _httpClient;
    private readonly AppSettings _appSettings;
    private readonly ILogger<FastGptService> _logger;

    public FastGptService(HttpClient httpClient, IOptions<AppSettings> appSettings, ILogger<FastGptService> logger)
    {
        _httpClient = httpClient;
        _appSettings = appSettings.Value;
        _logger = logger;
    }

    /// <summary>
    /// 查询FastGPT知识库
    /// </summary>
    /// <param name="query">用户问题</param>
    /// <param name="chatId">聊天ID（用于FastGPT会话管理）</param>
    /// <returns>FastGPT回答文本，出错则返回错误信息</returns>
    public async Task<(string answer, bool success)> QueryKnowledgeBaseAsync(string query, string chatId)
    {
        // 检查配置是否完整
        if (string.IsNullOrEmpty(_appSettings.FastGptApiEndpoint) || 
            string.IsNullOrEmpty(_appSettings.FastGptApiKey) || 
            string.IsNullOrEmpty(_appSettings.FastGptKbId))
        {
            _logger.LogError("FastGPT配置不完整");
            return ("抱歉，知识库服务当前不可用。", false);
        }

        // 使用默认chatId如果未提供
        if (string.IsNullOrEmpty(chatId))
        {
            chatId = "default_user";
        }

        try
        {
            // 准备请求体
            var request = new FastGptChatRequest
            {
                ChatId = chatId,
                Stream = false, // 不使用流式响应
                Detail = false, // 不需要详细处理信息
                Model = _appSettings.FastGptKbId, // 使用知识库ID作为模型标识符
                Messages = new List<FastGptMessage>
                {
                    new FastGptMessage
                    {
                        Role = "user",
                        Content = query
                    }
                }
            };

            var requestJson = JsonSerializer.Serialize(request);
            var content = new StringContent(requestJson, Encoding.UTF8, "application/json");
            
            // 设置请求头
            _httpClient.DefaultRequestHeaders.Authorization = 
                new AuthenticationHeaderValue("Bearer", _appSettings.FastGptApiKey);

            // 发送请求
            var response = await _httpClient.PostAsync(
                $"{_appSettings.FastGptApiEndpoint}/api/v1/chat/completions", 
                content);

            // 检查HTTP状态码
            if (!response.IsSuccessStatusCode)
            {
                _logger.LogError("FastGPT API返回非200状态码: {StatusCode}", response.StatusCode);
                return ("抱歉，知识库返回了错误。", false);
            }

            // 读取响应内容
            var responseContent = await response.Content.ReadAsStringAsync();
            
            // 解析JSON响应
            var chatResponse = JsonSerializer.Deserialize<FastGptChatResponse>(responseContent);
            
            // 检查是否有有效答案
            if (chatResponse?.Choices == null || chatResponse.Choices.Count == 0)
            {
                _logger.LogError("FastGPT API返回了空的choices数组, 响应: {Response}", responseContent);
                return ("抱歉，无法从知识库获取有效的回答。", false);
            }

            // 返回处理后的答案
            var answer = chatResponse.Choices[0].Message.Content;
            return (answer, true);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "查询FastGPT知识库时出错");
            return ("抱歉，查询知识库时出错。", false);
        }
    }
}
