package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/pkg/cipher"
	"lbe_crypto_signer/pkg/common/apistruct"
	"lbe_crypto_signer/pkg/tools/log"
	"strconv"
	"strings"
	"time"
)

const (
	Pos64               = 64
	Pos27               = 27
	Pos28               = 28
	EthSignConst        = "\u0019Ethereum Signed Message:\n"
	TokenLength         = 32
	ToKenValidTime      = 1 * time.Hour // 1 hour
	LoginSeedsValidTime = 5 * 60 * 1000 // 5 minute
)

type AutoToken struct {
	Token        string
	StartTime    int64
	EndTimeStamp int64
	Fingerprint  string // 指纹
	PrivateKey   string // 公钥
}

const (
	FingerprintKey = "fingerprint"
	PrivateKey     = "private"
)

// Login by admin
func (l *LBESignerT) Login(ctx *gin.Context, req *apistruct.AdminLoginReq) (*apistruct.AdminLoginResp, error) {
	log.ZInfo(ctx, "admin login started", "stage", "admin_login", "result", "started")

	// check seeds
	if err := l.checkSeeds(ctx, req.Seeds); err != nil {
		log.ZError(ctx, "admin login seed validation failed", err, "stage", "admin_login", "result", "failure", "error_code", "invalid_login_seed")
		return nil, err
	}

	hashedMessage := []byte(fmt.Sprintf("%s%s%s", EthSignConst, strconv.Itoa(len(req.Seeds)), req.Seeds))
	hash := crypto.Keccak256Hash(hashedMessage)
	decodedMessage, err := hexutil.Decode(req.Sign)
	if err != nil || len(decodedMessage) != 65 {
		log.ZWarn(ctx, "admin login signature is malformed", nil, "stage", "admin_login", "result", "failure", "error_code", "invalid_signature")
		return nil, fmt.Errorf("the signature data is corrupted")
	}
	if decodedMessage[Pos64] == Pos27 || decodedMessage[Pos64] == Pos28 {
		decodedMessage[Pos64] -= Pos27
	}
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), decodedMessage)
	if err != nil {
		log.ZError(ctx, "admin login signature recovery failed", err, "stage", "admin_login", "result", "failure", "error_code", "signature_recovery_failed")
		return nil, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*sigPublicKeyECDSA).Hex()
	if strings.ToLower(req.Address) != strings.ToLower(recoveredAddr) {
		log.ZWarn(ctx, "admin login address does not match signature", nil, "stage", "admin_login", "result", "failure", "error_code", "address_mismatch")
		return nil, fmt.Errorf("recovered address is not same as signed address")
	}
	if !strings.EqualFold(recoveredAddr, l.Conf.SignerConf.EthereumAccount) {
		log.ZWarn(ctx, "admin login address is not configured", nil, "stage", "admin_login", "result", "failure", "error_code", "unauthorized_admin_address")
		return nil, fmt.Errorf("recovered address is not same as configured address")
	}

	token, err := l.GenerateSecureToken(TokenLength)
	if err != nil {
		log.ZError(ctx, "admin login token generation failed", err, "stage", "admin_login", "result", "failure", "error_code", "token_generation_failed")
		return nil, err
	}

	pubKey, priKey, fingerprint, err := cipher.GenRSA()
	if err != nil {
		log.ZError(ctx, "admin login key generation failed", err, "stage", "admin_login", "result", "failure", "error_code", "key_generation_failed")
		return nil, err
	}

	l.TokenHolder.Store(token, &AutoToken{
		Token:        token,
		StartTime:    time.Now().Unix(),
		EndTimeStamp: time.Now().Add(ToKenValidTime).Unix(),
		Fingerprint:  fingerprint,
		PrivateKey:   priKey,
	})

	log.ZInfo(ctx, "admin login completed", "stage", "admin_login", "result", "success")
	return &apistruct.AdminLoginResp{Token: token, Fingerprint: fingerprint, PublicKey: pubKey}, nil
}

const (
	seedsPrefix = "LBE-"
)

func (l *LBESignerT) checkSeeds(ctx *gin.Context, seeds string) error {
	seedsTime := strings.Replace(seeds, seedsPrefix, "", 1)
	ts, err := strconv.ParseInt(seedsTime, 10, 64)
	if err != nil {
		return fmt.Errorf("the seed time format is invalid")
	}

	minTs, maxTs := time.Now().UnixMilli()-LoginSeedsValidTime, time.Now().UnixMilli()+LoginSeedsValidTime
	if ts >= minTs && ts <= maxTs {
		return nil
	}
	return fmt.Errorf("invalid seed time, please re-sign")
}

func (l *LBESignerT) GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	return token, nil
}

// LoginOut by Admin
func (l *LBESignerT) LoginOut(ctx *gin.Context, req *apistruct.AdminLoginOutReq) (*apistruct.AdminLoginOutResp, error) {
	log.ZInfo(ctx, "admin logout started", "stage", "admin_logout", "result", "started")
	l.TokenHolder.Delete(req.Token)
	log.ZInfo(ctx, "admin logout completed", "stage", "admin_logout", "result", "success")

	return &apistruct.AdminLoginOutResp{}, nil
}

func (l *LBESignerT) TokenAuth(ctx *gin.Context, token string) error {
	tokenValue, ok := l.TokenHolder.Load(ctx, token)
	if !ok {
		return fmt.Errorf("invalid token")
	}
	authToken, ok := tokenValue.(*AutoToken)
	if !ok {
		log.ZWarn(ctx, "configuration page token record is invalid", nil, "stage", "token_authentication", "result", "failure", "error_code", "invalid_token_record")
		return fmt.Errorf("invalid token")
	}
	if authToken.EndTimeStamp < time.Now().Unix() {
		l.TokenHolder.Delete(token)
		log.ZInfo(ctx, "configuration page token expired", "stage", "token_authentication", "result", "failure", "error_code", "expired_token")
		return fmt.Errorf("expired token")
	}
	ctx.Set(FingerprintKey, authToken.Fingerprint)
	ctx.Set(PrivateKey, authToken.PrivateKey)

	return nil
}
