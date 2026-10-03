package middleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apiresp"
	"lbe_crypto_signer/pkg/consts"
	"lbe_crypto_signer/pkg/signx"

	"github.com/gin-gonic/gin"
)

const maxJSONRequestBodyBytes int64 = 2 << 20

func Params(config *service.ActiveConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// beginTime := time.Now()
		var body bytes.Buffer
		// 保存原始的 gin.ResponseWriter
		originalWriter := ctx.Writer
		// 下面是响应参数缓存，传递 gin.Context 以便在 Write 时用
		ctx.Writer = &bodyParamsWriter{
			body:           &body,
			ResponseWriter: originalWriter,
			ctx:            ctx,
			config:         config,
		}

		// reqBody := "非json不保存Body"
		// 请求参数
		reqContentType := ctx.Request.Header.Get("Content-Type")
		if strings.Contains(reqContentType, "application/json") {
			reader := http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONRequestBodyBytes)
			body, err := io.ReadAll(reader)
			if err != nil {
				status := http.StatusBadRequest
				message := "invalid request body"
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					status = http.StatusRequestEntityTooLarge
					message = "request body too large"
					ctx.Set("error_code", "request_body_too_large")
				} else {
					ctx.Set("error_code", "invalid_request_body")
				}
				ctx.Set("result", "failure")
				ctx.AbortWithStatusJSON(status, apiresp.ApiResponse{Code: status, Msg: message})
				return
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			// reqBody = string(body)
			ctx.Set("user_req_params", string(body))
		}

		// rd := merchant_model.Request{
		// 	Method:   ctx.Request.Method,
		// 	Path:     ctx.Request.RequestURI,
		// 	ClientIP: ctx.ClientIP(),
		// 	Body:     reqBody,
		// 	Header:   ctx.Request.Header,
		// }

		ctx.Next()

		// 日志和异步保存
		// resp := body.String()
		// rsd := merchant_model.Response{
		// 	HttpCode: ctx.Writer.Status(),
		// 	Body:     resp,
		// 	Header:   ctx.Writer.Header(),
		// }
		// endTime := time.Now().UnixMilli()

	}
}

// 获取原报文
func GetReqParams(ctx context.Context) (string, bool) {
	reqParams, ok := ctx.Value("user_req_params").(string)
	return reqParams, ok
}

type bodyParamsWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	ctx    *gin.Context
	config *service.ActiveConfig
}

// 真正输出的时候，它会调用这个输出方法
func (w bodyParamsWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	vd, err := signx.NewSignV2().CreateSign(w.ctx, w.config.SignerConf.SecretKey, w.ctx.Request.RequestURI, consts.SignerAppId, w.config.SignerConf.MerchantId, w.ctx.Request.Method, string(b))
	if err == nil {
		w.ctx.Header("Authorization", vd.Authorization)
	}

	return w.ResponseWriter.Write(b)
}
