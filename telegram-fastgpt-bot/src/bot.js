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
