package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics provides lightweight Prometheus-compatible request metrics.
type Metrics struct {
	totalRequests   atomic.Int64
	activeRequests  atomic.Int64
	requestDuration atomic.Int64 // cumulative microseconds
	statusCodes     [6]atomic.Int64 // 2xx, 3xx, 4xx, 401, 403, 5xx
	startTime       time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{startTime: time.Now()}
}

func (m *Metrics) Collect() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.totalRequests.Add(1)
		m.activeRequests.Add(1)
		start := time.Now()

		c.Next()

		m.activeRequests.Add(-1)
		dur := time.Since(start).Microseconds()
		m.requestDuration.Add(dur)

		code := c.Writer.Status()
		switch {
		case code == 401:
			m.statusCodes[3].Add(1)
		case code == 403:
			m.statusCodes[4].Add(1)
		case code >= 500:
			m.statusCodes[5].Add(1)
		case code >= 400:
			m.statusCodes[2].Add(1)
		case code >= 300:
			m.statusCodes[1].Add(1)
		default:
			m.statusCodes[0].Add(1)
		}
	}
}

func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		total := m.totalRequests.Load()
		active := m.activeRequests.Load()
		dur := m.requestDuration.Load()
		uptime := time.Since(m.startTime).Seconds()

		output := strings.Builder{}
		output.WriteString("# HELP spark_http_requests_total Total HTTP requests.\n")
		output.WriteString("# TYPE spark_http_requests_total counter\n")
		output.WriteString(fmt.Sprintf("spark_http_requests_total %d\n", total))

		output.WriteString("# HELP spark_http_requests_active Currently active requests.\n")
		output.WriteString("# TYPE spark_http_requests_active gauge\n")
		output.WriteString(fmt.Sprintf("spark_http_requests_active %d\n", active))

		output.WriteString("# HELP spark_request_duration_microseconds_total Cumulative request duration.\n")
		output.WriteString("# TYPE spark_request_duration_microseconds_total counter\n")
		output.WriteString(fmt.Sprintf("spark_request_duration_microseconds_total %d\n", dur))

		output.WriteString("# HELP spark_http_responses_total HTTP responses by status class.\n")
		output.WriteString("# TYPE spark_http_responses_total counter\n")
		for i, label := range []string{"2xx", "3xx", "4xx", "401", "403", "5xx"} {
			output.WriteString(fmt.Sprintf("spark_http_responses_total{status=\"%s\"} %d\n", label, m.statusCodes[i].Load()))
		}

		output.WriteString("# HELP spark_uptime_seconds Process uptime in seconds.\n")
		output.WriteString("# TYPE spark_uptime_seconds gauge\n")
		output.WriteString(fmt.Sprintf("spark_uptime_seconds %.0f\n", uptime))

		c.String(http.StatusOK, output.String())
	}
}
