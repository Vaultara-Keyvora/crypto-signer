package wallet

import (
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apistruct"
)

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

// Sign
// @Tags		Sign for
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.SignerTxReq   true	"request parameter"
// @Success	200	{object}	apistruct.SignerTxResp  "request response"
// @Router		/lbe-signer-api/tx/sign [post]
func (a *TransactionHandler) Sign(ctx *gin.Context) {
	handle[apistruct.SignerTxReq, apistruct.SignerTxResp](service.Core.Sign, ctx)
}
