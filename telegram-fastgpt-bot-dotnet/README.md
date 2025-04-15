# Telegram FastGPT 机器人（.NET 9 版本）

## 项目简介

这是一个基于 .NET 8 开发的 Telegram 机器人，可以连接到 FastGPT 知识库 API，为用户提供智能问答服务。机器人通过 Telegram 的 Webhook 接收消息，然后调用 FastGPT API 获取回答，并将结果返回给用户。

## 技术栈

- .NET 9
- ASP.NET Core Web API
- Telegram.Bot (19.0.0)
- DotNetEnv (2.5.0)
- System.Text.Json

## 功能特点

- 通过 Webhook 接收 Telegram 消息
- 调用 FastGPT 知识库 API 进行智能问答
- 配置简单，使用 .env 文件管理敏感信息
- 支持部署为单一可执行文件
- 适合低资源服务器和高并发场景
- 支持 Docker 容器化部署

## 快速开始

### 前提条件

- .NET 9 SDK
- Telegram 机器人 Token（通过 BotFather 创建）
- FastGPT API 访问凭证
- 可公网访问的服务器（用于接收 Webhook）

### 配置步骤

1. 克隆本仓库
2. 复制 `.env.example` 为 `.env` 并填入你的配置信息：

```bash
# 复制配置文件模板
cp src/.env.example src/.env

# 编辑配置文件
nano src/.env
```

3. 修改 `.env` 文件中的以下配置：

```
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
FASTGPT_API_ENDPOINT=https://api.fastgpt.com
FASTGPT_API_KEY=your_fastgpt_api_key_here
FASTGPT_KB_ID=your_fastgpt_knowledge_base_id_here
PORT=5000
WEBHOOK_URL=https://your-domain.com
```

### 运行项目

```bash
# 进入项目目录
cd src

# 恢复依赖
dotnet restore

# 构建项目
dotnet build

# 运行项目
dotnet run
```

### 发布项目

```bash
# 发布为单一可执行文件
dotnet publish -c Release -r linux-x64 --self-contained -p:PublishSingleFile=true -o ./publish
```

## 部署方式

### 使用 systemd（Linux）

1. 创建 systemd 服务文件

```bash
sudo nano /etc/systemd/system/telegram-fastgpt-bot.service
```

2. 添加以下内容

```
[Unit]
Description=Telegram FastGPT Bot Service
After=network.target

[Service]
WorkingDirectory=/path/to/your/app
ExecStart=/path/to/your/app/TelegramFastGptBot
Restart=always
RestartSec=10
SyslogIdentifier=telegram-fastgpt-bot
User=your-user
Environment=ASPNETCORE_ENVIRONMENT=Production

[Install]
WantedBy=multi-user.target
```

3. 启用并启动服务

```bash
sudo systemctl enable telegram-fastgpt-bot
sudo systemctl start telegram-fastgpt-bot
```

### 使用 Docker

1. 创建 Dockerfile

```dockerfile
FROM mcr.microsoft.com/dotnet/aspnet:9.0 AS base
WORKDIR /app
EXPOSE 5000

FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src
COPY ["TelegramFastGptBot.csproj", "./"]
RUN dotnet restore "TelegramFastGptBot.csproj"
COPY . .
RUN dotnet build "TelegramFastGptBot.csproj" -c Release -o /app/build

FROM build AS publish
RUN dotnet publish "TelegramFastGptBot.csproj" -c Release -o /app/publish

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
COPY .env .
ENTRYPOINT ["dotnet", "TelegramFastGptBot.dll"]
```

2. 构建并运行容器

```bash
docker build -t telegram-fastgpt-bot .
docker run -d -p 5000:5000 --name telegram-bot telegram-fastgpt-bot
```

## 项目结构

```
src/
├── Configuration/
│   └── AppSettings.cs          # 应用程序配置类
├── Controllers/
│   └── TelegramWebhookController.cs  # Webhook控制器
├── Models/
│   └── FastGptModels.cs        # FastGPT API请求和响应模型
├── Services/
│   ├── FastGptService.cs       # FastGPT知识库服务
│   └── TelegramBotService.cs   # Telegram机器人服务
├── .env.example                # 环境变量示例文件
├── Program.cs                  # 应用程序入口点
└── TelegramFastGptBot.csproj   # 项目文件
```

## 流程说明

1. 用户向 Telegram 机器人发送消息
2. Telegram 通过 Webhook 将消息转发到我们的服务
3. 控制器接收消息后，交由 TelegramBotService 处理
4. TelegramBotService 发送"正在思考..."消息给用户
5. FastGptService 将用户问题发送给 FastGPT 知识库 API
6. 获取 FastGPT 回答后，编辑之前的"正在思考..."消息为最终答案
7. 用户收到 FastGPT 的回答

## 注意事项

- 确保 `.env` 文件中的所有配置都已正确填写
- 确保 `WEBHOOK_URL` 可以通过公网访问，Telegram 需要能够向此 URL 发送 Webhook 请求
- 如需使用 HTTPS，建议使用 Nginx 等反向代理服务器处理 SSL 证书

## 常见问题

1. **Q: 为什么我的 Webhook 设置失败？**  
   A: 请确保你的服务器可以通过公网访问，且 `WEBHOOK_URL` 配置正确。如果使用了反向代理，请确保正确转发请求到应用程序。

2. **Q: 如何检查机器人是否正常工作？**  
   A: 查看应用程序日志，确保 Webhook 设置成功。然后向你的 Telegram 机器人发送消息测试。

3. **Q: 如何更新 FastGPT 知识库？**  
   A: 这需要在 FastGPT 平台上进行操作，本机器人仅负责调用 API。

## 许可证

MIT
