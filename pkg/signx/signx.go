package signx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"sync"

	// "crypto_project/common/consts"
	"encoding/hex"
	"errors"
	"fmt"
	"lbe_crypto_signer/pkg/tools/log"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cast"
)

var (
	ErrApiSignVersion error = errors.New("签名版本错误")
)

type SignVersion string

const (
	SignVersionV1 SignVersion = "Crypto-V1"
)

// 设计思想参考微信支付
// https://pay.weixin.qq.com/doc/v3/merchant/4012365334

type SignV2 struct {
	// Appsecret string `json:"appsecret"` // 必填
}

func NewSignV2() *SignV2 {
	return &SignV2{
		// Appsecret: appsecret,
	}
}

// 签名有效期
var WindowsDuration = time.Minute * 5

// usedNonces 记录已使用过的 nonce:sign_time, 防止请求重放
var (
	usedNonces    sync.Map // key: "nonce:sign_time", value: 过期毫秒时间戳
	nonceCacheTTL = WindowsDuration
	lock          sync.Mutex
	lastClearAt   time.Time
)

// 签名参数
type AuthInfo struct {
	MerchantId string `json:"merchant_id"` // 必填 商户ID Header
	Appid      string `json:"appid"`       // 必填 Header
	SignTime   int64  `json:"sign_time"`   // 毫秒时间戳 Header
	Nonce      string `json:"nonce"`       // 随机数 Header
	Sign       string `json:"sign"`        // 签名值 Header
}

// 签名信息
type SignInfo struct {
	Method       string `json:"method"`        // 必填
	CanonicalURL string `json:"canonical_url"` // 必填 参数(去掉Host) 如/user/info?name=123&age=18
	SignBody     string `json:"sign_body"`     // 请求体
}

// 验签参数
type VerifyData struct {
	Authorization string `json:"authorization"` // 必填(在Header中) 如 Crypto-V2 merchant_id:xxx&appid:xxx&nonce:xxx&sign_time:xxx&sign:xxx
	SignInfo
}

// 签名参数
type SignData struct {
	AuthInfo
	SignInfo
}

// 创建签名
func (s *SignV2) CreateSign(ctx context.Context, appsecret, canonicalURL string, appid string, merchantId string, method string, signBody string) (*VerifyData, error) {

	info := SignData{
		SignInfo: SignInfo{
			Method:       method,
			CanonicalURL: canonicalURL,
			SignBody:     signBody,
		},
	}
	info.Appid = appid
	info.MerchantId = merchantId
	signHeader, _, err := s.getSign(ctx, appsecret, info)
	if err != nil {
		log.ZError(ctx, "getSign err: %v", err)
		return nil, err
	}

	return &VerifyData{
		Authorization: signHeader,
		SignInfo:      info.SignInfo,
	}, nil
}

// 验证签名
func (s *SignV2) VerifySign(ctx context.Context, appsecret string, verifyData VerifyData) error {
	if verifyData.Authorization == "" {
		return errors.New("Authorization 不能为空")
	}

	authInfo, err := s.ParseAuthStr(ctx, verifyData.Authorization)
	if err != nil {
		log.ZError(ctx, "ParseAuthStr err: %v", err)
		return err
	}

	// 重放保护: 同一 nonce+sign_time 在签名窗口内只能使用一次
	nonceKey := fmt.Sprintf("%s:%d", authInfo.Nonce, authInfo.SignTime)
	if exp, loaded := usedNonces.Load(nonceKey); loaded {
		if exp.(int64) > time.Now().UnixMilli() {
			log.ZError(ctx, "nonce reused", nil)
			return errors.New("nonce 已使用")
		}
		usedNonces.Delete(nonceKey)
	}

	signData := SignData{
		AuthInfo: *authInfo,
		SignInfo: verifyData.SignInfo,
	}

	_, sign, err := s.getSign(ctx, appsecret, signData)
	if err != nil {
		log.ZError(ctx, "getSign err: %v", err)
		return err
	}

	if signData.Sign != sign {
		log.ZError(ctx, "Signature verification failed.", nil)
		return errors.New("签名验证失败")
	}

	// 验签通过后记录 nonce, 过期时间与签名窗口一致
	usedNonces.Store(nonceKey, time.Now().Add(nonceCacheTTL).UnixMilli())

	ClearNonce(ctx)

	return nil
}

// 生成签名
func (s *SignV2) getSign(ctx context.Context, appsecret string, info SignData) (signHeader string, sign string, err error) {
	// 创建副本避免修改原始参数
	signData := info

	// Never create or verify a signature with an empty or short shared
	// secret: HMAC over an empty key is trivially forgettable by any caller.
	if appsecret == "" || len(appsecret) < 32 {
		log.ZError(ctx, "The appsecret must not be empty or too short.", nil)
		return "", "", errors.New("appsecret 不能为空或过短")
	}
	if signData.MerchantId == "" ||
		signData.Appid == "" ||
		signData.Method == "" ||
		signData.CanonicalURL == "" {
		log.ZError(ctx, "The signature parameter is incorrect.", err)
		return "", "", errors.New("签名参数错误")
	}

	// Method 转换为大写
	signData.Method = strings.ToUpper(signData.Method)

	if signData.SignTime == 0 {
		signData.SignTime = time.Now().UnixMilli()
	}
	// 有效时间正负5分钟
	if signData.SignTime < time.Now().Add(-WindowsDuration).UnixMilli() || signData.SignTime > time.Now().Add(WindowsDuration).UnixMilli() {
		log.ZError(ctx, "The signature time is incorrect.", nil)
		return "", "", errors.New("签名时间错误")
	}
	// 随机数
	if signData.Nonce == "" {
		u, err := uuid.NewV7()
		if err != nil {
			log.ZError(ctx, "uuid.NewV7 err: %v", err)
			return "", "", err
		}
		signData.Nonce = u.String()
	}
	// 解析URL 去掉host
	u, err := url.Parse(signData.CanonicalURL)
	if err != nil {
		log.ZError(ctx, "signature canonical URL is invalid", nil, "stage", "signature_generation", "result", "failure", "error_code", "invalid_canonical_url")
		return "", "", err
	}
	signData.CanonicalURL = u.RequestURI()

	signStr := s.buildSignStr(ctx, signData)
	log.ZInfo(ctx, "signature payload prepared", "stage", "signature_generation", "result", "prepared", "request_bytes", len(signData.SignBody))

	// 使用 HMAC-SHA256 算法生成签名
	h := hmac.New(sha256.New, []byte(appsecret))
	h.Write([]byte(signStr))
	signature := hex.EncodeToString(h.Sum(nil))
	signData.Sign = signature

	auth, err := s.buildAuthStr(ctx, signData)
	if err != nil {
		log.ZError(ctx, "BuildAuthorizationV2 err: %v", err)
		return "", "", err
	}

	return auth, signData.Sign, nil
}

// 构建签名字符串
func (s *SignV2) buildSignStr(ctx context.Context, info SignData) string {
	str := fmt.Sprintf("merchant_id:%s&appid:%s&method:%s&nonce:%s&sign_time:%d&canonical_url:%s&sign_body:%s",
		info.MerchantId, info.Appid, info.Method, info.Nonce, info.SignTime, info.CanonicalURL, info.SignBody)
	return str
}

// 构造 V2 Auth
func (s *SignV2) buildAuthStr(ctx context.Context, info SignData) (string, error) {
	str := fmt.Sprintf("merchant_id:%s&appid:%s&nonce:%s&sign_time:%d&sign:%s", info.MerchantId, info.Appid, info.Nonce, info.SignTime, info.Sign)

	str = fmt.Sprintf("%s %s", SignVersionV1, str)

	return str, nil
}

// 解析 V2 Auth
// Crypto-V2 merchant_id:xxx&appid:xxx&nonce:xxx&sign_time:xxx&sign:xxx
func (s *SignV2) ParseAuthStr(ctx context.Context, auth string) (*AuthInfo, error) {

	str, ok := strings.CutPrefix(auth, string(SignVersionV1))
	if !ok {
		log.ZError(ctx, "strings.CutPrefix cut not ok", nil)
		return nil, ErrApiSignVersion
	}

	// 版本2签名
	str = strings.TrimSpace(str)

	strs := strings.Split(str, "&")
	data := AuthInfo{}
	for _, s := range strs {
		kv := strings.Split(s, ":")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "appid":
			data.Appid = kv[1]
		case "sign":
			data.Sign = kv[1]
		case "nonce":
			data.Nonce = kv[1]
		case "sign_time":
			data.SignTime = cast.ToInt64(kv[1])
		case "merchant_id":
			data.MerchantId = kv[1]
		}
	}

	// 验证参数
	// 有效时间正负5分钟
	if data.SignTime < time.Now().Add(-WindowsDuration).UnixMilli() || data.SignTime > time.Now().Add(WindowsDuration).UnixMilli() {
		log.ZError(ctx, "The signature time is incorrect.", nil)
		return nil, errors.New("签名时间错误")
	}

	return &data, nil

}

func ClearNonce(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()

	if lastClearAt.Add(WindowsDuration).After(time.Now()) {
		return
	}
	cleared := 0
	usedNonces.Range(func(k, v interface{}) bool {
		if v.(int64) < time.Now().UnixMilli() {
			usedNonces.Delete(k)
			cleared++
		}
		return true
	})
	if cleared > 0 {
		log.ZInfo(ctx, "expired signature nonces cleared", "stage", "nonce_cleanup", "result", "success", "cleared_count", cleared)
	}

	lastClearAt = time.Now()

	return
}
