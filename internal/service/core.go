package service

import (
	"context"
	"lbe_crypto_signer/internal/logic"
	"lbe_crypto_signer/internal/service/crypto"
	"lbe_crypto_signer/pkg/common/config"
)

type ActiveConfig struct {
	ApiConf    *config.ApiConfig
	SignerConf *config.Signer
}

type LBESignerT struct {
	Conf        *ActiveConfig
	WalletCli   *crypto.WalletBase
	TokenHolder *logic.TokenHolder
}

func NewLBESignerT(conf *ActiveConfig) (*LBESignerT, error) {
	t := &LBESignerT{
		Conf:        conf,
		TokenHolder: logic.NewTokenHolder(),
	}
	err := t.initWalletCli(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = logic.InitSignerAuth(conf.SignerConf)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (l *LBESignerT) initWalletCli(ctx context.Context) error {
	l.WalletCli = crypto.NewWalletBase(l.Conf.ApiConf.Chain)
	return nil
}

var Core *LBESignerT

func Init(config *ActiveConfig) {
	var err error
	Core, err = NewLBESignerT(config)
	if err != nil {
		panic(err)
	}
}
