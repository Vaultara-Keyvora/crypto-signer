package tron

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"lbe_crypto_signer/pkg/tools/log"
	"strings"
)

func (w *WalletTron) TronTrc20DataDecode(ctx context.Context, dataHex string) (string, error) {
	decResult, err := w.DecodeTriggerSmartContractData(w.UsdtAbi(), dataHex)
	if err != nil {
		log.ZError(ctx, "TronTrc20DataDecode DecodeTriggerSmartContractData exception", err)
		return "", err
	}
	if toAddr, ok := decResult["_to"].(common.Address); ok {
		tronAddrHex := "41" + strings.TrimPrefix(toAddr.Hex(), "0x")
		tronAddrBytes, err := hex.DecodeString(tronAddrHex)
		if err != nil {
			log.ZError(ctx, "TronTrc20DataDecode hex.DecodeString exception", err, "stage", "transaction_decode", "error_code", "invalid_recipient_encoding")
			return "", err
		}
		addr := address.BytesToAddress(tronAddrBytes)
		return addr.String(), nil
	}
	return "", fmt.Errorf("TronTrc20DataDecode exception invalid dataHex")
}

func (w *WalletTron) UsdtAbi() string {
	ustABI := `[
        {
            "constant": false,
            "inputs": [
                {"name": "_to", "type": "address"},
                {"name": "_value", "type": "uint256"}
            ],
            "name": "transfer",
            "outputs": [{"name": "", "type": "bool"}],
            "type": "function"
        }
    ]`
	return ustABI
}

func (w *WalletTron) DecodeTriggerSmartContractData(contractABI string, dataHex string) (map[string]interface{}, error) {
	if !strings.HasPrefix(dataHex, "0x") {
		dataHex = "0x" + dataHex
	}
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("parse ABI failed: %v", err)
	}
	dataBytes := common.FromHex(dataHex)
	if len(dataBytes) < 4 {
		return nil, fmt.Errorf("data too short: %d bytes", len(dataBytes))
	}
	methodID := dataBytes[:4]
	method, err := parsedABI.MethodById(methodID)
	if err != nil {
		return nil, fmt.Errorf("method not found: %v", err)
	}
	inputs := map[string]interface{}{}
	err = method.Inputs.UnpackIntoMap(inputs, dataBytes[4:])
	if err != nil {
		return nil, fmt.Errorf("unpack failed: %v", err)
	}
	return inputs, nil
}
