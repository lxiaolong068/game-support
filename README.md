# Telegram 智能客服机器人 (Go版本)

## 项目简介

本项目是一个基于 Go 语言开发的 Telegram 智能客服机器人，使用 FastGPT 知识库提供智能应答服务。机器人接收用户消息后，自动调用 FastGPT 知识库 API，获取答案并回复用户，实现自动化智能问答。

## 功能特性

- 支持 Telegram 聊天机器人自动应答
- 自动调用 FastGPT 知识库进行智能问答
- 使用 Webhook 通信方式，响应速度快
- 基于 Go 语言开发，性能更高，资源占用更低
- 支持容器化部署
- 内置速率限制（基于令牌桶算法，防止单用户恶意刷消息，线程安全、高性能）

## 技术栈

- Go 语言 (建议 1.18+)
- go-telegram-bot-api (Telegram Bot SDK)
- Fiber (高性能 Web 框架)
- godotenv (环境变量管理)

## 接口说明

本项目主要涉及以下两个外部接口交互：

### 1. Telegram Webhook 接收接口

当用户向 Telegram Bot 发送消息时，Telegram 服务器会向本项目配置的 Webhook URL 发送 POST 请求。

- **URL:** 由 `.env` 文件中的 `WEBHOOK_URL` 变量定义。
- **端口:** 由 `.env` 文件中的 `PORT` 变量定义（默认为 3000）。
- **HTTP 方法:** `POST`
- **请求体 (Payload):** Telegram 服务器发送的 `Update` 对象 (JSON 格式)，包含了用户消息等信息。具体结构请参考 [Telegram Bot API 文档](https://core.telegram.org/bots/api#update)。
- **处理逻辑:** 应用程序接收到 `Update` 后，提取用户消息，并调用 FastGPT API 获取回复。

### 2. FastGPT 知识库查询接口

本项目调用 FastGPT 提供的 API 来查询知识库并获取智能回复。

- **URL:** 由 `.env` 文件中的 `FASTGPT_API_ENDPOINT` 变量定义。
- **HTTP 方法:** `POST` (通常用于查询)
- **请求头 (Headers):**
    - `Authorization: Bearer <FASTGPT_API_KEY>` (使用 `.env` 文件中的 `FASTGPT_API_KEY`)
    - `Content-Type: application/json`
- **请求体 (Payload):** JSON 格式，包含需要查询的问题等信息。通常结构如下（具体请参考您使用的 FastGPT 版本文档）：
  ```json
  {
    "kbId": "你的_FastGPT_知识库ID", // 从 .env 读取 FASTGPT_KB_ID
    "prompt": "用户发送的问题文本",
    // 可能包含其他参数，如 stream, detail 等
  }
  ```
- **响应体 (Response):** FastGPT 返回的包含答案的 JSON 数据。
- **处理逻辑:** 将从 FastGPT 获取到的答案格式化后，通过 Telegram Bot API 回复给用户。

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
```

## 编译说明

本项目为原生 Go 代码，无需额外依赖，支持跨平台编译。推荐使用 Go 1.18 及以上版本。

- 编译 Linux/macOS 二进制：
  ```bash
  go build -o bot cmd/main.go
  ```
- 编译 Windows 可执行文件：
  ```bash
  go build -o bot.exe cmd/main.go
  ```
- 交叉编译示例（如在 macOS 下编译 Linux 版）：
  ```bash
  GOOS=linux GOARCH=amd64 go build -o telegram-fastgpt-bot-go cmd/main.go
  ```

编译成功后，直接运行生成的二进制文件即可。
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

- 日志系统升级为结构化日志（zap），并优化错误处理，便于后期排查和监控
- 实现会话管理和上下文记忆
- 增加多语言支持
- 增加用户权限控制
- 已实现速率限制：每个用户每分钟最多5次请求，采用 Go 官方 `golang.org/x/time/rate` 令牌桶算法，线程安全且无锁高效。- 接入其他聊天平台（如微信、企业微信等）
- 使用 Redis 或其他缓存系统优化性能
- 增加健康检查和监控功能

## 免责声明

本项目仅供学习与个人使用，涉及的 API Key、Token 等敏感信息请妥善保管，避免泄露。
