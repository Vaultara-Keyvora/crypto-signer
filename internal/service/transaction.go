package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/pkg/common/apistruct"
	"lbe_crypto_signer/pkg/tools/log"
)

// Transaction Sign
func (l *LBESignerT) Sign(ctx *gin.Context, req *apistruct.SignerTxReq) (*apistruct.SignerTxResp, error) {
	startedAt := time.Now()
	ctx.Set("chain", req.Chain)
	log.ZInfo(ctx, "transaction signing started",
		"stage", "transaction_signing",
		"result", "started",
		"chain_id", req.ChainID,
		"address_index", req.AddrIndex)
	signRes, err := l.WalletCli.Sign(ctx, req.BaseAddr, req.Chain, req.ChainID, req.SignAddr, req.AddrIndex, req.ToAddr, req.ToAddrIndex, req.ToAddrBasic, req.Amount, req.SignData)
	if err != nil {
		log.ZError(ctx, "transaction signing failed", err,
			"stage", "transaction_signing",
			"result", "failure",
			"error_code", "signing_failed",
			"duration_ms", time.Since(startedAt).Milliseconds())
		return nil, err
	}
	log.ZInfo(ctx, "transaction signing completed",
		"stage", "transaction_signing",
		"result", "success",
		"duration_ms", time.Since(startedAt).Milliseconds())
	return &apistruct.SignerTxResp{SignTx: signRes}, err
}
