# Telegram 智能客服机器人（FastGPT 知识库驱动）

## 项目简介

本项目是一个基于 Node.js + Fastify 框架开发的 Telegram 智能客服机器人。机器人接收用户消息后，自动调用 FastGPT 知识库 API，获取答案并回复用户，实现自动化智能问答。

## 功能特性

- 支持 Telegram 聊天机器人自动应答
- 自动调用 FastGPT 知识库进行智能问答
- 支持 Webhook 通信方式，响应速度快
- 代码结构清晰，易于扩展和二次开发

## 技术栈

- Node.js
- Fastify
- node-telegram-bot-api
- axios
- dotenv
- pnpm（包管理器）

## 目录结构

```
telegram-fastgpt-bot/
├── src/
│   ├── bot.js         # Telegram 机器人逻辑
│   ├── server.js      # Fastify 服务器设置
│   └── fastgpt.js     # FastGPT API 调用逻辑
├── .env               # 环境变量文件（需手动配置）
├── .gitignore         # Git 忽略文件
├── package.json       # 项目配置
└── README.md          # 项目说明文档
```

## 安装与运行

### 1. 安装 pnpm

如未安装 pnpm，请先全局安装：
```bash
npm install -g pnpm
```

### 2. 克隆或复制项目

```bash
git clone <你的仓库地址>
cd telegram-fastgpt-bot
```

### 3. 安装依赖

```bash
pnpm install
```

### 4. 配置环境变量

在项目根目录下编辑 `.env` 文件，填写如下内容（已给出模板，需替换为你的实际信息）：

```env
# Telegram
TELEGRAM_BOT_TOKEN=你的_Telegram_Bot_Token

# FastGPT
FASTGPT_API_ENDPOINT=你的_FastGPT_API_Endpoint
FASTGPT_API_KEY=你的_FastGPT_API_Key
FASTGPT_KB_ID=你的_FastGPT_知识库_ID

# Server (用于 Webhook)
PORT=3000
WEBHOOK_URL=你的_公网可访问_URL # 例如: https://your-domain.com 或 ngrok 临时 URL
```

> **注意：**  
> - `WEBHOOK_URL` 必须为 Telegram 能访问到的公网 https 地址。开发时推荐使用 [ngrok](https://ngrok.com/) 暴露本地端口。
> - `.env` 文件已被 `.gitignore` 忽略，请勿上传敏感信息。

### 5. 启动服务

```bash
node src/server.js
```

启动后，机器人会自动设置 Webhook 并监听来自 Telegram 的消息。

### 6. 测试

在 Telegram 中找到你的机器人，发送消息进行测试。机器人会自动调用 FastGPT 知识库并回复答案。

## 常见问题

### Q: 如何获得 Telegram Bot Token？
A: 在 Telegram 搜索 `BotFather`，按照指引创建机器人即可获得 Token。

### Q: FastGPT 相关参数如何获取？
A: 需注册 FastGPT 服务，创建知识库后在后台获取 API 地址、API Key 和知识库 ID。

### Q: WEBHOOK_URL 如何填写？
A: 必须为公网可访问的 https 地址。开发阶段可用 ngrok：
```bash
ngrok http 3000
```
将 ngrok 提供的 https 地址填入 `.env` 文件。

### Q: 如何后台守护运行？
A: 推荐使用 pm2：
```bash
pnpm add pm2 -g
pm2 start src/server.js --name telegram-fastgpt-bot
```

## 进阶开发建议

- 支持多知识库切换
- 增加上下文对话能力
- 支持 Markdown 或富文本消息格式
- 增加用户权限与日志记录
- 对接更多平台（如微信、企业微信等）

## 免责声明

本项目仅供学习与个人使用，涉及的 API Key、Token 等敏感信息请妥善保管，避免泄露。

---

如有问题或建议，欢迎 issue 或联系作者。
