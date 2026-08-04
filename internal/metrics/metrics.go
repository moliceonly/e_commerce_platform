package metrics

import "github.com/gin-gonic/gin"

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
func Register(r *gin.Engine) {
	// TODO(H4): 注册 collector + 中间件 + /metrics 路由
	_ = r
}

// Middleware 可选：返回 gin 中间件统计延迟。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO(H4)
		c.Next()
	}
}
