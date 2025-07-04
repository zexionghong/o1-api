package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
	Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
	Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Do(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest HTTP请求
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Timeout time.Duration     `json:"timeout"`
}

// HTTPResponse HTTP响应
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Duration   time.Duration     `json:"duration"`
}

// httpClientImpl HTTP客户端实现
type httpClientImpl struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(timeout time.Duration) HTTPClient {
	return &httpClientImpl{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Get 发送GET请求
func (c *httpClientImpl) Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     url,
		Headers: headers,
		Timeout: c.timeout,
	}
	return c.Do(ctx, req)
}

// Post 发送POST请求
func (c *httpClientImpl) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "POST",
		URL:     url,
		Body:    body,
		Headers: headers,
		Timeout: c.timeout,
	}
	return c.Do(ctx, req)
}

// Put 发送PUT请求
func (c *httpClientImpl) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "PUT",
		URL:     url,
		Body:    body,
		Headers: headers,
		Timeout: c.timeout,
	}
	return c.Do(ctx, req)
}

// Delete 发送DELETE请求
func (c *httpClientImpl) Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
		Timeout: c.timeout,
	}
	return c.Do(ctx, req)
}

// Do 执行HTTP请求
func (c *httpClientImpl) Do(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	start := time.Now()
	
	// 准备请求体
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	
	// 设置请求头
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}
	
	// 如果有请求体，设置Content-Type
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	
	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// 构造响应头
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}
	
	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       respBody,
		Duration:   time.Since(start),
	}, nil
}

// IsSuccess 检查响应是否成功
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// UnmarshalJSON 解析JSON响应体
func (r *HTTPResponse) UnmarshalJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// GetBodyString 获取响应体字符串
func (r *HTTPResponse) GetBodyString() string {
	return string(r.Body)
}
