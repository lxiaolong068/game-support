using Microsoft.AspNetCore.Mvc;
using Telegram.Bot.Types;
using TelegramFastGptBot.Services;

namespace TelegramFastGptBot.Controllers;

[ApiController]
[Route("[controller]")]
public class TelegramWebhookController : ControllerBase
{
    private readonly TelegramBotService _telegramBotService;
    private readonly ILogger<TelegramWebhookController> _logger;

    public TelegramWebhookController(
        TelegramBotService telegramBotService,
        ILogger<TelegramWebhookController> logger)
    {
        _telegramBotService = telegramBotService;
        _logger = logger;
    }

    [HttpPost("webhook/{token}")]
    public async Task<IActionResult> Post([FromRoute] string token, [FromBody] Update update)
    {
        _logger.LogInformation("收到Webhook请求: {UpdateType}", update.Type);
        await _telegramBotService.HandleUpdateAsync(update);
        return Ok();
    }
}
