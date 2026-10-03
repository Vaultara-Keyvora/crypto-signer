package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"lbe_crypto_signer/pkg/tools/log"
	"strings"
)

const (
	FirstHardenedChild = uint32(0x80000000)
	PurposeCoinType    = uint32(44) // purpose: 44'
	CoinTypeETH        = uint32(60) // eth coin_type: 60'
	AccountIndex       = uint32(0)  // account: 0'
	ExternalBranch     = uint32(0)
)

type WalletEthereum struct{}

func (w *WalletEthereum) AccountsFromPriBase(ctx context.Context, prvByte []byte) (string, error) {
	privateKey, err := crypto.ToECDSA(prvByte)
	if err != nil {
		log.ZError(ctx, "WalletEthereum AccountsFromPriBase invalid private key:", err)
		return "", err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.ZError(ctx, "WalletEthereum AccountsFromPriBase invalid publicKeyECDSA key:", err)
		return "", err
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}

func (w *WalletEthereum) AccountsFromMnemonicIndex(ctx context.Context, Mnemonic string, index int) (string, error) {
	keySeeds := bip39.NewSeed(Mnemonic, "")
	masterKey, err := bip32.NewMasterKey(keySeeds)
	if err != nil {
		log.ZError(ctx, "SeedToMastKey NewMasterKey failed:", err)
		return "", err
	}
	return w.deriveEthereumAddress(ctx, masterKey, uint32(index))
}

func (w *WalletEthereum) AccountsFromDeriveIndex(ctx context.Context, masterKey *bip32.Key, index int) (string, error) {
	return w.deriveEthereumAddress(ctx, masterKey, 0)
}

func (w *WalletEthereum) SignTx(ctx context.Context, Mnemonic string, addr string, index int, chainId string, toAddr string, amount string, needSign []byte) ([]byte, error) {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(needSign)
	if err != nil {
		log.ZError(ctx, "ethChainSign tx.UnmarshalBinary exception", err)
		return nil, err
	}
	id, err := decimal.NewFromString(chainId)
	if err != nil {
		log.ZError(ctx, "ethChainSign chain ID parsing failed", err, "stage", "transaction_decode", "error_code", "invalid_chain_id")
		return nil, err
	}

	keySeeds := bip39.NewSeed(Mnemonic, "")
	masterKey, err := bip32.NewMasterKey(keySeeds)
	if err != nil {
		log.ZError(ctx, "SignTx NewMasterKey failed:", err)
		return nil, err
	}

	priKeys, err := w.deriveEthereumPriKey(ctx, masterKey, uint32(index), addr)
	if err != nil {
		log.ZError(ctx, "SignTx NewMasterKey failed:", err)
		return nil, err
	}

	prv, err := crypto.HexToECDSA(hex.EncodeToString(priKeys))
	if err != nil {
		log.ZError(ctx, "SignTx HexToECDSA failed:", err)
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(id.BigInt()), prv)
	if err != nil {
		log.ZError(ctx, "ethChainSign types.SignTx exception", err)
		return nil, err
	}
	signByte, err := signedTx.MarshalBinary()
	if err != nil {
		log.ZError(ctx, "ethChainSign signedTx.MarshalBinary exception", err)
		return nil, err
	}
	return signByte, nil
}

func (w *WalletEthereum) deriveEthereumAddress(ctx context.Context, masterKey *bip32.Key, addressIndex uint32) (string, error) {
	_, pubKey, err := w.basicDerive(ctx, masterKey, addressIndex)
	if err != nil {
		log.ZError(ctx, "basicDerive failed:", err)
		return "", err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

func (w *WalletEthereum) deriveEthereumPriKey(ctx context.Context, masterKey *bip32.Key, addressIndex uint32, addr string) ([]byte, error) {
	prv, pubKey, err := w.basicDerive(ctx, masterKey, addressIndex)
	if err != nil {
		log.ZError(ctx, "basicDerive failed:", err)
		return nil, err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	if strings.ToLower(address.Hex()) != strings.ToLower(addr) {
		return nil, errors.New("derived signer address mismatch")
	}
	return prv, nil
}

func (w *WalletEthereum) basicDerive(ctx context.Context, masterKey *bip32.Key, addressIndex uint32) ([]byte, *ecdsa.PublicKey, error) {
	// first layer: purpose' (44')
	purpose, err := masterKey.NewChildKey(FirstHardenedChild + PurposeCoinType)
	if err != nil {
		log.ZError(ctx, "derive purpose failed:", err)
		return nil, nil, err
	}

	// second layer: coin_type' (60' eth)
	coinType, err := purpose.NewChildKey(FirstHardenedChild + CoinTypeETH)
	if err != nil {
		log.ZError(ctx, "derive coin_type failed:", err)
		return nil, nil, err
	}

	// third layer : account' (0')
	account, err := coinType.NewChildKey(FirstHardenedChild + AccountIndex)
	if err != nil {
		log.ZError(ctx, "derive account failed:", err)
		return nil, nil, err
	}

	// fourth layer: change (0 = outer input)
	change, err := account.NewChildKey(ExternalBranch)
	if err != nil {
		log.ZError(ctx, "derive change failed:", err)
		return nil, nil, err
	}

	// fifth layer: address_index
	childKey, err := change.NewChildKey(addressIndex)
	if err != nil {
		log.ZError(ctx, "derive childKey failed:", err)
		return nil, nil, err
	}

	publicKeyBytes := childKey.PublicKey().Key
	pubKey, err := crypto.DecompressPubkey(publicKeyBytes)
	if err != nil {
		log.ZError(ctx, "Decompress PubKey childKey failed:", err)
		return nil, nil, err
	}

	return childKey.Key, pubKey, nil
}

func (w *WalletEthereum) SafetyCheck(ctx context.Context, addr string, toAddr string, amount string, needSign []byte) (string, error) {
	//todo
	return "", errors.New("not implement")
}
