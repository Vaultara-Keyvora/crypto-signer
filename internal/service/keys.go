package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/internal/service/crypto"
	"lbe_crypto_signer/pkg/cipher"
	"lbe_crypto_signer/pkg/common/apistruct"
	"lbe_crypto_signer/pkg/tools/log"
)

// AddMnemonic signer machine add Mnemonic
func (l *LBESignerT) AddMnemonic(ctx *gin.Context, req *apistruct.KeysAddMnemonicReq) (*apistruct.KeysAddMnemonicResp, error) {
	var (
		err  error
		resp = &apistruct.KeysAddMnemonicResp{}
	)

	mnemonic, err := l.decodeMnemonic(ctx, req)
	if err != nil {
		log.ZError(ctx, "AddMnemonic decodeMnemonic", err)
		return nil, err
	}

	mastery, err := l.WalletCli.SeedToMasterKey(ctx, mnemonic, "")
	if err != nil {
		log.ZError(ctx, "AddMnemonic SeedToMasterKey", err)
		return nil, err
	}
	adds, err := l.WalletCli.AccountsFromDerive(ctx, mastery, mnemonic, req.Name)
	if err != nil {
		log.ZError(ctx, "AddMnemonic AccountsFromDerive", err)
		return nil, err
	}
	var accounts []apistruct.AccountBasic
	for _, v := range adds {
		t := apistruct.AccountBasic{
			Addr:  v.Addr,
			Chain: v.Chain,
		}
		accounts = append(accounts, t)
	}
	resp.Accounts = accounts
	return resp, err
}

func (l *LBESignerT) decodeMnemonic(ctx context.Context, req *apistruct.KeysAddMnemonicReq) (string, error) {
	fingerprint, ok := ctx.Value(FingerprintKey).(string)
	if !ok || req.Fingerprint != fingerprint {
		log.ZWarn(ctx, "mnemonic fingerprint validation failed", nil, "stage", "fingerprint_validation", "result", "failure", "error_code", "fingerprint_mismatch")
		return "", errors.New("invalid fingerprint")
	}
	privateKey, ok := ctx.Value(PrivateKey).(string)
	if !ok {
		return "", errors.New("private key not found")
	}

	mnemonic, err := cipher.RSADecrypt(privateKey, req.Mnemonic)
	return string(mnemonic), err
}

// KeyList signer machine list key
func (l *LBESignerT) KeyList(ctx *gin.Context, req *apistruct.KeysListReq) (*apistruct.KeysListResp, error) {
	var (
		resp = &apistruct.KeysListResp{}
	)
	keys, err := l.WalletCli.KeyList()
	if err != nil {
		log.ZError(ctx, "KeyList WalletCli.KeyList", err)
		return nil, err
	}
	for _, v := range keys {
		t := apistruct.KeyMeta{}
		t.KeyFLag = v.Key
		t.Name = v.AccountsComp.Name
		for _, a := range v.AccountsComp.Accounts {
			mid := apistruct.AccountBasic{}
			mid.Addr = a.Addr
			mid.Chain = a.Chain
			t.Accounts = append(t.Accounts, mid)
		}
		resp.KeySlice = append(resp.KeySlice, t)
	}
	return resp, err
}

// AddAccounts expand address by index
func (l *LBESignerT) AddAccounts(ctx *gin.Context, req *apistruct.AddAccountsReq) (*apistruct.AddAccountsResp, error) {
	ctx.Set("chain", req.Chain)
	var (
		resp = &apistruct.AddAccountsResp{}
	)
	accountSli, err := l.WalletCli.AddAccounts(ctx, req.BaseAddr, req.Chain, req.StartIndex, req.Count)
	if err != nil {
		log.ZError(ctx, "AddAccounts WalletCli.AddAccounts", err)
		return nil, err
	}
	resp.BaseAddr = req.BaseAddr
	for _, v := range accountSli {
		t := &apistruct.AccountDetail{}
		t.Addr = v.Addr
		t.Index = v.Index
		t.Chain = v.Chain
		resp.Adds = append(resp.Adds, t)
	}
	return resp, err
}

// AddDepositAddr add pooling addr
func (l *LBESignerT) AddDepositAddr(ctx *gin.Context, req *apistruct.AddDepositAddrReq) (*apistruct.AddDepositAddrResp, error) {
	var (
		resp = &apistruct.AddDepositAddrResp{}
	)
	err := l.WalletCli.AddDepositAddr(ctx, req.Addr, req.Name)
	if err != nil {
		log.ZError(ctx, "AddDepositAddr WalletCli.AddDepositAddr", err)
		return nil, err
	}
	resp.Addr = req.Addr
	return resp, err
}

// ListDepositAddr list pooling addr
func (l *LBESignerT) ListDepositAddr(ctx *gin.Context, req *apistruct.ListDepositAddrReq) (*apistruct.ListDepositAddrResp, error) {
	var (
		resp = &apistruct.ListDepositAddrResp{}
	)
	accounts, err := l.WalletCli.ListDepositAddr()
	if len(accounts) == 0 || err != nil {
		log.ZError(ctx, "AddDepositAddr WalletCli.AddDepositAddr valid accounts is 0", err)
		return nil, err
	}
	for _, v := range accounts {
		resp.DepositAccounts = append(resp.DepositAccounts, &apistruct.DepositAddrMeta{Addr: v.Addr, Name: v.Name})
	}
	return resp, err
}

// CheckAccounts check accounts by index
func (l *LBESignerT) CheckAccounts(ctx *gin.Context, req *apistruct.CheckAccountsReq) (*apistruct.CheckAccountsResp, error) {
	var (
		resp = &apistruct.CheckAccountsResp{}
	)

	accounts := make([]crypto.Account, 0, len(req.Accounts))
	for _, v := range req.Accounts {
		account := crypto.Account{
			BaseAddr: v.BaseAddr,
			Chain:    v.Chain,
			List:     make([]crypto.SubAccount, 0, len(v.List)),
		}
		for _, a := range v.List {
			account.List = append(account.List, crypto.SubAccount{
				Addr:  a.Addr,
				Index: int(a.Index),
			})
		}
		accounts = append(accounts, account)
	}

	err := l.WalletCli.CheckAccounts(ctx, accounts)
	if err != nil {
		log.ZError(ctx, "CheckAccounts WalletCli.CheckAccounts", err)
		return nil, err
	}

	return resp, nil
}
