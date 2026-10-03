package wallet

import (
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apistruct"
)

type KeysHandler struct{}

func NewKeysHandler() *KeysHandler {
	return &KeysHandler{}
}

// KeysAddMnemonic
// @Tags		add wallet basic info
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.KeysAddMnemonicReq   true	"request parameter"
// @Success	200	{object}	apistruct.KeysAddMnemonicResp  "request response"
// @Router		/lbe-signer-api/key/addMnemonic [post]
func (a *KeysHandler) KeysAddMnemonic(ctx *gin.Context) {
	handle[apistruct.KeysAddMnemonicReq, apistruct.KeysAddMnemonicResp](service.Core.AddMnemonic, ctx)
}

// KeysList
// @Tags		list key meta info
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.KeysListReq   false	"request parameter"
// @Success	200	{object}	apistruct.KeysListResp  "request response"
// @Router		/lbe-signer-api/key/list [post]
func (a *KeysHandler) KeysList(ctx *gin.Context) {
	handle[apistruct.KeysListReq, apistruct.KeysListResp](service.Core.KeyList, ctx)
}

// AddAccounts
// @Tags		expand accounts by index
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.AddAccountsReq   false	"request parameter"
// @Success	200	{object}	apistruct.AddAccountsResp  "request response"
// @Router		/lbe-signer-api/key/addAccounts [post]
func (a *KeysHandler) AddAccounts(ctx *gin.Context) {
	handle[apistruct.AddAccountsReq, apistruct.AddAccountsResp](service.Core.AddAccounts, ctx)
}

// AddDepositAddr
// @Tags		add deposit address
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.AddDepositAddrReq   false	"request parameter"
// @Success	200	{object}	apistruct.AddDepositAddrResp  "request response"
// @Router		/lbe-signer-api/key/addDepositAddr [post]
func (a *KeysHandler) AddDepositAddr(ctx *gin.Context) {
	handle[apistruct.AddDepositAddrReq, apistruct.AddDepositAddrResp](service.Core.AddDepositAddr, ctx)
}

// ListDepositAddr
// @Tags		list deposit address
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.ListDepositAddrReq   false	"request parameter"
// @Success	200	{object}	apistruct.ListDepositAddrResp  "request response"
// @Router		/lbe-signer-api/key/listDepositAddr [post]
func (a *KeysHandler) ListDepositAddr(ctx *gin.Context) {
	handle[apistruct.ListDepositAddrReq, apistruct.ListDepositAddrResp](service.Core.ListDepositAddr, ctx)
}

// CheckAccounts
// @Tags		check accounts by index
// @Summary	    support eth, tron, bsc, solana
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.CheckAccountsReq   false	"request parameter"
// @Success	200	{object}	apistruct.CheckAccountsResp  "request response"
// @Router		/lbe-signer-api/key/checkAccounts [post]
func (a *KeysHandler) CheckAccounts(ctx *gin.Context) {
	handle[apistruct.CheckAccountsReq, apistruct.CheckAccountsResp](service.Core.CheckAccounts, ctx)
}
