package bsc

import (
	"context"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"lbe_crypto_signer/internal/service/crypto/ethereum"
)

type WalletBSC struct {
	extendsEth *ethereum.WalletEthereum
}

func NewWalletBSC(ethWallet *ethereum.WalletEthereum) *WalletBSC {
	return &WalletBSC{
		extendsEth: ethWallet,
	}
}

// same rules as ethereum
func (w *WalletBSC) AccountsFromPriBase(ctx context.Context, prvByte []byte) (string, error) {
	return w.extendsEth.AccountsFromPriBase(ctx, prvByte)
}

func (w *WalletBSC) AccountsFromDeriveIndex(ctx context.Context, masterKey *bip32.Key, index int) (string, error) {
	return w.extendsEth.AccountsFromDeriveIndex(ctx, masterKey, index)
}

func (w *WalletBSC) AccountsFromMnemonicIndex(ctx context.Context, Mnemonic string, index int) (string, error) {
	return w.extendsEth.AccountsFromMnemonicIndex(ctx, Mnemonic, index)
}

func (w *WalletBSC) SignTx(ctx context.Context, Mnemonic string, addr string, index int, chainId string, toAddr string, amount string, needSign []byte) ([]byte, error) {
	return w.extendsEth.SignTx(ctx, Mnemonic, addr, index, chainId, toAddr, amount, needSign)
}

func (w *WalletBSC) SafetyCheck(ctx context.Context, addr string, toAddr string, amount string, needSign []byte) (string, error) {
	//todo
	return "", errors.New("not implement")
}
