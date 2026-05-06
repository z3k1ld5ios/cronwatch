package api

import (
	"log"
	"net/http"
	"time"
)

// ServeHTTP makes Server implement http.Handler so it can be used
// directly in tests via httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// loggingMiddleware wraps a handler and logs each request's method,
// path, status code, and elapsed time.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.code, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	code int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.code = code
	lrw.ResponseWriter.WriteHeader(code)
}

// WithLogging wraps the server's internal mux with request logging.
func (s *Server) WithLogging() *Server {
	logged := loggingMiddleware(s.mux)
	logged_mux := http.NewServeMux()
	logged_mux.Handle("/", logged)
	s.mux = logged_mux
	return s
}
