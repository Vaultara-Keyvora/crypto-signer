package wallet

import (
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apistruct"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// login
// @Tags		login for local administrator
// @Summary	    etc
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.AdminLoginReq   true	"request parameter"
// @Success	200	{object}	apistruct.AdminLoginResp  "request response"
// @Router		/lbe-signer-api/user/login [post]
func (a *AdminHandler) Login(ctx *gin.Context) {
	handle[apistruct.AdminLoginReq, apistruct.AdminLoginResp](service.Core.Login, ctx)
}

// login
// @Tags		login out
// @Summary	    etc
// @Accept		application/json
// @Produce	application/json
// @Param		{object}	body apistruct.AdminLoginOutReq   true	"request parameter"
// @Success	200	{object}	apistruct.AdminLoginOutResp  "request response"
// @Router		/lbe-signer-api/user/loginOut [post]
func (a *AdminHandler) LoginOut(ctx *gin.Context) {
	handle[apistruct.AdminLoginOutReq, apistruct.AdminLoginOutResp](service.Core.LoginOut, ctx)
}
