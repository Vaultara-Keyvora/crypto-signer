package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"lbe_crypto_signer/pkg/common/config"
	"lbe_crypto_signer/pkg/httpx"
	"lbe_crypto_signer/pkg/signx"
	"lbe_crypto_signer/pkg/tools/log"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var SignerInitIns = NewSignerInit()

type SignerInit struct{}

func NewSignerInit() *SignerInit {
	return &SignerInit{}
}

func (s *SignerInit) Init(ctx context.Context, param *config.Signer) error {
	ctx = context.WithValue(ctx, "operation", "signer_initialization")
	ctx = context.WithValue(ctx, "merchant_id", param.MerchantId)
	log.ZInfo(ctx, "signer initialization started", "stage", "initialization", "result", "started")

	err := s.checkParam(ctx, param)
	if err != nil {
		log.ZError(ctx, "signer initialization configuration is invalid", err, "stage", "configuration_validation", "result", "failure", "error_code", "invalid_configuration")
		return err
	}
	log.ZInfo(ctx, "signer configuration validated", "stage", "configuration_validation", "result", "success")

	err = s.save2Config(ctx, param)
	if err != nil {
		log.ZError(ctx, "signer configuration could not be saved", err, "stage", "configuration_save", "result", "failure", "error_code", "configuration_save_failed")
		return err
	}
	log.ZInfo(ctx, "signer configuration saved", "stage", "configuration_save", "result", "success", "file", ConfigFilePath)

	err = s.report2Merchant(ctx, param)
	if err != nil {
		log.ZError(ctx, "signer registration report failed", nil, "stage", "merchant_registration", "result", "failure", "error_code", "merchant_registration_failed")
		return err
	}
	log.ZInfo(ctx, "signer registration reported", "stage", "merchant_registration", "result", "success", "merchant_id", param.MerchantId)

	err = s.getSignerConfig(ctx, param)

	return nil
}

func (s *SignerInit) checkParam(ctx context.Context, param *config.Signer) error {
	// SecretKey authenticates every merchant API request via the Crypto-V1
	// HMAC signature; an empty or short key makes signatures trivially
	// forgettable, so refuse to start the signer with one.
	if param.SecretKey == "" {
		return fmt.Errorf("secretKey must not be empty")
	}
	if len(param.SecretKey) < 32 {
		return fmt.Errorf("secretKey must be at least 32 characters, got %d", len(param.SecretKey))
	}
	// check is valid ethereum address

	return nil
}

func (s *SignerInit) save2ConfigWithViper(ctx context.Context, param *config.Signer) error {
	viper.SetConfigName("signer")   // name of file (without extension)
	viper.SetConfigType("yaml")     //
	viper.AddConfigPath("./config") //

	err := viper.ReadInConfig()
	if err != nil {
		log.ZError(ctx, "[Init] read config file err", err, "file", viper.ConfigFileUsed())
	}

	viper.Set(FlagEthereumAccount, fmt.Sprintf("\"%v\"", param.EthereumAccount))
	viper.Set(FlagMerchantId, param.MerchantId)
	viper.Set(FlagMerchantUrl, param.MerchantUrl)
	viper.Set(FlagSignerUrl, param.SignerUrl)
	viper.Set(FlagSecretKey, param.SecretKey)

	err = viper.WriteConfig()
	if err != nil {
		log.ZError(ctx, "[Init] write config file err", err, "file", viper.ConfigFileUsed())
		return err
	}
	return nil
}

func (s *SignerInit) save2Config(ctx context.Context, param *config.Signer) error {
	cfg := map[string]interface{}{
		FlagConfigKeyMap[FlagEthereumAccount]: param.EthereumAccount,
		FlagConfigKeyMap[FlagMerchantId]:      param.MerchantId,
		FlagConfigKeyMap[FlagMerchantUrl]:     param.MerchantUrl,
		FlagConfigKeyMap[FlagSignerUrl]:       param.SignerUrl,
		FlagConfigKeyMap[FlagSecretKey]:       param.SecretKey,
	}

	err := s.writeQuotedYAML(ConfigFilePath, cfg)
	if err != nil {
		log.ZError(ctx, "[Init] write quoted yaml err", err, "file", ConfigFilePath)
		return err
	}
	return nil
}

func (s *SignerInit) report2Merchant(ctx context.Context, param *config.Signer) error {

	reportURL, err := url.JoinPath(param.MerchantUrl, URISignerReport)
	if err != nil {
		log.ZError(ctx, "merchant endpoint configuration is invalid", nil, "stage", "merchant_registration", "result", "failure", "error_code", "invalid_endpoint")
		return err
	}

	req := &ReportReq{
		MerchantId:      param.MerchantId,
		EthereumAccount: param.EthereumAccount,
		SignerUrl:       param.SignerUrl,
	}
	resp := &ReportResp{}

	err = HttpPost(ctx, reportURL, param, req, resp)
	if err != nil {
		log.ZError(ctx, "merchant registration request failed", nil, "stage", "merchant_registration", "result", "failure", "error_code", "merchant_request_failed")
		return err
	}
	return nil
}

// writeQuotedYAML writes a map[string]interface{} to a YAML file,
// forcing all string values to be double-quoted.
func (s *SignerInit) writeQuotedYAML(filename string, data map[string]interface{}) error {
	rootNode := mapToYAMLNode(data)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	return encoder.Encode(rootNode)
}

// mapToYAMLNode recursively converts a map to a yaml.Node with quoted strings
func mapToYAMLNode(data map[string]interface{}) *yaml.Node {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	for k, v := range data {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: k,
		}

		var valueNode *yaml.Node

		switch val := v.(type) {
		case string:
			valueNode = &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: val,
				Style: yaml.DoubleQuotedStyle, // force double quotes
			}
		case map[string]interface{}:
			valueNode = mapToYAMLNode(val) // recursive
		case int, int64, float64, bool:
			valueNode = &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fmt.Sprintf("%v", val),
			}
		default:
			// fallback: convert to string with quotes
			valueNode = &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fmt.Sprintf("%v", val),
				Style: yaml.DoubleQuotedStyle,
			}
		}

		node.Content = append(node.Content, keyNode, valueNode)
	}

	return node
}

func (s *SignerInit) getSignerConfig(ctx context.Context, param *config.Signer) error {

	req := &MerchantSignerConfigReq{
		MerchantId: param.MerchantId,
	}
	resp := &MerchantSignerConfigResp{}

	cfgURL, err := url.JoinPath(param.MerchantUrl, URISignerConfig)
	if err != nil {
		log.ZError(ctx, "merchant endpoint configuration is invalid", nil, "stage", "merchant_configuration", "result", "failure", "error_code", "invalid_endpoint")
		return err
	}

	err = HttpPost(ctx, cfgURL, param, req, resp)
	if err != nil {
		log.ZError(ctx, "merchant configuration request failed", nil, "stage", "merchant_configuration", "result", "failure", "error_code", "merchant_request_failed")
		return err
	}

	log.ZInfo(ctx, "merchant configuration received", "stage", "merchant_configuration", "result", "success",
		"config_page_enabled", resp.Data.EnableConfigPage, "ip_whitelist_count", len(resp.Data.IPWhiteList))

	return nil
}

func HttpPost(ctx context.Context, urlPath string, param *config.Signer, req, result interface{}) error {

	u, err := url.Parse(urlPath)
	if err != nil {
		log.ZError(ctx, "merchant endpoint is invalid", nil, "stage", "merchant_request", "result", "failure", "error_code", "invalid_endpoint")
		return err
	}

	reqBy, err := json.Marshal(req)
	if err != nil {
		log.ZError(ctx, "merchant request could not be encoded", err, "stage", "merchant_request", "result", "failure", "error_code", "request_encode_failed")
		return err
	}

	sg := signx.NewSignV2()
	vd, err := sg.CreateSign(ctx, param.SecretKey, u.RequestURI(), "Signer", param.MerchantId, "POST", string(reqBy))
	if err != nil {
		log.ZError(ctx, "[Init] CreateSign err", err)
		return err
	}

	ops := []httpx.ClientOptions{
		httpx.WithHeaderKV("Authorization", vd.Authorization),
		httpx.WithTimeout(time.Second * 30),
	}

	respBody, respHeader, statusCode, err := httpx.PostWithHeader(ctx, urlPath, reqBy, ops...)
	if err != nil {
		log.ZError(ctx, "merchant request failed", nil, "stage", "merchant_request", "result", "failure", "error_code", "merchant_request_failed", "http_status", statusCode)
		return fmt.Errorf("merchant request failed")
	}
	log.ZInfo(ctx, "merchant response received", "stage", "merchant_response", "result", "received", "response_bytes", len(respBody))

	auth := respHeader.Get("Authorization")
	if auth == "" {
		log.ZError(ctx, "merchant response is missing authorization", nil, "stage", "response_verification", "result", "failure", "error_code", "missing_authorization")
		return fmt.Errorf("missing Authorization header in response")
	}

	respVd := signx.VerifyData{
		Authorization: auth,
		SignInfo: signx.SignInfo{
			Method:       "POST",
			CanonicalURL: u.RequestURI(),
			SignBody:     string(respBody),
		},
	}

	err = sg.VerifySign(ctx, param.SecretKey, respVd)
	if err != nil {
		log.ZError(ctx, "merchant response signature verification failed", nil, "stage", "response_verification", "result", "failure", "error_code", "invalid_response_signature")
		return fmt.Errorf("merchant response signature verification failed")
	}

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		log.ZError(ctx, "merchant response could not be decoded", nil, "stage", "merchant_response_decode", "result", "failure", "error_code", "response_decode_failed")
		return fmt.Errorf("invalid merchant response")
	}

	return nil
}
