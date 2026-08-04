package observability

import (
	"net/http"
	_ "net/http/pprof" // 注册默认 mux 上的 pprof；也可挂到 gin

	"github.com/gin-gonic/gin"
)

// MountPprof 仅非 prod 暴露 /debug/pprof/*（阶段 H · 3.4）。
// 注意：生产公网不要裸奔 pprof。
func MountPprof(r *gin.Engine, enabled bool) {
	if !enabled {
		return
	}
	// TODO(H4): 将 pprof 挂到 gin，例如：
	//  r.GET("/debug/pprof/", gin.WrapH(http.DefaultServeMux))
	//  或对 /debug/pprof/* 写一组 WrapF(pprof.Index) 等
	_ = r
	_ = http.DefaultServeMux
}
