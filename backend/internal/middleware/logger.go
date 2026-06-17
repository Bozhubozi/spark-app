package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger replaces Gin's default logger with request ID and structured fields.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		rid, _ := c.Get("request_id")

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		fmt.Printf("[spark] %s | %3d | %8v | %15s | %-6s %s | rid=%v\n",
			time.Now().Format("2006-01-02 15:04:05"),
			status,
			latency.Truncate(time.Microsecond),
			clientIP,
			method,
			path,
			rid,
		)
	}
}
