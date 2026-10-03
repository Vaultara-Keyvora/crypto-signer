package wallet

import (
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/pkg/common/apiresp"
	"lbe_crypto_signer/pkg/common/apistruct"
	"lbe_crypto_signer/pkg/common/errs"
)

func handle[A, B any](fn func(ctx *gin.Context, req *A) (*B, error), c *gin.Context) {
	var req A
	if err := c.ShouldBind(&req); err != nil {
		apiresp.ApiError(c, errs.NewCodeError(errs.ArgsError, err.Error()))
		return
	}

	if err := apistruct.Validate(&req); err != nil {
		apiresp.ApiError(c, err)
		return
	}

	resp, err := fn(c, &req)
	if err != nil {
		apiresp.ApiError(c, err)
		return
	}
	apiresp.ApiSuccess(c, resp)
}
