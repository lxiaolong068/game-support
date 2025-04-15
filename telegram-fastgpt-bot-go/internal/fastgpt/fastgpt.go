package fastgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/telegram-fastgpt-bot-go/internal/config"
)

// FastGPT API请求和响应结构体
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	ChatID  string    `json:"chatId"`
	Stream  bool      `json:"stream"`
	Detail  bool      `json:"detail"`
	Messages []Message `json:"messages"`
	Model   string    `json:"model"`
}

type Choice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// 创建可复用的 HTTP 客户端
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// QueryKnowledgeBase 调用FastGPT知识库API
// query: 用户的问题
// chatID: 用户ID (用于FastGPT的会话管理)
// 返回FastGPT的回答文本，如果出错则返回错误提示
func QueryKnowledgeBase(query string, chatID string) (string, error) {
	cfg := &config.AppConfig

	// 检查配置是否完整
	if cfg.FastGPTAPIEndpoint == "" || cfg.FastGPTAPIKey == "" || cfg.FastGPTKBID == "" {
		return "抱歉，知识库服务当前不可用。", errors.New("FastGPT配置不完整")
	}

	// 使用默认chatID如果未提供
	if chatID == "" {
		chatID = "default_user"
	}

	// 准备请求体
	reqBody := ChatRequest{
		ChatID:  chatID,
		Stream:  false, // 不使用流式响应
		Detail:  false, // 不需要详细的内部处理信息
		Model:   cfg.FastGPTKBID, // 使用知识库ID作为模型标识符
		Messages: []Message{
			{
				Role:    "user",
				Content: query,
			},
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("JSON编码错误: %v", err)
		return "抱歉，处理请求时出错。", err
	}

	// 准备HTTP请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/chat/completions", cfg.FastGPTAPIEndpoint), bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Printf("创建HTTP请求错误: %v", err)
		return "抱歉，连接知识库服务时出错。", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.FastGPTAPIKey))

	// 发送请求（使用全局 httpClient）
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP请求错误: %v", err)
		return "抱歉，查询知识库时遇到网络问题。", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应内容错误: %v", err)
		return "抱歉，处理知识库响应时出错。", err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("FastGPT API返回非200状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return "抱歉，知识库返回了错误。", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 解析JSON响应
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		log.Printf("解析JSON响应错误: %v, 响应内容: %s", err, string(respBody))
		return "抱歉，解析知识库答案时出错。", err
	}

	// 检查是否有有效答案
	if len(chatResp.Choices) == 0 {
		log.Printf("FastGPT API返回了空的choices数组, 响应: %s", string(respBody))
		return "抱歉，无法从知识库获取有效的回答。", errors.New("空的choices数组")
	}

	// 返回处理后的答案
	answer := chatResp.Choices[0].Message.Content
	return answer, nil
}
