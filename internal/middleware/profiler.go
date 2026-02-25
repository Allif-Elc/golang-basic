package middleware

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

// Profiler returns a chi sub-router that mounts pprof endpoints
// Usage: r.Mount("/debug", middleware.Profiler())
func Profiler() http.Handler {
	r := chi.NewRouter()

	// pprof HTTP handlers
	r.HandleFunc("/", pprof.Index)
	r.HandleFunc("/cmdline", pprof.Cmdline)
	r.HandleFunc("/profile", pprof.Profile)
	r.HandleFunc("/symbol", pprof.Symbol)
	r.HandleFunc("/trace", pprof.Trace)

	// Manually add support for profiles accessed via /debug/pprof/heap, /debug/pprof/goroutine, etc.
	r.HandleFunc("/heap", pprof.Handler("heap").ServeHTTP)
	r.HandleFunc("/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.HandleFunc("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	r.HandleFunc("/block", pprof.Handler("block").ServeHTTP)
	r.HandleFunc("/mutex", pprof.Handler("mutex").ServeHTTP)
	r.HandleFunc("/allocs", pprof.Handler("allocs").ServeHTTP)

	// expvar handler
	r.Handle("/vars", expvar.Handler())

	return r
}
