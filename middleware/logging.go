package middleware

import (
	"log"
	"net/http"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "Turl: ", 0)

type LogResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *LogResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *LogResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	return size, err
}

func (rw *LogResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.SetFlags(log.Ldate | log.Ltime)
		logger.Print(r.Method, r.URL.Path)
		startTime := time.Now()
		lrw := LogResponseWriter{rw, http.StatusOK}
		h.ServeHTTP(&lrw, r)
		endTime := time.Now()
		logger.SetFlags(0)
		logger.Println("status", lrw.StatusCode, "response time: ", endTime.Sub(startTime))
	})

}
