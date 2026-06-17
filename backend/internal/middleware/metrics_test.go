package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMetricsCollect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewMetrics()

	r := gin.New()
	r.Use(m.Collect())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/error", func(c *gin.Context) { c.JSON(500, gin.H{"err": true}) })

	// Make some requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if m.totalRequests.Load() != 4 {
		t.Errorf("total: %d", m.totalRequests.Load())
	}
	if m.statusCodes[0].Load() != 3 {
		t.Errorf("2xx: %d", m.statusCodes[0].Load())
	}
	if m.statusCodes[5].Load() != 1 {
		t.Errorf("5xx: %d", m.statusCodes[5].Load())
	}
}

func TestMetricsOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewMetrics()

	r := gin.New()
	r.Use(m.Collect())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/metrics", m.Handler())

	// Make a request to generate metrics
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Get metrics
	req2 := httptest.NewRequest("GET", "/metrics", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Fatalf("metrics: %d", w2.Code)
	}

	body := w2.Body.String()
	for _, expected := range []string{
		"spark_http_requests_total", "spark_http_requests_active",
		"spark_request_duration_microseconds_total", "spark_http_responses_total",
		"spark_uptime_seconds",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("missing metric: %s", expected)
		}
	}
}
