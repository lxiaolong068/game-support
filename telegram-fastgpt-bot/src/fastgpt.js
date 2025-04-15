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
