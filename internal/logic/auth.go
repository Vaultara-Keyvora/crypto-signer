package logic

import (
	"context"
	"fmt"
	"lbe_crypto_signer/pkg/common/apiresp"
	"lbe_crypto_signer/pkg/common/config"
	"lbe_crypto_signer/pkg/tools/log"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var SignerAuthIns *SignerAuth

type SignerAuth struct {
	cfg                  *config.Signer
	merchantSignerConfig *MerchantSignerConfig
	lastUpdatedAt        int64
	lock                 sync.Mutex
}

func NewSignerAuth(cfg *config.Signer) *SignerAuth {
	return &SignerAuth{
		cfg: cfg,
	}
}

func InitSignerAuth(cfg *config.Signer) (*SignerAuth, error) {
	ctx := context.Background()
	err := SignerInitIns.checkParam(ctx, cfg)
	if err != nil {
		log.ZError(ctx, "signer authentication configuration is invalid", err, "stage", "configuration_validation", "result", "failure", "error_code", "invalid_configuration", "merchant_id", cfg.MerchantId)
		return nil, err
	}

	SignerAuthIns = NewSignerAuth(cfg)
	return SignerAuthIns, nil
}

func (s *SignerAuth) ConfigPageAuth(c *gin.Context) {
	c.Set("merchant_id", s.cfg.MerchantId)
	signerConfig, err := s.getMerchantSignerConfig(c, s.cfg)
	if err != nil {
		log.ZError(c, "merchant signer configuration lookup failed", nil, "stage", "merchant_configuration", "result", "failure", "error_code", "merchant_configuration_unavailable")
		apiresp.ApiError(c, fmt.Errorf("fail to get signerConfig error:%v", err))
		c.Abort()
		return
	}

	// check if ip is in white list
	ip := c.ClientIP()
	log.ZInfo(c, "configuration page access checked", "stage", "configuration_page_authorization", "client_ip", ip)
	if ok := s.checkIP(c, ip, signerConfig.IPWhiteList); !ok {
		apiresp.ApiError(c, fmt.Errorf("ip:%v not in white list", ip))
		c.Abort()
		return
	}

	// check if config is enabled
	if !signerConfig.EnableConfigPage {
		apiresp.ApiError(c, fmt.Errorf("signer config page not enabled"))
		c.Abort()
		return
	}

	c.Next()
}

func (s *SignerAuth) getMerchantSignerConfig(ctx context.Context, param *config.Signer) (*MerchantSignerConfig, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.isNeedUpdateMerchantSignerConfig(ctx) {
		log.ZInfo(ctx, "merchant signer configuration cache hit", "stage", "merchant_configuration", "result", "success",
			"config_page_enabled", s.merchantSignerConfig.EnableConfigPage, "ip_whitelist_count", len(s.merchantSignerConfig.IPWhiteList))
		return s.merchantSignerConfig, nil
	}

	cfgURL, err := url.JoinPath(param.MerchantUrl, URISignerConfig)
	if err != nil {
		log.ZError(ctx, "merchant endpoint configuration is invalid", nil, "stage", "merchant_configuration", "result", "failure", "error_code", "invalid_endpoint")
		return nil, err
	}

	req := &MerchantSignerConfigReq{
		MerchantId: param.MerchantId,
	}
	resp := &MerchantSignerConfigResp{}

	err = HttpPost(ctx, cfgURL, param, req, resp)
	if err != nil {
		log.ZError(ctx, "merchant signer configuration request failed", nil, "stage", "merchant_configuration", "result", "failure", "error_code", "merchant_request_failed")
		return nil, err
	}

	log.ZInfo(ctx, "merchant signer configuration received", "stage", "merchant_configuration", "result", "success",
		"config_page_enabled", resp.Data.EnableConfigPage, "ip_whitelist_count", len(resp.Data.IPWhiteList))
	s.refreshMerchantSignerConfig(ctx, &resp.Data)

	return &resp.Data, nil
}

func (s *SignerAuth) isNeedUpdateMerchantSignerConfig(ctx context.Context) bool {
	if s.merchantSignerConfig == nil {
		return true
	}
	if s.lastUpdatedAt < time.Now().Unix() {
		return true
	}
	return false
}

func (s *SignerAuth) refreshMerchantSignerConfig(ctx context.Context, c *MerchantSignerConfig) {
	s.merchantSignerConfig = c
	s.lastUpdatedAt = time.Now().Unix() + int64(MerchantSignerConfigValidTime.Seconds())
}

func (s *SignerAuth) checkIP(c *gin.Context, ip string, whiteList []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}
	for _, entry := range whiteList {
		if _, cidr, err := net.ParseCIDR(entry); err == nil {
			if cidr.Contains(clientIP) {
				return true
			}
			continue
		}
		if entryIP := net.ParseIP(entry); entryIP != nil && entryIP.Equal(clientIP) {
			return true
		}
	}
	return false
}
