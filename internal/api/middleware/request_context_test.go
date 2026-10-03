package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequestContextAddsTraceAndOperation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestContext())
	router.POST("/tx/sign", func(c *gin.Context) {
		traceID, _ := c.Value("trace_id").(string)
		operation, _ := c.Value("operation").(string)
		if traceID == "" || operation != "POST /tx/sign" {
			t.Errorf("request context fields = (%q, %q), want trace ID and POST /tx/sign", traceID, operation)
		}
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/tx/sign", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if _, err := uuid.Parse(response.Header().Get("X-Request-ID")); err != nil {
		t.Fatalf("X-Request-ID is not a UUID: %v", err)
	}
}
