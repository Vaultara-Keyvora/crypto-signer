package api

import (
	"context"
	"errors"
	"fmt"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/tools/log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func Start(ctx context.Context, apiConfig *service.ActiveConfig) error {

	ctx = context.WithValue(ctx, "operation", "api_server_startup")
	router := NewGinRouter(apiConfig)
	address := net.JoinHostPort(apiConfig.ApiConf.Api.ListenIP, strconv.Itoa(apiConfig.ApiConf.Api.Ports))
	server := http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.ZInfo(ctx, "API server is starting", "stage", "startup", "result", "started", "listen_ip", apiConfig.ApiConf.Api.ListenIP, "listen_port", apiConfig.ApiConf.Api.Ports)
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.ZError(ctx, "API server failed to start", err, "stage", "startup", "result", "failure", "error_code", "api_start_failed")
		netErr := errors.New(fmt.Sprintf("Api start err: %v", err))
		err = server.Shutdown(ctx)
		if err != nil {
			log.ZError(ctx, "API server shutdown failed", err, "stage", "shutdown", "result", "failure")
			return netErr
		}
	}
	return nil
}
