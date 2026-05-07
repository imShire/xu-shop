package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestTotal HTTP 请求总数
	HTTPRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "xushop_http_requests_total",
		Help: "HTTP requests total",
	}, []string{"method", "path", "status"})

	// HTTPRequestDuration HTTP 请求延迟
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "xushop_http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"method", "path"})

	// OrderCreatedTotal 下单总数
	OrderCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "xushop_orders_created_total",
		Help: "Orders created total",
	})

	// PaymentSuccessTotal 支付成功总数
	PaymentSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "xushop_payments_success_total",
		Help: "Payments success total",
	})

	// StockLockFailedTotal 库存锁定失败次数
	StockLockFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "xushop_stock_lock_failed_total",
		Help: "Stock lock failed total",
	})
)

// PrometheusMiddleware gin 中间件，记录请求指标。
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timer := prometheus.NewTimer(HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()))
		c.Next()
		timer.ObserveDuration()
		HTTPRequestTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			fmt.Sprintf("%d", c.Writer.Status()),
		).Inc()
	}
}
