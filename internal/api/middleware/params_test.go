package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/config"

	"github.com/gin-gonic/gin"
)

func TestParamsRejectsOversizedJSONBeforeNextHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	activeConfig := &service.ActiveConfig{SignerConf: &config.Signer{}}
	called := false
	router := gin.New()
	router.Use(Params(activeConfig))
	router.POST("/sign", func(ctx *gin.Context) {
		called = true
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(strings.Repeat("x", int(maxJSONRequestBodyBytes+1))))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
	if called {
		t.Fatal("next handler ran for an oversized body")
	}
}

func TestParamsRestoresBodyForNextHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	activeConfig := &service.ActiveConfig{SignerConf: &config.Signer{}}
	var received string
	router := gin.New()
	router.Use(Params(activeConfig))
	router.POST("/sign", func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		received = string(body)
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(`{"amount":1}`))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if received != `{"amount":1}` {
		t.Fatalf("handler received body %q", received)
	}
}
