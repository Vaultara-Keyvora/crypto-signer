package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"lbe_crypto_signer/pkg/tools/log"
)

func Post(ctx context.Context, targetURL string, reqBody []byte, ops ...ClientOptions) ([]byte, error) {
	body, _, _, err := post(ctx, targetURL, reqBody, ops...)
	return body, err
}

func PostWithHeader(ctx context.Context, targetURL string, reqBody []byte, ops ...ClientOptions) (respBody []byte, respHeader http.Header, respCode int, err error) {
	return post(ctx, targetURL, reqBody, ops...)
}

func post(ctx context.Context, targetURL string, reqBody []byte, ops ...ClientOptions) (respBody []byte, respHeader http.Header, respCode int, err error) {
	def := defaultClientOptions()
	for _, op := range ops {
		op(def)
	}

	startedAt := time.Now()
	host := safeHost(targetURL)
	logFields := []any{"stage", "outbound_http", "result", "started", "http_method", http.MethodPost, "host", host, "request_bytes", len(reqBody)}
	if !def.ignore_log {
		log.ZInfo(ctx, "outbound request started", logFields...)
	}

	ctx, cancel := context.WithTimeout(ctx, def.timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(reqBody))
	if err != nil {
		log.ZError(ctx, "outbound request could not be created", nil,
			"stage", "outbound_http", "result", "failure", "error_code", "request_creation_failed",
			"host", host, "duration_ms", time.Since(startedAt).Milliseconds())
		return nil, nil, 0, err
	}

	for key, val := range def.header {
		request.Header.Add(key, val)
	}
	request.Header.Set("Content-Type", "application/json")
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		request.Header.Set("trace_id", traceID)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.ZError(ctx, "outbound request failed", nil,
			"stage", "outbound_http", "result", "failure", "error_code", "request_failed",
			"host", host, "duration_ms", time.Since(startedAt).Milliseconds())
		return nil, nil, 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.ZError(ctx, "outbound request returned an unsuccessful status", nil,
			"stage", "outbound_http", "result", "failure", "error_code", "upstream_status_error",
			"host", host, "http_status", response.StatusCode, "duration_ms", time.Since(startedAt).Milliseconds())
		return nil, nil, response.StatusCode, fmt.Errorf("http request failed with status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.ZError(ctx, "outbound response could not be read", nil,
			"stage", "outbound_http", "result", "failure", "error_code", "response_read_failed",
			"host", host, "http_status", response.StatusCode, "duration_ms", time.Since(startedAt).Milliseconds())
		return nil, nil, response.StatusCode, err
	}

	if !def.ignore_log {
		log.ZInfo(ctx, "outbound request completed",
			"stage", "outbound_http", "result", "success", "host", host,
			"http_status", response.StatusCode, "response_bytes", len(body),
			"duration_ms", time.Since(startedAt).Milliseconds())
	}
	return body, response.Header, response.StatusCode, nil
}

func safeHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}
