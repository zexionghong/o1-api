package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// SwaggerInfo Swagger基本信息
type SwaggerInfo struct {
	Version     string `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Contact     struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		URL   string `json:"url"`
	} `json:"contact"`
	License struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"license"`
}

// SwaggerDoc Swagger文档结构
type SwaggerDoc struct {
	Swagger     string                 `json:"swagger"`
	Info        SwaggerInfo            `json:"info"`
	Host        string                 `json:"host"`
	BasePath    string                 `json:"basePath"`
	Schemes     []string               `json:"schemes"`
	Consumes    []string               `json:"consumes"`
	Produces    []string               `json:"produces"`
	Paths       map[string]interface{} `json:"paths"`
	Definitions map[string]interface{} `json:"definitions"`
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions"`
}

func main() {
	fmt.Println("🔧 Generating Swagger documentation...")

	// 创建Swagger文档
	doc := SwaggerDoc{
		Swagger:  "2.0",
		Host:     "localhost:8080",
		BasePath: "/",
		Schemes:  []string{"http", "https"},
		Consumes: []string{"application/json"},
		Produces: []string{"application/json"},
		Info: SwaggerInfo{
			Version:     "1.0.0",
			Title:       "AI API Gateway",
			Description: "AI API Gateway是一个高性能的AI API网关，提供统一的API接口来访问多个AI提供商。\n\n主要功能：\n- 多AI提供商支持（OpenAI、Anthropic等）\n- 智能负载均衡和故障转移\n- 精确的配额管理和计费\n- 完整的认证和授权\n- 实时监控和统计",
		},
		Paths:       make(map[string]interface{}),
		Definitions: make(map[string]interface{}),
		SecurityDefinitions: map[string]interface{}{
			"ApiKeyAuth": map[string]interface{}{
				"type": "apiKey",
				"in":   "header",
				"name": "Authorization",
				"description": "API密钥认证，格式：Bearer YOUR_API_KEY",
			},
		},
	}

	// 设置联系信息
	doc.Info.Contact.Name = "AI API Gateway Team"
	doc.Info.Contact.Email = "support@example.com"
	doc.Info.Contact.URL = "https://example.com/support"

	// 设置许可证信息
	doc.Info.License.Name = "MIT"
	doc.Info.License.URL = "https://opensource.org/licenses/MIT"

	// 添加健康检查路径
	doc.Paths["/health/ready"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"health"},
			"summary":     "就绪检查",
			"description": "检查服务是否已准备好接收请求",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "服务就绪",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/HealthResponse",
					},
				},
				"503": map[string]interface{}{
					"description": "服务未就绪",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
			},
		},
	}

	// 添加聊天补全路径
	doc.Paths["/v1/chat/completions"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"ai"},
			"summary":     "聊天补全",
			"description": "创建聊天补全请求，兼容OpenAI API格式",
			"security": []map[string]interface{}{
				{"ApiKeyAuth": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "body",
					"in":          "body",
					"description": "聊天补全请求",
					"required":    true,
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ChatCompletionRequest",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "聊天补全响应",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ChatCompletionResponse",
					},
				},
				"400": map[string]interface{}{
					"description": "请求参数错误",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
				"401": map[string]interface{}{
					"description": "认证失败",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
				"429": map[string]interface{}{
					"description": "请求过于频繁",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
			},
		},
	}

	// 添加模型列表路径
	doc.Paths["/v1/models"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"ai"},
			"summary":     "列出模型",
			"description": "获取可用的AI模型列表",
			"security": []map[string]interface{}{
				{"ApiKeyAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "模型列表",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ModelsResponse",
					},
				},
				"401": map[string]interface{}{
					"description": "认证失败",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
			},
		},
	}

	// 添加用户管理路径
	doc.Paths["/admin/users"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"admin"},
			"summary":     "列出用户",
			"description": "获取用户列表",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "用户列表",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/UsersListResponse",
					},
				},
			},
		},
		"post": map[string]interface{}{
			"tags":        []string{"admin"},
			"summary":     "创建用户",
			"description": "创建新的用户账户",
			"parameters": []map[string]interface{}{
				{
					"name":        "body",
					"in":          "body",
					"description": "用户创建请求",
					"required":    true,
					"schema": map[string]interface{}{
						"$ref": "#/definitions/CreateUserRequest",
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "用户创建成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/UserResponse",
					},
				},
				"400": map[string]interface{}{
					"description": "请求参数错误",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/ErrorResponse",
					},
				},
			},
		},
	}

	// 添加基本定义
	doc.Definitions["ErrorResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": false,
			},
			"error": map[string]interface{}{
				"$ref": "#/definitions/Error",
			},
			"timestamp": map[string]interface{}{
				"type":    "string",
				"example": "2024-01-01T00:00:00Z",
			},
		},
	}

	doc.Definitions["Error"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":    "string",
				"example": "INVALID_REQUEST",
			},
			"message": map[string]interface{}{
				"type":    "string",
				"example": "请求参数无效",
			},
			"details": map[string]interface{}{
				"type": "object",
			},
		},
	}

	doc.Definitions["HealthResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": true,
			},
			"status": map[string]interface{}{
				"type":    "string",
				"example": "healthy",
			},
			"message": map[string]interface{}{
				"type":    "string",
				"example": "Service is healthy",
			},
		},
	}

	// 确保docs目录存在
	if err := os.MkdirAll("docs", 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// 生成JSON文件
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal swagger doc: %v", err)
	}

	if err := os.WriteFile("docs/swagger.json", jsonData, 0644); err != nil {
		log.Fatalf("Failed to write swagger.json: %v", err)
	}

	fmt.Println("✅ Swagger documentation generated successfully!")
	fmt.Println("📄 Files created:")
	fmt.Println("   - docs/swagger.json")
	fmt.Println()
	fmt.Println("🌐 Access Swagger UI at:")
	fmt.Println("   - http://localhost:8080/swagger/index.html")
	fmt.Println()
	fmt.Println("💡 To view the documentation:")
	fmt.Println("   1. Start the server: go run cmd/server/main.go")
	fmt.Println("   2. Open browser: http://localhost:8080/swagger/index.html")
}
