// Package middleware 提供通用 Gin 中间件。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

// Recovery 捕获 panic 并返回标准错误响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				pkglogger.L().Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":       errs.ErrInternal.Code,
					"message":    errs.ErrInternal.Message,
					"request_id": c.GetString("request_id"),
					"data":       nil,
				})
			}
		}()
		c.Next()
	}
}
