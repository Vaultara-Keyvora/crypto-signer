package httpx

import (
	"maps"
	"time"
)

type clientOptions struct {
	timeout time.Duration
	ignore_log bool
	header  map[string]string
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		timeout: 30 * time.Second,
		ignore_log: false,
		header:  make(map[string]string),
	}
}

type ClientOptions func(*clientOptions)

func WithTimeout(timeout time.Duration) ClientOptions {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

func WithHeaderKV(key, value string) ClientOptions {
	return func(o *clientOptions) {
		if o.header == nil {
			o.header = make(map[string]string)
		}
		o.header[key] = value
	}
}

func WithHeaders(headers map[string]string) ClientOptions {
	return func(o *clientOptions) {
		if o.header == nil {
			o.header = make(map[string]string)
		}
		maps.Copy(o.header, headers)
	}
}

// 忽略打印请求和响应参数
func WithIgnoreLog() ClientOptions {
	return func(o *clientOptions) {
		o.ignore_log = true
	}
}
