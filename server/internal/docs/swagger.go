// Package docs 提供 Swagger UI 与 OpenAPI 规范文件服务。
//
// 访问 /docs          → Swagger UI 交互式调试界面
// 访问 /openapi.yaml  → 原始 OpenAPI 3.1 规范（可导入 Postman / Apifox）
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openapiYAML []byte

// swaggerHTML 内嵌 Swagger UI 页面，通过 CDN 加载 swagger-ui-dist。
// 页面启动时自动读取同域 /openapi.yaml，无需额外静态资源。
const swaggerHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>xu-shop API 文档</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; }
    .topbar { background: #1a1a2e !important; }
    .topbar-wrapper img { content: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 30"><text y="24" font-size="20" fill="white" font-weight="bold">xu-shop</text></svg>'); height: 28px; }
    .swagger-ui .info .title small.version-stamp { background: #e74c3c; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        deepLinking: true,
        displayRequestDuration: true,
        filter: true,
        tryItOutEnabled: true,
        requestInterceptor: function(req) {
          // 从 localStorage 读取 token（调试用）
          var token = localStorage.getItem('xu_shop_token');
          if (token) {
            req.headers['Authorization'] = 'Bearer ' + token;
          }
          return req;
        },
        onComplete: function() {
          // 自动用 localStorage 里存的 token 授权
          var token = localStorage.getItem('xu_shop_token');
          if (token) {
            var ui = window.ui;
            if (ui) {
              ui.preauthorizeApiKey('BearerAuth', token);
            }
          }
        }
      });
    };
  </script>
</body>
</html>`

// RegisterRoutes 在 gin.Engine 上注册文档相关路由。
//
//	GET /docs          → Swagger UI
//	GET /openapi.yaml  → 原始 YAML 规范文件（供 Postman / Apifox 导入）
func RegisterRoutes(r *gin.Engine) {
	// Swagger UI 页面
	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
	// 兼容 /docs/ 带尾斜杠
	r.GET("/docs/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	// OpenAPI YAML 原始规范
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapiYAML)
	})

	// JSON 格式（Postman 偏好）— 简单透传 YAML，Postman 也可以直接导入 YAML
	r.GET("/openapi.json", func(c *gin.Context) {
		// Postman / Apifox 支持直接导入 YAML，这里返回 YAML 并设置 json content-type
		// 如需真正 JSON 格式，可接入 yaml-to-json 转换库
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapiYAML)
	})
}
