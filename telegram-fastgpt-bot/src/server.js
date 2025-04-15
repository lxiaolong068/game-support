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
