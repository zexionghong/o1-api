package functioncall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// searchWithSearch1API 使用 Search1API 进行搜索
func (s *searchServiceImpl) searchWithSearch1API(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	var apiURL string
	if isNews {
		apiURL = "https://api.search1api.com/news"
	} else {
		apiURL = "https://api.search1api.com/search/"
	}

	requestBody := map[string]interface{}{
		"query":          query,
		"search_service": "google",
		"max_results":    s.config.MaxResults,
		"crawl_results":  s.config.CrawlResults,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.config.Search1APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.Search1APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Results []SearchResult `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResponse.Results, nil
}

// searchWithGoogle 使用 Google Custom Search API 进行搜索
func (s *searchServiceImpl) searchWithGoogle(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("cx", s.config.GoogleCX)
	params.Set("key", s.config.GoogleKey)
	params.Set("q", query)
	if isNews {
		params.Set("tbm", "nws")
	}

	apiURL := baseURL + "?" + params.Encode()

	s.logger.WithFields(map[string]interface{}{
		"url":     apiURL,
		"query":   query,
		"is_news": isNews,
		"cx":      s.config.GoogleCX,
	}).Info("Making Google Custom Search API request")

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to create Google API request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to execute Google API request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
	}).Info("Received Google API response")

	if resp.StatusCode != http.StatusOK {
		s.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
		}).Error("Google API request failed with non-200 status")
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to decode Google API response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查API错误
	if apiResponse.Error.Code != 0 {
		s.logger.WithFields(map[string]interface{}{
			"error_code":    apiResponse.Error.Code,
			"error_message": apiResponse.Error.Message,
		}).Error("Google API returned error")
		return nil, fmt.Errorf("Google API error: %s (code: %d)", apiResponse.Error.Message, apiResponse.Error.Code)
	}

	s.logger.WithFields(map[string]interface{}{
		"items_count": len(apiResponse.Items),
	}).Info("Successfully decoded Google API response")

	var results []SearchResult
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	for i, item := range apiResponse.Items {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		})
	}

	s.logger.WithFields(map[string]interface{}{
		"results_count": len(results),
		"max_results":   maxResults,
	}).Info("Google search completed successfully")

	return results, nil
}

// searchWithBing 使用 Bing Search API 进行搜索
func (s *searchServiceImpl) searchWithBing(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	var apiURL string
	if isNews {
		apiURL = "https://api.bing.microsoft.com/v7.0/news/search?q=" + url.QueryEscape(query)
	} else {
		apiURL = "https://api.bing.microsoft.com/v7.0/search?q=" + url.QueryEscape(query)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", s.config.BingKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var results []SearchResult
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	if isNews {
		var apiResponse struct {
			Value []struct {
				Name        string `json:"name"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"value"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.Value {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Name,
				Link:    item.URL,
				Snippet: item.Description,
			})
		}
	} else {
		var apiResponse struct {
			WebPages struct {
				Value []struct {
					Name    string `json:"name"`
					URL     string `json:"url"`
					Snippet string `json:"snippet"`
				} `json:"value"`
			} `json:"webPages"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.WebPages.Value {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Name,
				Link:    item.URL,
				Snippet: item.Snippet,
			})
		}
	}

	return results, nil
}

// searchWithSerpAPI 使用 SerpAPI 进行搜索
func (s *searchServiceImpl) searchWithSerpAPI(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	baseURL := "https://serpapi.com/search"
	params := url.Values{}
	params.Set("api_key", s.config.SerpAPIKey)
	if isNews {
		params.Set("engine", "google_news")
	} else {
		params.Set("engine", "google")
	}
	params.Set("q", query)
	params.Set("google_domain", "google.com")

	apiURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var results []SearchResult
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	if isNews {
		var apiResponse struct {
			NewsResults []struct {
				Title   string `json:"title"`
				Link    string `json:"link"`
				Snippet string `json:"snippet"`
			} `json:"news_results"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.NewsResults {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Title,
				Link:    item.Link,
				Snippet: item.Snippet,
			})
		}
	} else {
		var apiResponse struct {
			OrganicResults []struct {
				Title   string `json:"title"`
				Link    string `json:"link"`
				Snippet string `json:"snippet"`
			} `json:"organic_results"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.OrganicResults {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Title,
				Link:    item.Link,
				Snippet: item.Snippet,
			})
		}
	}

	return results, nil
}

// searchWithSerper 使用 Serper API 进行搜索
func (s *searchServiceImpl) searchWithSerper(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	var apiURL string
	if isNews {
		apiURL = "https://google.serper.dev/news"
	} else {
		apiURL = "https://google.serper.dev/search"
	}

	requestBody := map[string]interface{}{
		"q":  query,
		"gl": "us",
		"hl": "en",
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-KEY", s.config.SerperKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var results []SearchResult
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	if isNews {
		var apiResponse struct {
			News []struct {
				Title   string `json:"title"`
				Link    string `json:"link"`
				Snippet string `json:"snippet"`
			} `json:"news"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.News {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Title,
				Link:    item.Link,
				Snippet: item.Snippet,
			})
		}
	} else {
		var apiResponse struct {
			Organic []struct {
				Title   string `json:"title"`
				Link    string `json:"link"`
				Snippet string `json:"snippet"`
			} `json:"organic"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for i, item := range apiResponse.Organic {
			if i >= maxResults {
				break
			}
			results = append(results, SearchResult{
				Title:   item.Title,
				Link:    item.Link,
				Snippet: item.Snippet,
			})
		}
	}

	return results, nil
}

// searchWithDuckDuckGo 使用 DuckDuckGo 进行搜索
func (s *searchServiceImpl) searchWithDuckDuckGo(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	var apiURL string
	if isNews {
		apiURL = "https://ddg.search2ai.online/searchNews"
	} else {
		apiURL = "https://ddg.search2ai.online/search"
	}

	requestBody := map[string]interface{}{
		"q":           query,
		"max_results": s.config.MaxResults,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Results []struct {
			Title string `json:"title"`
			Href  string `json:"href"`
			URL   string `json:"url"`
			Body  string `json:"body"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []SearchResult
	for _, item := range apiResponse.Results {
		link := item.Href
		if link == "" {
			link = item.URL
		}
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    link,
			Snippet: item.Body,
		})
	}

	return results, nil
}

// searchWithSearXNG 使用 SearXNG 进行搜索
func (s *searchServiceImpl) searchWithSearXNG(ctx context.Context, query string, isNews bool) ([]SearchResult, error) {
	baseURL := s.config.SearXNGBaseURL + "/search"
	params := url.Values{}
	params.Set("q", query)
	if isNews {
		params.Set("category", "news")
	} else {
		params.Set("category", "general")
	}
	params.Set("format", "json")

	apiURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []SearchResult
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	for i, item := range apiResponse.Results {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    item.URL,
			Snippet: item.Content,
		})
	}

	return results, nil
}
