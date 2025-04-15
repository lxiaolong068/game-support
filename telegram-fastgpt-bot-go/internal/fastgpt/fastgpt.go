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

	// 构造请求体，符合 searchTest 接口要求
	reqBody := map[string]interface{}{
		"datasetId": cfg.FastGPTKBID,
		"text":      query,
		"limit":     5000,
		"similarity": 0,
		"searchMode": "embedding",
		"usingReRank": false,
		"datasetSearchUsingExtensionQuery": true,
		"datasetSearchExtensionModel": "deepseek-v3",
		"datasetSearchExtensionBg": "",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("JSON编码错误: %v", err)
		return "抱歉，处理请求时出错。", err
	}

	// 请求路径改为 /core/dataset/searchTest
	url := fmt.Sprintf("%s/core/dataset/searchTest", cfg.FastGPTAPIEndpoint)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Printf("创建HTTP请求错误: %v", err)
		return "抱歉，连接知识库服务时出错。", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.FastGPTAPIKey))

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP请求错误: %v", err)
		return "抱歉，查询知识库时遇到网络问题。", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应内容错误: %v", err)
		return "抱歉，处理知识库响应时出错。", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("FastGPT API返回非200状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return "抱歉，知识库返回了错误。", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 解析 searchTest 响应结构
	type SearchTestResp struct {
		Code int `json:"code"`
		StatusText string `json:"statusText"`
		Data []struct {
			Q string `json:"q"`
			A string `json:"a"`
			Score float64 `json:"score"`
		} `json:"data"`
	}

	var result SearchTestResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("解析JSON响应错误: %v, 响应内容: %s", err, string(respBody))
		return "抱歉，解析知识库答案时出错。", err
	}

	if len(result.Data) == 0 {
		log.Printf("知识库未返回有效答案, 响应: %s", string(respBody))
		return "抱歉，未能从知识库中检索到相关答案。", nil
	}

	// 返回最相关的答案
	answer := result.Data[0].A
	return answer, nil
}
