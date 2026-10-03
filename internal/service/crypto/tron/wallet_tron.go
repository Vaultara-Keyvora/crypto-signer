package tron

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/ryanbekhen/tronwallet"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"google.golang.org/protobuf/proto"
	"lbe_crypto_signer/pkg/tools/log"
	"strings"
)

const (
	FirstHardenedChild = uint32(0x80000000)
	PurposeCoinType    = uint32(44)  // purpose: 44'
	CoinTypeTron       = uint32(195) // eth coin_type: 60'
	AccountIndex       = uint32(0)   // account: 0'
	ExternalBranch     = uint32(0)
	TRXTransferType    = "TransferContract"
	TRX20TransferType  = "TriggerSmartContract"
)

type WalletTron struct{}

func (w *WalletTron) AccountsFromDeriveIndex(ctx context.Context, masterKey *bip32.Key, index int) (string, error) {
	return w.deriveTronAddress(ctx, masterKey, uint32(index))
}

func (w *WalletTron) AccountsFromMnemonicIndex(ctx context.Context, Mnemonic string, index int) (string, error) {
	keySeeds := bip39.NewSeed(Mnemonic, "")
	masterKey, err := bip32.NewMasterKey(keySeeds)
	if err != nil {
		log.ZError(ctx, "SeedToMastKey NewMasterKey failed:", err)
		return "", err
	}
	return w.deriveTronAddress(ctx, masterKey, uint32(index))
}

func (w *WalletTron) SignTx(ctx context.Context, Mnemonic string, addr string, index int, chainId string, toAddr string, amount string, needSign []byte) ([]byte, error) {
	keySeeds := bip39.NewSeed(Mnemonic, "")
	masterKey, err := bip32.NewMasterKey(keySeeds)
	if err != nil {
		log.ZError(ctx, "WalletTron SignTx failed:", err)
		return nil, err
	}
	prv, _, err := w.basicDerive(ctx, masterKey, uint32(index))
	if err != nil {
		log.ZError(ctx, "WalletTron basicDerive failed:", err)
		return nil, err
	}
	address := tronwallet.TronAddressFromPrivate(prv)
	if strings.ToLower(address) != strings.ToLower(addr) {
		err := fmt.Errorf("derived signer address mismatch")
		log.ZError(ctx, "derived signer address mismatch", err, "stage", "signer_address_validation", "result", "failure", "error_code", "address_mismatch")
		return nil, err
	}
	h256h := sha256.New()
	h256h.Write(needSign)
	hash := h256h.Sum(nil)
	sign, err := crypto.Sign(hash, prv)
	if err != nil {
		log.ZError(ctx, "WalletTron crypto.Sign err=%+v", err)
		return nil, err
	}
	return sign, nil
}

func (w *WalletTron) deriveTronAddress(ctx context.Context, masterKey *bip32.Key, addressIndex uint32) (string, error) {
	prv, _, err := w.basicDerive(ctx, masterKey, addressIndex)
	if err != nil {
		log.ZError(ctx, "deriveTronAddress basicDerive failed:", err)
		return "", err
	}
	address := tronwallet.TronAddressFromPrivate(prv)
	return address, nil
}

func (w *WalletTron) basicDerive(ctx context.Context, masterKey *bip32.Key, addressIndex uint32) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	// first layer: purpose' (44')
	purpose, err := masterKey.NewChildKey(FirstHardenedChild + PurposeCoinType)
	if err != nil {
		log.ZError(ctx, "derive purpose failed:", err)
		return nil, nil, err
	}
	// second layer: coin_type' (195' tron)
	coinType, err := purpose.NewChildKey(FirstHardenedChild + CoinTypeTron)
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
	// fourth layer: change (0 = 外部地址)
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
	privateKey, err := crypto.ToECDSA(childKey.Key)
	if err != nil {
		log.ZError(ctx, "AccountsFromPriBase invalid private key:", err)
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

func (w *WalletTron) SafetyCheck(ctx context.Context, addr string, toAddr string, amount string, needSign []byte) (string, error) {
	log.ZInfo(ctx, "Tron transaction safety check started", "stage", "transaction_safety_check", "result", "started", "transaction_bytes", len(needSign))
	var c core.TransactionRaw
	if err := proto.Unmarshal(needSign, &c); err != nil {
		log.ZError(ctx, "proto.Unmarshal needSign exception", err)
		return "", err
	}
	contract := c.GetContract()
	if len(contract) != 1 {
		err := fmt.Errorf("expected exactly 1 contract, got %d", len(contract))
		log.ZError(ctx, "safetyCheck contract count invalid", err)
		return "", fmt.Errorf("signdata is invalid: %v", err.Error())
	}
	for _, v := range contract {
		contractType := fmt.Sprintf("%s", v.Type)
		log.ZInfo(ctx, "Tron transaction contract identified", "stage", "transaction_safety_check", "contract_type", contractType)
		if contractType == TRXTransferType {
			tx := &core.TransferContract{}
			err := proto.Unmarshal(v.GetParameter().GetValue(), tx)
			if err != nil {
				log.ZError(ctx, "proto.Unmarshal contract exception", err)
				return "", err
			}
			to := address.BytesToAddress(tx.ToAddress)
			log.ZInfo(ctx, "Tron transfer contract validated", "stage", "transaction_safety_check", "result", "success", "contract_type", contractType)
			return to.String(), nil
		} else if contractType == TRX20TransferType {
			tx := &core.TriggerSmartContract{}
			err := proto.Unmarshal(v.GetParameter().GetValue(), tx)
			if err != nil {
				log.ZError(ctx, "proto.Unmarshal contract exception", err)
				return "", err
			}
			hexData := hex.EncodeToString(tx.GetData())
			receiverAddr, err := w.TronTrc20DataDecode(ctx, hexData)
			if err != nil {
				log.ZError(ctx, "safetyCheck TronTrc20DataDecode exception", err)
				return "", err
			}
			log.ZInfo(ctx, "Tron token transfer contract validated", "stage", "transaction_safety_check", "result", "success", "contract_type", contractType)
			return receiverAddr, nil
		}
	}
	return "", fmt.Errorf("signdata is invalid")
}
