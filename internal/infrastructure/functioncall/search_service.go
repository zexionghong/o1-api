package functioncall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-api-gateway/internal/infrastructure/logger"

	md "github.com/JohannesKaufmann/html-to-markdown"
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
	CrawlContent   bool   `json:"crawl_content"`    // 是否爬取网页内容并转换为Markdown
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
	Content string `json:"content,omitempty"` // 网页内容（Markdown格式）
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
		"config":  fmt.Sprintf("%+v", s.config),
	}).Info("Executing search")

	var results []SearchResult
	var err error

	switch s.config.Service {
	case "search1api":
		s.logger.Info("Using Search1API service")
		results, err = s.searchWithSearch1API(ctx, query, false)
	case "google":
		keyLength := len(s.config.GoogleKey)
		if keyLength > 10 {
			keyLength = 10
		}
		s.logger.WithFields(map[string]interface{}{
			"google_cx":  s.config.GoogleCX,
			"google_key": s.config.GoogleKey[:keyLength] + "...",
		}).Info("Using Google Custom Search service")
		results, err = s.searchWithGoogle(ctx, query, false)
	case "bing":
		s.logger.Info("Using Bing Search service")
		results, err = s.searchWithBing(ctx, query, false)
	case "serpapi":
		s.logger.Info("Using SerpAPI service")
		results, err = s.searchWithSerpAPI(ctx, query, false)
	case "serper":
		s.logger.Info("Using Serper service")
		results, err = s.searchWithSerper(ctx, query, false)
	case "duckduckgo":
		s.logger.Info("Using DuckDuckGo service")
		results, err = s.searchWithDuckDuckGo(ctx, query, false)
	case "searxng":
		s.logger.Info("Using SearXNG service")
		results, err = s.searchWithSearXNG(ctx, query, false)
	default:
		s.logger.WithFields(map[string]interface{}{
			"service": s.config.Service,
		}).Error("Unsupported search service")
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

	s.logger.WithFields(map[string]interface{}{
		"service":      s.config.Service,
		"query":        query,
		"result_count": len(results),
		"results":      results,
	}).Info("Search results obtained")

	// 如果启用了内容爬取，则对每个搜索结果进行内容抓取
	if s.config.CrawlContent {
		s.logger.WithFields(map[string]interface{}{
			"result_count": len(results),
		}).Info("Starting content crawling for search results")

		for i := range results {
			if results[i].Link != "" {
				s.logger.WithFields(map[string]interface{}{
					"url":   results[i].Link,
					"title": results[i].Title,
					"index": i + 1,
				}).Info("Crawling content for search result")

				content, err := s.crawlAndConvertToMarkdown(ctx, results[i].Link)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"url":   results[i].Link,
						"error": err.Error(),
						"index": i + 1,
					}).Warn("Failed to crawl content for search result, continuing with next")
					// 继续处理下一个结果，不因为单个失败而中断整个搜索
					continue
				}

				// 将清理后的内容添加到搜索结果中
				cleanedContent := s.summarizeContent(content, 2000) // 限制每个结果的内容长度
				results[i].Content = cleanedContent
				s.logger.WithFields(map[string]interface{}{
					"url":             results[i].Link,
					"content_length":  len(cleanedContent),
					"original_length": len(content),
					"index":           i + 1,
				}).Info("Successfully crawled and processed content")
			}
		}

		s.logger.WithFields(map[string]interface{}{
			"result_count": len(results),
		}).Info("Content crawling completed for all search results")
	}

	// 格式化搜索结果为结构化文本
	formattedResults := s.formatSearchResults(query, results)

	s.logger.WithFields(map[string]interface{}{
		"service":         s.config.Service,
		"query":           query,
		"result_count":    len(results),
		"response_length": len(formattedResults),
		"crawl_enabled":   s.config.CrawlContent,
	}).Info("Search completed successfully")

	return formattedResults, nil
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

	// 如果启用了内容爬取，则对每个新闻结果进行内容抓取
	if s.config.CrawlContent {
		s.logger.WithFields(map[string]interface{}{
			"result_count": len(results),
		}).Info("Starting content crawling for news results")

		for i := range results {
			if results[i].Link != "" {
				s.logger.WithFields(map[string]interface{}{
					"url":   results[i].Link,
					"title": results[i].Title,
					"index": i + 1,
				}).Info("Crawling content for news result")

				content, err := s.crawlAndConvertToMarkdown(ctx, results[i].Link)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"url":   results[i].Link,
						"error": err.Error(),
						"index": i + 1,
					}).Warn("Failed to crawl content for news result, continuing with next")
					// 继续处理下一个结果，不因为单个失败而中断整个搜索
					continue
				}

				// 将清理后的内容添加到新闻结果中
				cleanedContent := s.summarizeContent(content, 2000) // 限制每个结果的内容长度
				results[i].Content = cleanedContent
				s.logger.WithFields(map[string]interface{}{
					"url":             results[i].Link,
					"content_length":  len(cleanedContent),
					"original_length": len(content),
					"index":           i + 1,
				}).Info("Successfully crawled and processed news content")
			}
		}

		s.logger.WithFields(map[string]interface{}{
			"result_count": len(results),
		}).Info("Content crawling completed for all news results")
	}

	// 格式化新闻搜索结果为结构化文本
	formattedResults := s.formatSearchResults("新闻搜索: "+query, results)

	s.logger.WithFields(map[string]interface{}{
		"service":         s.config.Service,
		"query":           query,
		"result_count":    len(results),
		"response_length": len(formattedResults),
		"crawl_enabled":   s.config.CrawlContent,
	}).Info("News search completed successfully")

	return formattedResults, nil
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

// crawlAndConvertToMarkdown 爬取网页内容并转换为Markdown
func (s *searchServiceImpl) crawlAndConvertToMarkdown(ctx context.Context, url string) (string, error) {
	s.logger.WithFields(map[string]interface{}{
		"url": url,
	}).Info("Crawling URL and converting to Markdown")

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to create HTTP request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置User-Agent避免被反爬虫
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 执行请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to execute HTTP request")
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		s.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"url":         url,
		}).Error("HTTP request failed with non-200 status")
		return "", fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// 读取响应内容
	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to read response body")
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 转换HTML到Markdown
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(htmlContent))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to convert HTML to Markdown")
		return "", fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"url":             url,
		"content_length":  len(markdown),
		"original_length": len(htmlContent),
	}).Info("Successfully converted HTML to Markdown")

	return markdown, nil
}

// cleanMarkdownContent 清理Markdown内容，去除无关部分
func (s *searchServiceImpl) cleanMarkdownContent(content string) string {
	// 去除常见的导航和页脚内容
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	// 正则表达式匹配常见的无关内容
	skipPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(menu|navigation|nav|header|footer|sidebar)`),
		regexp.MustCompile(`(?i)(cookie|privacy|terms|contact|about us|subscribe|newsletter)`),
		regexp.MustCompile(`(?i)^(home|back to top|skip to|jump to)`),
		regexp.MustCompile(`^[\s\*\-\+]*$`), // 空行或只有符号的行
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否应该跳过这一行
		shouldSkip := false
		for _, pattern := range skipPatterns {
			if pattern.MatchString(line) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && len(line) > 10 { // 只保留有意义的内容
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// summarizeContent 总结内容，提取关键信息
func (s *searchServiceImpl) summarizeContent(content string, maxLength int) string {
	cleaned := s.cleanMarkdownContent(content)

	if len(cleaned) <= maxLength {
		return cleaned
	}

	// 按段落分割
	paragraphs := strings.Split(cleaned, "\n")
	var result []string
	currentLength := 0

	for _, paragraph := range paragraphs {
		if currentLength+len(paragraph) > maxLength {
			break
		}
		if len(paragraph) > 20 { // 只保留有意义的段落
			result = append(result, paragraph)
			currentLength += len(paragraph) + 1
		}
	}

	summary := strings.Join(result, "\n")
	if len(summary) < len(cleaned) {
		summary += "\n\n[内容已截断...]"
	}

	return summary
}

// formatSearchResults 格式化搜索结果为结构化文本
func (s *searchServiceImpl) formatSearchResults(query string, results []SearchResult) string {
	var output strings.Builder

	if len(results) == 0 {
		output.WriteString("未找到相关结果。\n")
		return output.String()
	}

	// 提取和整合关键信息
	keyPoints := s.extractKeyPointsWithSources(results)

	// 生成综合回答
	output.WriteString(fmt.Sprintf("根据搜索查询「%s」，为您整理了以下信息：\n\n", query))

	// 主要内容部分 - 包含来源引用
	if len(keyPoints) > 0 {
		for i, point := range keyPoints {
			output.WriteString(fmt.Sprintf("%d. **%s**：%s", i+1, point.Category, point.Content))
			// 添加来源引用
			if len(point.Sources) > 0 {
				output.WriteString(" ")
				for j, source := range point.Sources {
					if j > 0 {
						output.WriteString(", ")
					}
					output.WriteString(fmt.Sprintf("[[%d]]", source.Index))
				}
			}
			output.WriteString("\n\n")
		}
	}

	// 添加数据源URL列表（用户要求的格式）
	output.WriteString("\n\n【重要】请在您的回答最后，必须按照以下确切格式添加数据源：\n\n")
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Link))
	}
	output.WriteString("\n请严格按照上述编号格式在回答末尾列出这些URL，不要修改格式。")

	return output.String()
}

// KeyPoint 关键信息点
type KeyPoint struct {
	Category string      // 分类
	Content  string      // 内容
	Sources  []SourceRef // 来源引用
}

// SourceRef 来源引用
type SourceRef struct {
	Index int    // 来源索引
	Title string // 来源标题
	URL   string // 来源URL
}

// extractKeyPointsWithSources 从搜索结果中提取关键信息点（包含来源）
func (s *searchServiceImpl) extractKeyPointsWithSources(results []SearchResult) []KeyPoint {
	var keyPoints []KeyPoint

	// 分析每个搜索结果，提取关键信息
	for i, result := range results {
		sourceRef := SourceRef{
			Index: i + 1,
			Title: result.Title,
			URL:   result.Link,
		}

		if result.Content != "" {
			// 根据内容长度和质量提取关键信息
			points := s.analyzeContentWithSource(result.Title, result.Content, sourceRef)
			keyPoints = append(keyPoints, points...)
		} else if result.Snippet != "" {
			// 如果没有详细内容，使用摘要
			keyPoints = append(keyPoints, KeyPoint{
				Category: s.categorizeContent(result.Title),
				Content:  result.Snippet,
				Sources:  []SourceRef{sourceRef},
			})
		}
	}

	// 去重和合并相似的信息点
	return s.deduplicateKeyPointsWithSources(keyPoints)
}

// extractKeyPoints 从搜索结果中提取关键信息点（兼容性方法）
func (s *searchServiceImpl) extractKeyPoints(results []SearchResult) []KeyPoint {
	keyPointsWithSources := s.extractKeyPointsWithSources(results)
	// 移除来源信息，保持向后兼容
	var keyPoints []KeyPoint
	for _, point := range keyPointsWithSources {
		keyPoints = append(keyPoints, KeyPoint{
			Category: point.Category,
			Content:  point.Content,
			Sources:  nil,
		})
	}
	return keyPoints
}

// analyzeContentWithSource 分析内容并提取关键信息（包含来源）
func (s *searchServiceImpl) analyzeContentWithSource(title, content string, source SourceRef) []KeyPoint {
	var points []KeyPoint

	// 按段落分析内容
	paragraphs := strings.Split(content, "\n")
	category := s.categorizeContent(title)

	var meaningfulParagraphs []string
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if len(paragraph) > 50 && !s.isBoilerplate(paragraph) {
			meaningfulParagraphs = append(meaningfulParagraphs, paragraph)
		}
	}

	// 合并相关段落
	if len(meaningfulParagraphs) > 0 {
		// 取前3个最有意义的段落
		maxParagraphs := 3
		if len(meaningfulParagraphs) < maxParagraphs {
			maxParagraphs = len(meaningfulParagraphs)
		}

		combinedContent := strings.Join(meaningfulParagraphs[:maxParagraphs], " ")
		if len(combinedContent) > 500 {
			combinedContent = combinedContent[:500] + "..."
		}

		points = append(points, KeyPoint{
			Category: category,
			Content:  combinedContent,
			Sources:  []SourceRef{source},
		})
	}

	return points
}

// analyzeContent 分析内容并提取关键信息（兼容性方法）
func (s *searchServiceImpl) analyzeContent(title, content string) []KeyPoint {
	// 创建一个虚拟的来源引用
	dummySource := SourceRef{Index: 0, Title: title, URL: ""}
	pointsWithSources := s.analyzeContentWithSource(title, content, dummySource)

	// 移除来源信息
	var points []KeyPoint
	for _, point := range pointsWithSources {
		points = append(points, KeyPoint{
			Category: point.Category,
			Content:  point.Content,
			Sources:  nil,
		})
	}

	return points
}

// categorizeContent 根据标题内容进行分类
func (s *searchServiceImpl) categorizeContent(title string) string {
	title = strings.ToLower(title)

	if strings.Contains(title, "定义") || strings.Contains(title, "什么是") || strings.Contains(title, "介绍") {
		return "定义说明"
	} else if strings.Contains(title, "使用") || strings.Contains(title, "应用") || strings.Contains(title, "功能") {
		return "使用方法"
	} else if strings.Contains(title, "特点") || strings.Contains(title, "特性") || strings.Contains(title, "优势") {
		return "主要特点"
	} else if strings.Contains(title, "发展") || strings.Contains(title, "历史") || strings.Contains(title, "趋势") {
		return "发展现状"
	} else if strings.Contains(title, "技术") || strings.Contains(title, "原理") || strings.Contains(title, "实现") {
		return "技术原理"
	} else {
		return "相关信息"
	}
}

// isBoilerplate 判断是否为样板文本（导航、广告等）
func (s *searchServiceImpl) isBoilerplate(text string) bool {
	boilerplatePatterns := []string{
		"cookie", "privacy", "terms", "contact", "subscribe",
		"menu", "navigation", "footer", "header", "sidebar",
		"click here", "read more", "learn more", "sign up",
		"copyright", "all rights reserved", "powered by",
	}

	textLower := strings.ToLower(text)
	for _, pattern := range boilerplatePatterns {
		if strings.Contains(textLower, pattern) {
			return true
		}
	}

	return false
}

// deduplicateKeyPointsWithSources 去重和合并相似的关键信息点（包含来源）
func (s *searchServiceImpl) deduplicateKeyPointsWithSources(points []KeyPoint) []KeyPoint {
	if len(points) <= 1 {
		return points
	}

	var result []KeyPoint
	categoryMap := make(map[string][]KeyPoint)

	// 按分类组织内容
	for _, point := range points {
		categoryMap[point.Category] = append(categoryMap[point.Category], point)
	}

	// 合并同类信息
	for category, categoryPoints := range categoryMap {
		if len(categoryPoints) == 1 {
			result = append(result, categoryPoints[0])
		} else {
			// 合并多个内容，取最有价值的部分
			var contents []string
			var allSources []SourceRef

			for _, point := range categoryPoints {
				contents = append(contents, point.Content)
				allSources = append(allSources, point.Sources...)
			}

			combinedContent := s.combineContents(contents)
			// 去重来源
			uniqueSources := s.deduplicateSources(allSources)

			result = append(result, KeyPoint{
				Category: category,
				Content:  combinedContent,
				Sources:  uniqueSources,
			})
		}
	}

	return result
}

// deduplicateKeyPoints 去重和合并相似的关键信息点（兼容性方法）
func (s *searchServiceImpl) deduplicateKeyPoints(points []KeyPoint) []KeyPoint {
	if len(points) <= 1 {
		return points
	}

	var result []KeyPoint
	categoryMap := make(map[string][]string)

	// 按分类组织内容
	for _, point := range points {
		categoryMap[point.Category] = append(categoryMap[point.Category], point.Content)
	}

	// 合并同类信息
	for category, contents := range categoryMap {
		if len(contents) == 1 {
			result = append(result, KeyPoint{
				Category: category,
				Content:  contents[0],
				Sources:  nil,
			})
		} else {
			// 合并多个内容，取最有价值的部分
			combinedContent := s.combineContents(contents)
			result = append(result, KeyPoint{
				Category: category,
				Content:  combinedContent,
				Sources:  nil,
			})
		}
	}

	return result
}

// deduplicateSources 去重来源引用
func (s *searchServiceImpl) deduplicateSources(sources []SourceRef) []SourceRef {
	seen := make(map[string]bool)
	var result []SourceRef

	for _, source := range sources {
		key := source.URL
		if !seen[key] {
			seen[key] = true
			result = append(result, source)
		}
	}

	return result
}

// combineContents 合并多个内容
func (s *searchServiceImpl) combineContents(contents []string) string {
	if len(contents) == 0 {
		return ""
	}

	// 选择最长且最有信息量的内容作为主要内容
	var bestContent string
	maxScore := 0

	for _, content := range contents {
		score := len(content) // 简单的评分机制，可以后续优化
		if score > maxScore {
			maxScore = score
			bestContent = content
		}
	}

	return bestContent
}
