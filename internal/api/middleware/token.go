package middleware

import (
	"context"
	"errors"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apiresp"
	"lbe_crypto_signer/pkg/signx"
	"lbe_crypto_signer/pkg/tools/log"

	"github.com/gin-gonic/gin"
)

// 接口验签
func VerifyToken(config *service.ActiveConfig) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		// 获取原文
		oriReq, ok := GetReqParams(ctx)
		if ctx.Request.Method == "POST" && !ok {
			log.ZWarn(ctx, "signature verification request body is unavailable", nil, "stage", "signature_verification", "result", "failure", "error_code", "request_body_unavailable")
			apiresp.ApiError(ctx, errors.New("获取请求参数失败"))
			ctx.Abort()
			return
		}

		// 获取授权信息
		authorization := ctx.Request.Header.Get("authorization")
		if authorization == "" {
			log.ZWarn(ctx, "signature verification authorization is missing", nil, "stage", "signature_verification", "result", "failure", "error_code", "authorization_missing")
			apiresp.ApiError(ctx, errors.New("获取授权信息失败"))
			ctx.Abort()
			return
		}

		// Do not log Authorization, request bodies, or RequestURI (which may contain query data).
		log.ZInfo(ctx, "signature verification started", "stage", "signature_verification", "result", "started")

		// 构建签名结构体
		err := buildAndVerifySign(ctx, config, authorization, ctx.Request.Method, ctx.Request.RequestURI, oriReq)
		if err != nil {
			apiresp.ApiError(ctx, err)
			ctx.Abort()
			return
		}

		ctx.Next()

	}

}

func GetOpenApiAuthInfo(ctx context.Context) (*signx.AuthInfo, bool) {
	sginfo := ctx.Value("OPENAPI_AUTH_INFO")
	if sginfo == nil {
		return nil, false
	}

	s, ok := sginfo.(*signx.AuthInfo)
	if !ok {
		return nil, false
	}

	return s, true
}

func GetMerchantId(ctx context.Context) (string, bool) {
	authInfo, ok := GetOpenApiAuthInfo(ctx)
	if !ok {
		return "", false
	}

	return authInfo.MerchantId, true
}

func GetAppid(ctx context.Context) (string, bool) {
	authInfo, ok := GetOpenApiAuthInfo(ctx)
	if !ok {
		return "", false
	}

	return authInfo.Appid, true
}

func buildAndVerifySign(ctx *gin.Context, config *service.ActiveConfig, auth string, method string, uri string, body string) error {
	ctx.Set("OPENAPI_AUTH_AUTHORIZATION", auth)

	sign := signx.NewSignV2()
	authInfo, err := sign.ParseAuthStr(ctx, auth)
	if err != nil {
		log.ZWarn(ctx, "signature authorization could not be parsed", nil, "stage", "signature_verification", "result", "failure", "error_code", "invalid_authorization")
		return err
	}
	// Bind the claimed identity to the configured merchant so the shared
	// secret cannot be presented under an arbitrary merchant_id/appid.
	if authInfo.MerchantId != config.SignerConf.MerchantId {
		log.ZWarn(ctx, "signature merchant does not match configured merchant", nil,
			"stage", "signature_verification", "result", "failure", "error_code", "merchant_mismatch",
			"claimed_merchant_id", authInfo.MerchantId, "configured_merchant_id", config.SignerConf.MerchantId)
		return errors.New("merchant_id mismatch")
	}

	// 验签
	err = sign.VerifySign(ctx, config.SignerConf.SecretKey, signx.VerifyData{
		Authorization: auth,
		SignInfo: signx.SignInfo{
			Method:       method,
			CanonicalURL: uri,
			SignBody:     body,
		},
	})
	if err != nil {
		log.ZWarn(ctx, "request signature verification failed", nil, "stage", "signature_verification", "result", "failure", "error_code", "signature_invalid")
		return errors.New("签名验证失败")
	}

	// 存储签名信息
	ctx.Set("OPENAPI_AUTH_INFO", authInfo)
	ctx.Set("merchant_id", authInfo.MerchantId)

	return nil
}

func GetAuthorization(ctx context.Context) (string, bool) {
	auth := ctx.Value("OPENAPI_AUTH_AUTHORIZATION")
	if auth == nil {
		return "", false
	}
	s, ok := auth.(string)
	if !ok {
		return "", false
	}
	return s, true
}
