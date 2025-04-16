package fastgpt

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lxiaolong068/game-support/telegram-fastgpt-bot-go/internal/config"
	"go.uber.org/zap"
	"github.com/patrickmn/go-cache"
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

// 全局缓存对象
var gptCache *cache.Cache

func init() {
	cfg := &config.AppConfig
	if cfg.CacheExpiration <= 0 {
		cfg.CacheExpiration = 120
	}
	gptCache = cache.New(time.Duration(cfg.CacheExpiration)*time.Second, 2*time.Minute)
}


// QueryKnowledgeBase 兼容原接口，内部调用支持 context 的实现
func QueryKnowledgeBase(query string, chatID string) (string, error) {
	return QueryKnowledgeBaseWithContext(context.Background(), query, chatID)
}

// QueryKnowledgeBaseWithContext 支持 context 的 API 调用
func QueryKnowledgeBaseWithContext(ctx context.Context, query string, chatID string) (string, error) {
	cfg := &config.AppConfig
	// 缓存开关
	if cfg.EnableCache {
		cacheKey := makeCacheKey(query, chatID)
		if v, found := gptCache.Get(cacheKey); found {
			if answer, ok := v.(string); ok {
				config.Logger.Info("命中本地缓存", zap.String("chat_id", chatID), zap.String("q", query))
				return answer, nil
			}
		}
	}

	// 检查配置是否完整
	if cfg.FastGPTAPIEndpoint == "" || cfg.FastGPTAPIKey == "" || cfg.FastGPTKBID == "" {
		config.Logger.Error("FastGPT配置不完整", zap.String("endpoint", cfg.FastGPTAPIEndpoint), zap.String("key", cfg.FastGPTAPIKey), zap.String("kb_id", cfg.FastGPTKBID))
		return "抱歉，知识库服务当前不可用。", errors.New("FastGPT配置不完整")
	}

	// 构造请求体，参数全部从 config 读取
	reqBody := map[string]interface{}{
		"datasetId": cfg.FastGPTKBID,
		"text":      query,
		"limit":     cfg.FastGPTLimit,
		"similarity": cfg.FastGPTSimilarity,
		"searchMode": cfg.FastGPTSearchMode,
		"usingReRank": cfg.FastGPTUsingReRank,
		"datasetSearchUsingExtensionQuery": true,
		"datasetSearchExtensionModel": cfg.FastGPTDatasetSearchExtensionModel,
		"datasetSearchExtensionBg": "",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		config.Logger.Error("JSON编码错误", zap.Error(err))
		return "抱歉，处理请求时出错。", err
	}

	// 请求路径改为 /core/dataset/searchTest
	url := fmt.Sprintf("%s/core/dataset/searchTest", cfg.FastGPTAPIEndpoint)
	
	var resp *http.Response
	var respBody []byte
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
		if reqErr != nil {
			config.Logger.Error("创建HTTP请求错误", zap.Error(reqErr))
			return "抱歉，连接知识库服务时出错。", reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.FastGPTAPIKey))

		resp, err = httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return "抱歉，请求超时或被取消。", ctx.Err()
			}
			config.Logger.Warn("HTTP请求失败，自动重试", zap.Error(err), zap.Int("attempt", attempt+1))
			time.Sleep(time.Duration(1<<attempt) * 300 * time.Millisecond) // 指数退避
			continue
		}
		defer resp.Body.Close()
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			config.Logger.Error("读取响应内容错误", zap.Error(err))
			return "抱歉，处理知识库响应时出错。", err
		}
		if resp.StatusCode != http.StatusOK {
			config.Logger.Error("FastGPT API返回非200状态码", zap.Int("status_code", resp.StatusCode), zap.ByteString("response", respBody))
			return "抱歉，知识库返回了错误。", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
		}
		break // 请求成功
	}
	if err != nil {
		return "抱歉，查询知识库时遇到网络问题。", err
	}

	// 解析 searchTest 响应结构
	type ScoreItem struct {
		Type  string  `json:"type"`
		Value float64 `json:"value"`
	}

	type SearchResultItem struct {
		ID         string      `json:"id"`
		UpdateTime string      `json:"updateTime"`
		Q          string      `json:"q"`
		A          string      `json:"a"`
		ChunkIndex int         `json:"chunkIndex"`
		DatasetID  string      `json:"datasetId"`
		CollectionID string    `json:"collectionId"`
		SourceID   string      `json:"sourceId"`
		SourceName string      `json:"sourceName"`
		Score      []ScoreItem `json:"score"` // Score 是一个对象数组
		Tokens     int         `json:"tokens"`
	}

	type DataPayload struct {
		List                []SearchResultItem `json:"list"`
		Duration            string             `json:"duration"`
		QueryExtensionModel string             `json:"queryExtensionModel"`
		SearchMode          string             `json:"searchMode"`
		Limit               int                `json:"limit"`
		Similarity          float64            `json:"similarity"` // 注意：JSON 中是 0，可能是 int 或 float
		UsingReRank         bool               `json:"usingReRank"`
		UsingSimilarityFilter bool             `json:"usingSimilarityFilter"`
	}

	type SearchTestResp struct {
		Code       int         `json:"code"`
		StatusText string      `json:"statusText"`
		Message    string      `json:"message"` // 添加 message 字段
		Data       DataPayload `json:"data"`    // Data 是一个对象，包含 List
	}

	var result SearchTestResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		config.Logger.Error("解析JSON响应错误", zap.Error(err), zap.ByteString("response", respBody))
		return "抱歉，解析知识库答案时出错。", err
	}

	// 检查 Data.List 是否为空
	if len(result.Data.List) == 0 {
		config.Logger.Warn("知识库未返回有效答案 (list is empty)", zap.ByteString("response", respBody))
		return "抱歉，未能从知识库中检索到相关答案。", nil
	}

	// 返回最相关的答案
	answer := result.Data.List[0].A
	config.Logger.Info("FastGPT应答成功", zap.String("answer", answer))
	// 写入缓存
	if cfg.EnableCache {
		cacheKey := makeCacheKey(query, chatID)
		gptCache.Set(cacheKey, answer, time.Duration(cfg.CacheExpiration)*time.Second)
	}
	return answer, nil
}

// 生成缓存 key，防止隐私泄漏
func makeCacheKey(q, chatID string) string {
	h := sha256.New()
	h.Write([]byte(q + ":" + chatID))
	return hex.EncodeToString(h.Sum(nil))
}
