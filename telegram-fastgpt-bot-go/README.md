# Telegram 智能客服机器人 (Go版本)

## 项目简介

本项目是一个基于 Go 语言开发的 Telegram 智能客服机器人，使用 FastGPT 知识库提供智能应答服务。机器人接收用户消息后，自动调用 FastGPT 知识库 API，获取答案并回复用户，实现自动化智能问答。

## 功能特性

- 支持 Telegram 聊天机器人自动应答
- 自动调用 FastGPT 知识库进行智能问答
- 使用 Webhook 通信方式，响应速度快
- 基于 Go 语言开发，性能更高，资源占用更低
- 支持容器化部署

## 技术栈

- Go 语言 (建议 1.18+)
- go-telegram-bot-api (Telegram Bot SDK)
- Fiber (高性能 Web 框架)
- godotenv (环境变量管理)

## 目录结构

```
telegram-fastgpt-bot-go/
├── cmd/
│   └── main.go          # 主程序入口
├── internal/
│   ├── bot/
│   │   └── bot.go       # Telegram 机器人逻辑
│   ├── config/
│   │   └── config.go    # 配置加载和管理
│   └── fastgpt/
│       └── fastgpt.go   # FastGPT API 调用逻辑
├── .env                 # 环境变量文件（需手动配置）
├── .gitignore           # Git 忽略文件
├── go.mod               # Go 模块定义
└── README.md            # 项目说明文档
```

## 安装与运行

### 1. 安装 Go

确保您已安装 Go 语言（推荐 1.18 或更高版本）。访问 [Go 官网](https://golang.org/dl/) 下载并安装。

可以通过以下命令检查安装是否成功：
```bash
go version
```

### 2. 克隆或下载项目

```bash
git clone <你的仓库地址>
cd telegram-fastgpt-bot-go
```

### 3. 安装依赖

```bash
go mod download
```

### 4. 配置环境变量

编辑项目根目录下的 `.env` 文件，填写如下信息：

```env
# Telegram
TELEGRAM_BOT_TOKEN=你的_Telegram_Bot_Token

# FastGPT
FASTGPT_API_ENDPOINT=你的_FastGPT_API_地址
FASTGPT_API_KEY=你的_FastGPT_API_Key
FASTGPT_KB_ID=你的_FastGPT_知识库ID

# Server (用于 Webhook)
PORT=3000
WEBHOOK_URL=你的_公网可访问_URL
```

> **注意：**  
> - `WEBHOOK_URL` 必须为 Telegram 能访问到的公网 HTTPS 地址。
> - `.env` 文件已被 .gitignore 忽略，请勿上传敏感信息。

### 5. 构建和运行

#### 直接运行

```bash
go run cmd/main.go
```

#### 构建二进制文件

```bash
go build -o bot cmd/main.go
./bot  # Linux/macOS
bot.exe  # Windows
```

启动后，机器人会自动设置 Webhook 并监听来自 Telegram 的消息。

### 6. 测试

在 Telegram 中找到你的机器人，发送消息进行测试。机器人会自动调用 FastGPT 知识库并回复答案。

## 部署

### 使用 systemd 守护进程（Linux）

创建服务文件 `/etc/systemd/system/telegram-bot.service`：

```ini
[Unit]
Description=Telegram FastGPT Bot
After=network.target

[Service]
ExecStart=/路径/到/你的/bot
WorkingDirectory=/路径/到/你的/项目目录
User=你的用户名
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

### 使用 Docker 部署

1. 创建 Dockerfile：

```Dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bot cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot /app/
COPY .env /app/
EXPOSE 3000
CMD ["./bot"]
```

2. 构建和运行容器：

```bash
docker build -t telegram-fastgpt-bot .
docker run -d -p 3000:3000 --name telegram-bot telegram-fastgpt-bot
```

## 常见问题

### Q: 如何获得 Telegram Bot Token？
A: 在 Telegram 搜索 `BotFather`，按照指引创建机器人即可获得 Token。

### Q: FastGPT 相关参数如何获取？
A: 需注册 FastGPT 服务，创建知识库后在后台获取 API 地址、API Key 和知识库 ID。

### Q: WEBHOOK_URL 如何填写？
A: 必须为公网可访问的 HTTPS 地址。开发阶段可用 ngrok：
```bash
ngrok http 3000
```
将 ngrok 提供的 HTTPS 地址填入 `.env` 文件的 `WEBHOOK_URL` 字段。

### Q: 为什么用 Go 而不是 Node.js？
A: Go 语言相比 Node.js 具有以下优势：
- 极低的内存占用，适合长时间运行的服务
- 高并发处理能力，适合处理大量消息
- 编译为单个二进制文件，部署简单
- 更好的性能和更低的延迟

## 进阶开发建议

- 增加日志记录到文件功能
- 实现会话管理和上下文记忆
- 增加多语言支持
- 增加用户权限控制
- 接入其他聊天平台（如微信、企业微信等）
- 使用 Redis 或其他缓存系统优化性能
- 增加健康检查和监控功能

## 免责声明

本项目仅供学习与个人使用，涉及的 API Key、Token 等敏感信息请妥善保管，避免泄露。
