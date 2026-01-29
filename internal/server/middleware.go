package server

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func remoteIP(r *http.Request) string {
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	if xc := r.Header.Get("X-Client-IP"); xc != "" {
		return xc
	}
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.SplitN(xf, ",", 1)
		if len(parts) == 1 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		r := c.Request
		ip := remoteIP(r)
		method := r.Method
		path := r.URL.RequestURI()
		status := c.Writer.Status()
		log.Debug().Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("latency", latency.String()).
			Str("IP", ip)
	}
}
