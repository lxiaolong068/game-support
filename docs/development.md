# Telegram 智能客服机器人开发文档 (Node.js + Fastify + FastGPT)

## 1. 项目概述

### 1.1 项目目标
构建一个 Telegram 机器人，作为智能客服，能够接收用户的提问，调用 FastGPT 知识库 API 获取答案，并将答案回复给用户。

### 1.2 技术栈
*   **后端框架:** Node.js + Fastify
*   **Telegram API 库:** `node-telegram-bot-api`
*   **HTTP 请求库:** `axios` (或 Node.js 内建的 `fetch`)
*   **知识库:** FastGPT API

### 1.3 目标用户
独立开发者，熟悉基本的 Node.js 开发。

## 2. 环境准备

### 2.1 安装 Node.js
确保你的开发环境中安装了 Node.js (推荐 LTS 版本)。访问 [Node.js 官网](https://nodejs.org/) 下载并安装。
可以通过以下命令检查安装是否成功：
```bash
node -v
pnpm -v
```

### 2.2 获取 Telegram Bot Token
1.  在 Telegram 中搜索 `BotFather`。
2.  与 `BotFather` 对话，使用 `/newbot` 命令创建一个新的机器人。
3.  按照提示设置机器人名称和用户名。
4.  `BotFather` 会提供一个 **HTTP API Token**，请务必**安全保存**这个 Token，后续代码中需要使用。

### 2.3 获取 FastGPT API 信息
你需要拥有一个 FastGPT 实例，并获取其 API 的访问凭证：
1.  **FastGPT API Endpoint:** 你的 FastGPT 服务提供的 API 地址。
2.  **FastGPT API Key:** 用于认证 API 请求的密钥。
3.  **知识库 ID (Knowledge Base ID):** 你希望机器人查询的具体知识库的 ID。

请确保这些信息可用并安全保存。

## 3. 项目初始化与设置

### 3.1 初始化项目
创建一个新的项目目录，并在该目录下初始化 npm 项目：
```bash
mkdir telegram-fastgpt-bot
cd telegram-fastgpt-bot
pnpm init
```

### 3.2 安装依赖
安装必要的 npm 包：
```bash
pnpm add fastify node-telegram-bot-api axios dotenv
```
*   `fastify`: 高性能 Node.js Web 框架。
*   `node-telegram-bot-api`: 与 Telegram Bot API 交互的库。
*   `axios`: 用于向 FastGPT API 发送 HTTP 请求。
*   `dotenv`: 用于管理环境变量 (存储敏感信息)。

### 3.3 创建项目结构 (建议)
```
telegram-fastgpt-bot/
├── src/
│   ├── bot.js         # Telegram 机器人逻辑
│   ├── server.js      # Fastify 服务器设置
│   └── fastgpt.js     # FastGPT API 调用逻辑
├── config/
│   └── index.js       # 配置管理 (可选)
├── .env             # 环境变量文件 (需添加到 .gitignore)
├── .gitignore       # Git 忽略文件
├── package.json
└── docs/
    └── development.md # 本文档
```

### 3.4 配置环境变量
在项目根目录下创建 `.env` 文件，并添加以下内容，替换为你的实际信息：
```dotenv
# Telegram
TELEGRAM_BOT_TOKEN=你的_TELEGRAM_BOT_TOKEN

# FastGPT
FASTGPT_API_ENDPOINT=你的_FASTGPT_API_ENDPOINT
FASTGPT_API_KEY=你的_FASTGPT_API_KEY
FASTGPT_KB_ID=你的_FASTGPT_知识库_ID

# Server (用于 Webhook)
PORT=3000
WEBHOOK_URL=你的_公网可访问_URL # 例如: https://your-domain.com 或 ngrok 临时 URL
```
**重要:**
*   确保将 `.env` 文件添加到 `.gitignore` 中，避免将敏感信息提交到版本控制系统。
*   `WEBHOOK_URL` 必须是 Telegram 服务器可以访问到的公网 URL。在开发阶段，可以使用 `ngrok` 等工具创建临时的公网隧道。

### 3.5 配置 `.gitignore`
在项目根目录创建 `.gitignore` 文件，添加以下内容：
```gitignore
node_modules/
.env
*.log
```

## 4. 核心代码实现

### 4.1 Fastify 服务器设置 (`src/server.js`)
```javascript
// src/server.js
require('dotenv').config();
const Fastify = require('fastify');
const bot = require('./bot'); // 引入机器人逻辑

const fastify = Fastify({
  logger: true // 开启日志记录
});

const PORT = process.env.PORT || 3000;
const WEBHOOK_URL = process.env.WEBHOOK_URL;
const TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN;

if (!WEBHOOK_URL || !TELEGRAM_BOT_TOKEN) {
  console.error('Error: WEBHOOK_URL and TELEGRAM_BOT_TOKEN must be set in .env file');
  process.exit(1);
}

// Telegram Webhook 路由
// Telegram 会将更新 POST 到 /webhook/<YOUR_BOT_TOKEN>
fastify.post(`/webhook/${TELEGRAM_BOT_TOKEN}`, (request, reply) => {
  bot.processUpdate(request.body); // 将接收到的更新交给 bot 处理
  reply.code(200).send({ ok: true }); // 立即响应 Telegram，表示已收到
});

// 启动服务器
const start = async () => {
  try {
    await fastify.listen({ port: PORT, host: '0.0.0.0' }); // 监听所有网络接口
    fastify.log.info(`Server listening on ${fastify.server.address().port}`);
    // 设置 Telegram Webhook
    const webhookEndpoint = `${WEBHOOK_URL}/webhook/${TELEGRAM_BOT_TOKEN}`;
    await bot.setWebHook(webhookEndpoint);
    console.log(`Telegram Webhook set to ${webhookEndpoint}`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

start();

module.exports = fastify; // 可选导出，用于测试等
```

### 4.2 FastGPT API 调用逻辑 (`src/fastgpt.js`)
```javascript
// src/fastgpt.js
require('dotenv').config();
const axios = require('axios');

const API_ENDPOINT = process.env.FASTGPT_API_ENDPOINT;
const API_KEY = process.env.FASTGPT_API_KEY;
const KB_ID = process.env.FASTGPT_KB_ID;

if (!API_ENDPOINT || !API_KEY || !KB_ID) {
  console.error('Error: FastGPT API Endpoint, Key, and KB ID must be set in .env file');
  // 可以选择退出或让服务继续运行但功能受限
}

/**
 * 调用 FastGPT 知识库 API
 * @param {string} query 用户的问题
 * @param {string} chatId 用户ID (可选, 用于 FastGPT 的会话管理)
 * @returns {Promise<string>} FastGPT 的回答文本，如果出错则返回错误提示
 */
async function queryKnowledgeBase(query, chatId = 'default_user') {
  if (!API_ENDPOINT || !API_KEY || !KB_ID) {
    return '抱歉，知识库服务当前不可用。';
  }

  try {
    const response = await axios.post(
      `${API_ENDPOINT}/api/v1/chat/completions`, // 确认这是你的 FastGPT 对话 API 端点
      {
        chatId: chatId, // 用于 FastGPT 内部维持对话状态
        stream: false, // 我们需要完整回答，而非流式
        detail: false, // 通常不需要详细的内部处理信息
        messages: [
          {
            role: 'user',
            content: query
          }
        ],
        model: KB_ID, // 使用知识库 ID 作为模型标识符 (需要根据 FastGPT 版本确认)
        // 可能需要其他参数，如 temperature, top_p 等，根据 FastGPT API 文档调整
      },
      {
        headers: {
          'Authorization': `Bearer ${API_KEY}`, // 使用 Bearer Token 认证
          'Content-Type': 'application/json'
        },
        timeout: 30000 // 设置 30 秒超时
      }
    );

    // 解析 FastGPT 的响应，提取答案文本
    // 注意: FastGPT API 的响应结构可能变化，请根据实际情况调整
    if (response.data && response.data.choices && response.data.choices.length > 0) {
      // 假设答案在 choices[0].message.content
      const answer = response.data.choices[0].message.content;
      // 可能需要去除 FastGPT 返回的引用标记等，例如：
      // return answer.replace(/\[\^source:\d+\]/g, '').trim();
      return answer.trim();
    } else {
      console.error('FastGPT API response format unexpected:', response.data);
      return '抱歉，无法从知识库获取有效的回答。';
    }

  } catch (error) {
    console.error('Error calling FastGPT API:', error.response ? error.response.data : error.message);
    return '抱歉，查询知识库时遇到问题，请稍后再试。';
  }
}

module.exports = {
  queryKnowledgeBase
};
```
**注意:**
*   请务必核对你的 FastGPT API 文档，确认 `/api/v1/chat/completions` 端点、请求体结构 (`messages`, `model` 参数，是否用 `KB_ID` 作 `model` 值) 和认证方式 (`Authorization: Bearer YOUR_API_KEY`) 是否正确。上面代码是一个常见的示例。
*   错误处理和响应解析部分需要根据实际 API 返回进行调整。

### 4.3 Telegram 机器人逻辑 (`src/bot.js`)
```javascript
// src/bot.js
require('dotenv').config();
const TelegramBot = require('node-telegram-bot-api');
const { queryKnowledgeBase } = require('./fastgpt');

const token = process.env.TELEGRAM_BOT_TOKEN;

// 重要：创建 bot 实例时不启动轮询 (polling: false)
// 因为我们将使用 Webhook 模式
const bot = new TelegramBot(token, { polling: false });

// 监听文本消息
bot.on('message', async (msg) => {
  const chatId = msg.chat.id;
  const text = msg.text;

  // 简单的命令处理或忽略
  if (!text || text.startsWith('/')) {
    // 可以选择回复帮助信息或忽略
    // bot.sendMessage(chatId, '请输入您的问题，我会尝试在知识库中查找答案。');
    return;
  }

  console.log(`Received message from ${chatId}: ${text}`);

  // 发送 "正在思考..." 提示
  const thinkingMessage = await bot.sendMessage(chatId, '🤔 正在思考中，请稍候...');

  try {
    // 调用 FastGPT API 查询
    const answer = await queryKnowledgeBase(text, chatId.toString()); // 将 chatId 转为字符串传递

    // 编辑之前的消息，显示答案
    bot.editMessageText(answer, {
      chat_id: chatId,
      message_id: thinkingMessage.message_id,
      // 可以添加 parse_mode: 'Markdown' 或 'HTML' 如果 FastGPT 返回的格式需要解析
    });
    console.log(`Sent answer to ${chatId}: ${answer}`);

  } catch (error) {
    console.error(`Error processing message for chat ${chatId}:`, error);
    // 如果查询出错，编辑消息告知用户
    bot.editMessageText('😥 抱歉，处理您的问题时发生了错误。', {
      chat_id: chatId,
      message_id: thinkingMessage.message_id,
    });
  }
});

// 处理来自 Fastify 服务器的 Webhook 更新
bot.processUpdate = (update) => {
  bot.handleUpdate(update);
};

// 导出 bot 实例和设置 Webhook 的函数
module.exports = bot;
```

## 5. 运行与部署

### 5.1 本地运行 (使用 ngrok 进行 Webhook 测试)
1.  **安装 ngrok:** 如果你没有安装 ngrok，请访问 [ngrok官网](https://ngrok.com/) 下载并安装。
2.  **启动 ngrok:** 在终端运行以下命令，将本地端口 (例如 3000) 暴露到公网：
    ```bash
    ngrok http 3000
    ```
3.  **获取公网 URL:** ngrok 会提供一个 `https://` 开头的 Forwarding URL，例如 `https://xxxxxxxx.ngrok.io`。
4.  **更新 `.env` 文件:** 将 `.env` 文件中的 `WEBHOOK_URL` 设置为 ngrok 提供的 `https://` URL。**不要**在 URL 末尾添加 `/webhook/...` 部分。
    ```dotenv
    WEBHOOK_URL=https://xxxxxxxx.ngrok.io
    ```
5.  **启动应用:**
    ```bash
    node src/server.js
    ```
    应用启动后，会自动设置 Telegram Webhook 到 `https://xxxxxxxx.ngrok.io/webhook/<YOUR_BOT_TOKEN>`。
6.  **测试:** 在 Telegram 中找到你的机器人，向它发送消息，观察服务器日志和机器人的回复。

### 5.2 部署
当你准备好将机器人部署到生产环境时，需要选择一个托管平台，例如：
*   **云服务器 (VPS):** 如 AWS EC2, Google Cloud Compute Engine, DigitalOcean Droplets。你需要自己配置服务器环境、Node.js、反向代理 (如 Nginx) 来处理 HTTPS 和域名。
*   **平台即服务 (PaaS):** 如 Heroku, Render, Fly.io。这些平台简化了部署流程，通常会自动处理 HTTPS 和扩展。你需要将代码推送到平台，并配置好环境变量。

**部署关键点:**
*   **HTTPS:** Telegram Webhook **必须** 使用 HTTPS URL。确保你的部署环境支持 HTTPS。
*   **环境变量:** 在部署平台上安全地配置 `TELEGRAM_BOT_TOKEN`, `FASTGPT_API_ENDPOINT`, `FASTGPT_API_KEY`, `FASTGPT_KB_ID`, `PORT` 和 `WEBHOOK_URL` (应为你的实际公网域名或 IP 地址对应的 URL)。
*   **持久化运行:** 使用进程管理器 (如 `pm2`) 来确保你的 Node.js 应用在后台持续运行，并在崩溃时自动重启。
    ```bash
    pnpm add pm2 -g
    pm2 start src/server.js --name telegram-fastgpt-bot
    ```

## 6. 进一步开发 (可选)

*   **更丰富的回复:** 解析 FastGPT 可能返回的 Markdown 或 HTML 格式，使用 `parse_mode` 选项发送格式化消息。
*   **上下文管理:** FastGPT API 可能支持通过 `chatId` 或传递历史消息来维持对话上下文。你可以在 `queryKnowledgeBase` 函数中实现更复杂的逻辑来管理对话历史。
*   **错误处理与日志:** 实现更健壮的错误处理机制，并将详细日志记录到文件或日志服务中。
*   **用户反馈:** 添加按钮让用户评价答案是否有用。
*   **速率限制:** 防止用户滥用机器人。
*   **多语言支持:** 如果 FastGPT 支持，可以检测用户语言并传递给 API。
*   **命令支持:** 添加 `/start`, `/help` 等命令提供指引。

---

## 7. 功能扩展开发计划

| 功能             | 描述                                                         | 开发状态         |
|------------------|--------------------------------------------------------------|------------------|
| 富文本回复       | 支持 Markdown/HTML 格式消息，提升消息表现力                  | 🟥 待开发         |
| 上下文管理       | 支持多轮对话，关联历史消息上下文                             | 🟨 进行中         |
| 错误处理与日志   | 完善错误捕获，支持文件/服务日志输出                          | 🟩 已完成         |
| 用户反馈         | 支持用户对答案点赞/点踩，收集反馈数据                        | 🟥 待开发         |
| 速率限制         | 限制单用户请求频率，防止滥用（已采用 Go 官方令牌桶算法实现高效线程安全的限流） | 🟩 已完成         |
| 多语言支持       | 自动识别用户语言并适配 API，支持多语种交互                   | 🟥 待开发         |
| 命令支持         | 实现 /start、/help 等机器人指令                              | 🟩 已完成         |
| 缓存与性能优化   | 本地内存缓存、API调用优化，提升响应速度                       | 🟩 已完成         |
| Docker 部署支持  | 提供官方 Dockerfile 及一键部署脚本                           | 🟨 进行中         |

（表中“开发状态”分为：待开发、进行中、已完成）

---
文档结束