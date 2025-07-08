package functioncall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-api-gateway/internal/infrastructure/logger"
)

// SearchService 搜索服务接口
type SearchService interface {
	Search(ctx context.Context, query string) (string, error)
	SearchNews(ctx context.Context, query string) (string, error)
	CrawlURL(ctx context.Context, url string) (string, error)
}

// SearchConfig 搜索服务配置
type SearchConfig struct {
	Service        string `json:"service"`          // 搜索服务类型
	MaxResults     int    `json:"max_results"`      // 最大结果数
	CrawlResults   int    `json:"crawl_results"`    // 深度搜索数量
	Search1APIKey  string `json:"search1api_key"`   // Search1API密钥
	GoogleCX       string `json:"google_cx"`        // Google自定义搜索引擎ID
	GoogleKey      string `json:"google_key"`       // Google API密钥
	BingKey        string `json:"bing_key"`         // Bing搜索API密钥
	SerpAPIKey     string `json:"serpapi_key"`      // SerpAPI密钥
	SerperKey      string `json:"serper_key"`       // Serper密钥
	SearXNGBaseURL string `json:"searxng_base_url"` // SearXNG服务地址
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// CrawlResponse 爬取响应
type CrawlResponse struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// searchServiceImpl 搜索服务实现
type searchServiceImpl struct {
	config     *SearchConfig
	logger     logger.Logger
	httpClient *http.Client
}

// NewSearchService 创建搜索服务
func NewSearchService(config *SearchConfig, logger logger.Logger) SearchService {
	return &searchServiceImpl{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search 执行搜索
func (s *searchServiceImpl) Search(ctx context.Context, query string) (string, error) {
	s.logger.WithFields(map[string]interface{}{
		"service": s.config.Service,
		"query":   query,
	}).Info("Executing search")

	var results []SearchResult
	var err error

	switch s.config.Service {
	case "search1api":
		results, err = s.searchWithSearch1API(ctx, query, false)
	case "google":
		results, err = s.searchWithGoogle(ctx, query, false)
	case "bing":
		results, err = s.searchWithBing(ctx, query, false)
	case "serpapi":
		results, err = s.searchWithSerpAPI(ctx, query, false)
	case "serper":
		results, err = s.searchWithSerper(ctx, query, false)
	case "duckduckgo":
		results, err = s.searchWithDuckDuckGo(ctx, query, false)
	case "searxng":
		results, err = s.searchWithSearXNG(ctx, query, false)
	default:
		return "", fmt.Errorf("unsupported search service: %s", s.config.Service)
	}

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"service": s.config.Service,
			"query":   query,
			"error":   err.Error(),
		}).Error("Search failed")
		return "", err
	}

	response := SearchResponse{Results: results}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search response: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"service":      s.config.Service,
		"query":        query,
		"result_count": len(results),
	}).Info("Search completed successfully")

	return string(responseJSON), nil
}

// SearchNews 执行新闻搜索
func (s *searchServiceImpl) SearchNews(ctx context.Context, query string) (string, error) {
	s.logger.WithFields(map[string]interface{}{
		"service": s.config.Service,
		"query":   query,
	}).Info("Executing news search")

	var results []SearchResult
	var err error

	switch s.config.Service {
	case "search1api":
		results, err = s.searchWithSearch1API(ctx, query, true)
	case "google":
		results, err = s.searchWithGoogle(ctx, query, true)
	case "bing":
		results, err = s.searchWithBing(ctx, query, true)
	case "serpapi":
		results, err = s.searchWithSerpAPI(ctx, query, true)
	case "serper":
		results, err = s.searchWithSerper(ctx, query, true)
	case "duckduckgo":
		results, err = s.searchWithDuckDuckGo(ctx, query, true)
	case "searxng":
		results, err = s.searchWithSearXNG(ctx, query, true)
	default:
		return "", fmt.Errorf("unsupported search service: %s", s.config.Service)
	}

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"service": s.config.Service,
			"query":   query,
			"error":   err.Error(),
		}).Error("News search failed")
		return "", err
	}

	response := SearchResponse{Results: results}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal news search response: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"service":      s.config.Service,
		"query":        query,
		"result_count": len(results),
	}).Info("News search completed successfully")

	return string(responseJSON), nil
}

// CrawlURL 爬取网页内容
func (s *searchServiceImpl) CrawlURL(ctx context.Context, targetURL string) (string, error) {
	s.logger.WithFields(map[string]interface{}{
		"url": targetURL,
	}).Info("Crawling URL")

	// 使用通用的爬取服务
	crawlURL := "https://crawl.search1api.com"

	requestBody := map[string]interface{}{
		"url": targetURL,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal crawl request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", crawlURL, strings.NewReader(string(requestJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create crawl request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute crawl request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("crawl request failed with status: %d", resp.StatusCode)
	}

	var crawlResponse CrawlResponse
	if err := json.NewDecoder(resp.Body).Decode(&crawlResponse); err != nil {
		return "", fmt.Errorf("failed to decode crawl response: %w", err)
	}

	responseJSON, err := json.Marshal(crawlResponse)
	if err != nil {
		return "", fmt.Errorf("failed to marshal crawl response: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"url":   targetURL,
		"title": crawlResponse.Title,
	}).Info("URL crawling completed successfully")

	return string(responseJSON), nil
}
