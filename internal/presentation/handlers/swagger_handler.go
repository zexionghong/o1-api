package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SwaggerHandler Swagger文档处理器
type SwaggerHandler struct{}

// NewSwaggerHandler 创建Swagger处理器
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// SwaggerUI 提供Swagger UI界面
func (h *SwaggerHandler) SwaggerUI(c *gin.Context) {
	// 获取请求路径
	path := c.Param("any")

	// 调试日志
	c.Header("X-Debug-Path", path)

	// 如果是根路径或空路径，重定向到index.html
	if path == "" || path == "/" || path == "/index.html" {
		h.serveSwaggerIndex(c)
		return
	}

	// 处理index.html请求
	if path == "index.html" || strings.HasSuffix(path, "index.html") {
		h.serveSwaggerIndex(c)
		return
	}

	// 处理swagger.json请求
	if path == "swagger.json" || strings.HasSuffix(path, "swagger.json") {
		h.serveSwaggerJSON(c)
		return
	}

	// 处理其他静态资源（简化版本）
	if strings.HasSuffix(path, ".css") {
		c.Header("Content-Type", "text/css")
		c.String(http.StatusOK, h.getSwaggerCSS())
		return
	}

	if strings.HasSuffix(path, ".js") {
		c.Header("Content-Type", "application/javascript")
		c.String(http.StatusOK, h.getSwaggerJS())
		return
	}

	// 如果没有匹配到任何路径，默认显示主页
	h.serveSwaggerIndex(c)
}

// serveSwaggerIndex 提供Swagger UI主页面
func (h *SwaggerHandler) serveSwaggerIndex(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI API Gateway - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                },
                requestInterceptor: function(request) {
                    // 可以在这里添加默认的请求头
                    console.log('Request:', request);
                    return request;
                },
                responseInterceptor: function(response) {
                    console.log('Response:', response);
                    return response;
                }
            });
        };
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// serveSwaggerJSON 提供Swagger JSON规范
func (h *SwaggerHandler) serveSwaggerJSON(c *gin.Context) {
	// 完整的Swagger JSON
	swaggerJSON := `{
  "swagger": "2.0",
  "info": {
    "title": "AI API Gateway",
    "description": "AI API Gateway是一个高性能的AI API网关，提供统一的API接口来访问多个AI提供商。支持OpenAI、Anthropic等多个AI提供商，具备完整的认证、配额管理、计费和监控功能。",
    "version": "1.0.0",
    "contact": {
      "name": "AI API Gateway Team",
      "email": "support@example.com"
    },
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    }
  },
  "host": "localhost:8080",
  "basePath": "/",
  "schemes": ["http", "https"],
  "consumes": ["application/json"],
  "produces": ["application/json"],
  "securityDefinitions": {
    "ApiKeyAuth": {
      "type": "apiKey",
      "in": "header",
      "name": "Authorization",
      "description": "API密钥认证，格式：Bearer YOUR_API_KEY"
    }
  },
  "definitions": {
    "ChatCompletionRequest": {
      "type": "object",
      "required": ["model", "messages"],
      "properties": {
        "model": {
          "type": "string",
          "description": "要使用的模型ID",
          "example": "gpt-3.5-turbo"
        },
        "messages": {
          "type": "array",
          "description": "聊天消息列表",
          "items": {
            "$ref": "#/definitions/ChatMessage"
          }
        },
        "max_tokens": {
          "type": "integer",
          "description": "生成的最大token数",
          "example": 100
        },
        "temperature": {
          "type": "number",
          "description": "采样温度，0-2之间",
          "example": 0.7
        },
        "stream": {
          "type": "boolean",
          "description": "是否流式返回",
          "example": false
        }
      }
    },
    "ChatMessage": {
      "type": "object",
      "required": ["role", "content"],
      "properties": {
        "role": {
          "type": "string",
          "enum": ["system", "user", "assistant"],
          "description": "消息角色"
        },
        "content": {
          "type": "string",
          "description": "消息内容"
        }
      }
    },
    "ChatCompletionResponse": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "description": "请求ID"
        },
        "object": {
          "type": "string",
          "example": "chat.completion"
        },
        "created": {
          "type": "integer",
          "description": "创建时间戳"
        },
        "model": {
          "type": "string",
          "description": "使用的模型"
        },
        "choices": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/ChatChoice"
          }
        },
        "usage": {
          "$ref": "#/definitions/Usage"
        }
      }
    },
    "ChatChoice": {
      "type": "object",
      "properties": {
        "index": {
          "type": "integer"
        },
        "message": {
          "$ref": "#/definitions/ChatMessage"
        },
        "finish_reason": {
          "type": "string",
          "enum": ["stop", "length", "content_filter"]
        }
      }
    },
    "Usage": {
      "type": "object",
      "properties": {
        "prompt_tokens": {
          "type": "integer",
          "description": "输入token数"
        },
        "completion_tokens": {
          "type": "integer",
          "description": "输出token数"
        },
        "total_tokens": {
          "type": "integer",
          "description": "总token数"
        }
      }
    },
    "Model": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "description": "模型ID"
        },
        "object": {
          "type": "string",
          "example": "model"
        },
        "created": {
          "type": "integer",
          "description": "创建时间戳"
        },
        "owned_by": {
          "type": "string",
          "description": "模型提供商"
        }
      }
    },
    "Error": {
      "type": "object",
      "properties": {
        "error": {
          "type": "object",
          "properties": {
            "message": {
              "type": "string",
              "description": "错误信息"
            },
            "type": {
              "type": "string",
              "description": "错误类型"
            },
            "code": {
              "type": "string",
              "description": "错误代码"
            }
          }
        }
      }
    }
  },
  "paths": {
    "/health/ready": {
      "get": {
        "tags": ["健康检查"],
        "summary": "就绪检查",
        "description": "检查服务是否已准备好接收请求",
        "responses": {
          "200": {
            "description": "服务就绪",
            "schema": {
              "type": "object",
              "properties": {
                "status": {"type": "string", "example": "ready"},
                "timestamp": {"type": "string", "example": "2024-01-01T00:00:00Z"}
              }
            }
          },
          "503": {
            "description": "服务未就绪",
            "schema": {"$ref": "#/definitions/Error"}
          }
        }
      }
    },
    "/health/live": {
      "get": {
        "tags": ["健康检查"],
        "summary": "存活检查",
        "description": "检查服务是否正在运行",
        "responses": {
          "200": {
            "description": "服务正在运行",
            "schema": {
              "type": "object",
              "properties": {
                "status": {"type": "string", "example": "alive"},
                "timestamp": {"type": "string", "example": "2024-01-01T00:00:00Z"}
              }
            }
          }
        }
      }
    },
    "/health/stats": {
      "get": {
        "tags": ["健康检查"],
        "summary": "系统统计",
        "description": "获取系统运行统计信息",
        "responses": {
          "200": {
            "description": "系统统计信息",
            "schema": {
              "type": "object",
              "properties": {
                "uptime": {"type": "string", "example": "24h30m15s"},
                "requests_total": {"type": "integer", "example": 12345},
                "requests_success": {"type": "integer", "example": 12000},
                "memory_usage": {"type": "string", "example": "256MB"}
              }
            }
          }
        }
      }
    },
    "/v1/chat/completions": {
      "post": {
        "tags": ["AI接口"],
        "summary": "聊天补全",
        "description": "创建聊天补全请求，兼容OpenAI API格式",
        "security": [{"ApiKeyAuth": []}],
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {"$ref": "#/definitions/ChatCompletionRequest"}
          }
        ],
        "responses": {
          "200": {
            "description": "聊天补全响应",
            "schema": {"$ref": "#/definitions/ChatCompletionResponse"}
          },
          "400": {
            "description": "请求参数错误",
            "schema": {"$ref": "#/definitions/Error"}
          },
          "401": {
            "description": "认证失败",
            "schema": {"$ref": "#/definitions/Error"}
          },
          "429": {
            "description": "请求过于频繁",
            "schema": {"$ref": "#/definitions/Error"}
          }
        }
      }
    },
    "/v1/models": {
      "get": {
        "tags": ["AI接口"],
        "summary": "列出模型",
        "description": "获取可用的AI模型列表",
        "security": [{"ApiKeyAuth": []}],
        "responses": {
          "200": {
            "description": "模型列表",
            "schema": {
              "type": "object",
              "properties": {
                "object": {"type": "string", "example": "list"},
                "data": {
                  "type": "array",
                  "items": {"$ref": "#/definitions/Model"}
                }
              }
            }
          },
          "401": {
            "description": "认证失败",
            "schema": {"$ref": "#/definitions/Error"}
          }
        }
      }
    },
    "/v1/usage": {
      "get": {
        "tags": ["AI接口"],
        "summary": "使用统计",
        "description": "获取当前用户的API使用统计",
        "security": [{"ApiKeyAuth": []}],
        "responses": {
          "200": {
            "description": "使用统计信息",
            "schema": {
              "type": "object",
              "properties": {
                "total_requests": {"type": "integer", "example": 100},
                "total_tokens": {"type": "integer", "example": 50000},
                "total_cost": {"type": "number", "example": 1.25},
                "current_balance": {"type": "number", "example": 98.75}
              }
            }
          },
          "401": {
            "description": "认证失败",
            "schema": {"$ref": "#/definitions/Error"}
          }
        }
      }
    }
  }
}`

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, swaggerJSON)
}

// getSwaggerCSS 返回基本的CSS样式
func (h *SwaggerHandler) getSwaggerCSS() string {
	return `/* Basic Swagger UI styles */
body { font-family: sans-serif; margin: 0; padding: 20px; }
.swagger-ui { max-width: 1200px; margin: 0 auto; }`
}

// getSwaggerJS 返回基本的JavaScript
func (h *SwaggerHandler) getSwaggerJS() string {
	return `/* Basic Swagger UI JavaScript */
console.log('Swagger UI loaded');`
}
