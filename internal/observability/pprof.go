package observability

import (
	"net/http/pprof"

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

	r.GET("/debug/pprof/", gin.WrapF(pprof.Index))

	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))

	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))

	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))

	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))

	r.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))

	r.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))

	r.GET("/debug/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))

	r.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))

	r.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))

	r.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))

}
