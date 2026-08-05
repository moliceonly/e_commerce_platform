package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Register 挂载 Prometheus 指标与 /metrics（阶段 H · 3.4）。
//
// 依赖（实现时）：
//
//	go get github.com/prometheus/client_golang
//
// 建议指标：
//   - http_requests_total{method,path,status}
//   - http_request_duration_seconds
//
// 中间件里 Observe；r.GET("/metrics", promhttp.Handler())

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

func Register(r *gin.Engine) {
	// TODO(H4): 注册 collector + 中间件 + /metrics 路由
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// Middleware 可选：返回 gin 中间件统计延迟。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequests.WithLabelValues(c.Request.Method, path, status).Inc()
		httpDuration.WithLabelValues(c.Request.Method, path).Observe(float64(time.Since(start).Seconds()))
	}
}
