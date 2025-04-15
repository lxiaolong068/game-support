using System.Net.Http.Headers;
using DotNetEnv;
using Microsoft.Extensions.Options;
using Telegram.Bot;
using TelegramFastGptBot.Configuration;
using TelegramFastGptBot.Services;

var builder = WebApplication.CreateBuilder(args);

// 加载环境变量
Env.Load();

// 配置AppSettings
builder.Services.Configure<AppSettings>(settings => {
    settings.TelegramBotToken = Environment.GetEnvironmentVariable("TELEGRAM_BOT_TOKEN") ?? string.Empty;
    settings.FastGptApiEndpoint = Environment.GetEnvironmentVariable("FASTGPT_API_ENDPOINT") ?? string.Empty;
    settings.FastGptApiKey = Environment.GetEnvironmentVariable("FASTGPT_API_KEY") ?? string.Empty;
    settings.FastGptKbId = Environment.GetEnvironmentVariable("FASTGPT_KB_ID") ?? string.Empty;
    
    var portStr = Environment.GetEnvironmentVariable("PORT");
    if (!string.IsNullOrEmpty(portStr) && int.TryParse(portStr, out int port))
    {
        settings.Port = port;
    }
    
    settings.WebhookUrl = Environment.GetEnvironmentVariable("WEBHOOK_URL") ?? string.Empty;
});

// 注册HttpClient
builder.Services.AddHttpClient();

// 注册Telegram机器人客户端
builder.Services.AddSingleton<ITelegramBotClient>(sp => {
    var options = sp.GetRequiredService<IOptions<AppSettings>>();
    var botToken = options.Value.TelegramBotToken;
    
    if (string.IsNullOrEmpty(botToken))
    {
        throw new InvalidOperationException("未配置Telegram机器人Token");
    }
    
    return new TelegramBotClient(botToken);
});

// 添加FastGPT服务
builder.Services.AddScoped<FastGptService>();

// 添加Telegram机器人服务
builder.Services.AddScoped<TelegramBotService>();

// 添加控制器
builder.Services.AddControllers().AddJsonOptions(options => {
    options.JsonSerializerOptions.PropertyNamingPolicy = null;
});

// 添加OpenAPI支持
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

// 配置HTTP请求管道
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseAuthorization();
app.MapControllers();

// 启动时设置Webhook
var scope = app.Services.CreateScope();
var telegramBotService = scope.ServiceProvider.GetRequiredService<TelegramBotService>();
var appSettings = scope.ServiceProvider.GetRequiredService<IOptions<AppSettings>>().Value;
var logger = scope.ServiceProvider.GetRequiredService<ILogger<Program>>();

// 检查配置是否完整
var missingConfigs = appSettings.Validate();
if (missingConfigs.Count > 0)
{
    logger.LogError("配置不完整，缺少以下环境变量: {MissingConfigs}", string.Join(", ", missingConfigs));
}
else
{
    try
    {
        // 设置Webhook
        await telegramBotService.SetupWebhookAsync();
        logger.LogInformation("已成功设置Telegram机器人Webhook");
    }
    catch (Exception ex)
    {
        logger.LogError(ex, "设置Webhook时出错");
    }
}

// 启动应用
app.Run();
