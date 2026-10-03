package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"lbe_crypto_signer/pkg/tools/log"
)

// RequestContext assigns a server-generated ID and attaches safe request
// metadata to every log emitted while handling the request.
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		requestID, err := uuid.NewV7()
		if err != nil {
			requestID = uuid.New()
		}
		traceID := requestID.String()
		operation := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			operation = c.Request.Method + " unmatched"
		}

		c.Set("trace_id", traceID)
		c.Set("operation", operation)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "trace_id", traceID))
		c.Header("X-Request-ID", traceID)

		c.Next()

		result := "success"
		if c.Writer.Status() >= http.StatusBadRequest {
			result = "failure"
		}
		if responseResult, ok := c.Get("result"); ok {
			if value, ok := responseResult.(string); ok && value != "" {
				result = value
			}
		}

		fields := []any{
			"stage", "request_complete",
			"result", result,
			"http_method", c.Request.Method,
			"http_status", c.Writer.Status(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
		}
		if errorCode, ok := c.Get("error_code"); ok {
			fields = append(fields, "error_code", errorCode)
		}
		log.ZInfo(c, "request completed", fields...)
	}
}
