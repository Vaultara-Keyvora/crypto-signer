package logic

import "time"

const (
	FlagEthereumAccount = "account"
	FlagMerchantId      = "merchant-id"
	FlagMerchantUrl     = "merchant-url"
	FlagSignerUrl       = "signer-url"
	FlagSecretKey       = "secret-key"

	ConfigFilePath = "./config/signer.yaml"

	URISignerPing   = "/signer/ping"
	URISignerConfig = "/signer/config"
	URISignerReport = "/signer/report"

	MerchantSignerConfigValidTime = time.Second * 10
)

var FlagUsageMap = map[string]string{
	FlagEthereumAccount: "Ethereum account, security authentication for signer config page",
	FlagMerchantId:      "Merchant id , use to communicate with merchant platform",
	FlagMerchantUrl:     "Merchant platform url",
	FlagSignerUrl:       "Signer service url",
	FlagSecretKey:       "Secret key, security authentication for communication with merchant platform",
}

var FlagShorthandMap = map[string]string{
	FlagEthereumAccount: "a",
}

var FlagConfigKeyMap = map[string]string{
	FlagEthereumAccount: "ethereumAccount",
	FlagMerchantId:      "merchantId",
	FlagMerchantUrl:     "merchantUrl",
	FlagSignerUrl:       "signerUrl",
	FlagSecretKey:       "secretKey",
}

type ReportReq struct {
	MerchantId      string `json:"merchant_id"`      // merchant id
	EthereumAccount string `json:"ethereum_account"` // ethereum account, security authentication for signer config page
	SignerUrl       string `json:"signer_url"`       // signer url
}

type ReportResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type MerchantSignerConfigReq struct {
	MerchantId string `json:"merchant_id"` // merchant id
}

type MerchantSignerConfigResp struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data MerchantSignerConfig `json:"data,omitempty"`
}

type MerchantSignerConfig struct {
	EnableConfigPage bool     `json:"enable_config_page"` // status of signer config page: true-Enable  false-Disable
	IPWhiteList      []string `json:"ip_white_list"`      // ip white list of signer config page
}
