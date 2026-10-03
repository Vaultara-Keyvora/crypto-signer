package api

import (
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/version"
	"net/http"
)

type MessageHandler struct {
}

func NewMessageHandler() *MessageHandler { return &MessageHandler{} }

func (f *MessageHandler) Ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (f *MessageHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.Version,
	})
}
