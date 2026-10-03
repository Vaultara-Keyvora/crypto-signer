package api

import (
	"fmt"
	"io/fs"
	"lbe_crypto_signer/internal/api/middleware"
	"lbe_crypto_signer/internal/api/wallet"
	"lbe_crypto_signer/internal/logic"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/common/apiresp"
	"lbe_crypto_signer/pkg/tools/log"
	"lbe_crypto_signer/static"
	"net/http"

	_ "lbe_crypto_signer/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// swag init -g .\cmd\active\main.go -o ./docs --parseDependency --parseInternal

const (
	PrefixRouter = "/lbe-signer-api"
)

func NewGinRouter(config *service.ActiveConfig) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestContext(), gin.Recovery(), middleware.CorsMiddleware())
	// Trust no proxies: the client IP must come from the peer socket, not
	// client-supplied X-Forwarded-For / X-Real-IP headers.
	_ = r.SetTrustedProxies(config.ApiConf.Api.TrustedProxy)

	service.Init(config)

	ConfigPageRouter(r)

	api := r.Group(PrefixRouter)
	{
		if config.ApiConf.Api.Swagger {
			swagHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
			if swagHandler != nil {
				api.GET("/doc/swagger/*any", swagHandler)
			}
		}

		f := NewMessageHandler()
		api.HEAD("/ping", f.Ping)
		api.GET("/version", f.Version)

		api.Use(middleware.Params(config), middleware.VerifyToken(config))

		api.GET("/version_with_sign", f.Version)

		keysApi := wallet.NewKeysHandler()
		accountRouter := api.Group("/key")
		{
			accountRouter.POST("/list", keysApi.KeysList)
			accountRouter.POST("/addAccounts", keysApi.AddAccounts)
			accountRouter.POST("/listDepositAddr", keysApi.ListDepositAddr)
			accountRouter.POST("/checkAccounts", keysApi.CheckAccounts)
		}
		txApi := wallet.NewTransactionHandler()
		txRouter := api.Group("/tx")
		{
			txRouter.POST("/sign", txApi.Sign)
		}
	}

	return r
}

func localWebInit(r *gin.Engine) *gin.RouterGroup {
	distFS, err := fs.Sub(static.Static, "web")
	if err != nil {
		panic(err)
	}
	configGroup := r.Group("/local", logic.SignerAuthIns.ConfigPageAuth)
	{
		configGroup.StaticFS("/config", http.FS(distFS))
		configGroup.StaticFileFS("/favicon.ico", "/web/favicon.ico", http.FS(distFS))
	}

	return configGroup
}

func ConfigPageRouter(r *gin.Engine) {
	// local web service init
	group := localWebInit(r)

	group.Use(gin.Recovery(), middleware.CorsMiddleware())
	api := group.Group(PrefixRouter)
	{
		f := NewMessageHandler()
		api.HEAD("/ping", f.Ping)
		api.GET("/version", f.Version)

		loginApi := wallet.NewAdminHandler()
		userRouter := api.Group("/user")
		{
			userRouter.POST("/login", loginApi.Login)
			userRouter.POST("/loginOut", loginApi.LoginOut)
		}

		keysApi := wallet.NewKeysHandler()
		accountRouter := api.Group("/key", ConfigPageTokenAuth)
		{
			accountRouter.POST("/addMnemonic", keysApi.KeysAddMnemonic)
			accountRouter.POST("/list", keysApi.KeysList)
			accountRouter.POST("/addAccounts", keysApi.AddAccounts)
			accountRouter.POST("/addDepositAddr", keysApi.AddDepositAddr)
			accountRouter.POST("/listDepositAddr", keysApi.ListDepositAddr)
		}
		txApi := wallet.NewTransactionHandler()
		txRouter := api.Group("/tx", ConfigPageTokenAuth)
		{
			txRouter.POST("/sign", txApi.Sign)
		}
	}

}

func ConfigPageTokenAuth(c *gin.Context) {
	token := c.GetHeader("Token")
	if token == "" {
		apiresp.ApiError(c, fmt.Errorf("token is required"))
		c.Abort()
		return
	}

	log.ZInfo(c, "configuration page token authentication started", "stage", "token_authentication")

	err := service.Core.TokenAuth(c, token)
	if err != nil {
		log.ZError(c, "configuration page token authentication failed", err, "stage", "token_authentication", "result", "failure")
		apiresp.ApiError(c, err)
		c.Abort()
		return
	}
	c.Next()
}
