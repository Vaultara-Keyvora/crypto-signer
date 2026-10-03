package crypto

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
	"lbe_crypto_signer/internal/service/crypto/ethereum"
	"lbe_crypto_signer/internal/service/crypto/solana"
	"lbe_crypto_signer/internal/service/crypto/tron"
	"lbe_crypto_signer/pkg/cipher"
	"lbe_crypto_signer/pkg/tools/log"
	"strings"
	"sync"
)

const (
	ETH_TYP_ACCOUNT    = "ethereum"
	BSC_TYP_ACCOUNT    = "bsc"
	TRON_TYP_ACCOUNT   = "tron"
	SOLANA_TYP_ACCOUNT = "solana"
)

const MasterAccountIndex = 0

const AesKeyLength = 16

type WalletBehaviors interface {
	AccountsFromDeriveIndex(ctx context.Context, masterKey *bip32.Key, index int) (string, error)
	AccountsFromMnemonicIndex(ctx context.Context, Mnemonic string, index int) (string, error)
	SignTx(ctx context.Context, Mnemonic string, addr string, index int, chainId string, toAddr string, amount string, needSign []byte) ([]byte, error)
	SafetyCheck(ctx context.Context, addr string, toAddr string, amount string, needSign []byte) (string, error)
}

type AccountDetail struct {
	Addr  string
	Chain string
	Index int64
}

type AccountsMeta struct {
	Key          string
	AccountsComp *AccountsComp
}

type AccountsComp struct {
	Name     string
	Accounts []AccountsBase
}

type AccountsBase struct {
	Addr  string `json:"addr"`
	Chain string `json:"chain"`
}

type DepositAddrMeta struct {
	Name string
	Addr string
}

type WalletBase struct {
	KeyRamStore      sync.Map
	KeyRelationStore sync.Map
	DepositAddrStore sync.Map
	EthWallet        WalletBehaviors
	TronWallet       WalletBehaviors
	SolanaWallet     WalletBehaviors
	ValidChain       []string
	aesKey           string
}

func NewWalletBase(chain []string) *WalletBase {
	return &WalletBase{
		EthWallet:    &ethereum.WalletEthereum{},
		TronWallet:   &tron.WalletTron{},
		SolanaWallet: &solana.WalletSolana{},
		ValidChain:   chain,
		aesKey:       cipher.RandString(AesKeyLength),
	}
}

func (w *WalletBase) GenerateMnemonic(ctx context.Context, size MnemonicBitSize) (string, error) {
	entropy, err := bip39.NewEntropy(int(size))
	if err != nil {
		log.ZError(ctx, "GenerateMnemonic NewEntropy failed:", err)
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		log.ZError(ctx, "GenerateMnemonic NewMnemonic failed:", err)
		return "", err
	}
	return mnemonic, nil
}

func (w *WalletBase) SeedToMasterKey(ctx context.Context, seed string, passPhase string) (*bip32.Key, error) {
	keySeeds := bip39.NewSeed(seed, passPhase)
	masterKey, err := bip32.NewMasterKey(keySeeds)
	if err != nil {
		log.ZError(ctx, "SeedToMastKey NewMasterKey failed:", err)
		return nil, err
	}
	return masterKey, nil
}

func (w *WalletBase) AccountsFromDerive(ctx context.Context, masterKey *bip32.Key, Mnemonic string, name string) ([]AccountsBase, error) {
	accountSli := []AccountsBase{}
	for _, c := range w.ValidChain {
		chainCtx := context.WithValue(ctx, "chain", c)
		var (
			masterAddress string
			err           error
		)
		if c == ETH_TYP_ACCOUNT {
			masterAddress, err = w.EthWallet.AccountsFromDeriveIndex(chainCtx, masterKey, MasterAccountIndex)
		} else if c == TRON_TYP_ACCOUNT {
			masterAddress, err = w.TronWallet.AccountsFromDeriveIndex(chainCtx, masterKey, MasterAccountIndex)
		} else if c == SOLANA_TYP_ACCOUNT {
			masterAddress, err = w.SolanaWallet.AccountsFromMnemonicIndex(chainCtx, Mnemonic, MasterAccountIndex)
		} else {
			log.ZError(chainCtx, "AccountsFromDeriveIndex continue", fmt.Errorf("unknown chain"), "stage", "account_derivation")
			continue
		}
		if err != nil {
			log.ZError(chainCtx, "AccountsFromDeriveIndex failed", err, "stage", "account_derivation")
			return nil, err
		}

		err = w.storeMnemonic(chainCtx, masterAddress, Mnemonic)
		if err != nil {
			log.ZError(chainCtx, "storeMnemonic failed", err, "stage", "account_storage")
			continue
		}
		accountSli = append(accountSli, AccountsBase{Addr: masterAddress, Chain: c})

	}
	MnemonicSeed := fmt.Sprintf("%s", Mnemonic)
	md := md5.Sum([]byte(MnemonicSeed))
	mdStr := hex.EncodeToString(md[:])
	comp := &AccountsComp{
		Name:     name,
		Accounts: accountSli,
	}
	w.KeyRelationStore.Store(mdStr, comp)

	return accountSli, nil
}

func (w *WalletBase) KeyList() ([]AccountsMeta, error) {
	var metaSli []AccountsMeta
	w.KeyRelationStore.Range(func(key, value any) bool {
		a := AccountsMeta{}
		a.Key = key.(string)
		a.AccountsComp = value.(*AccountsComp)
		metaSli = append(metaSli, a)
		return true
	})
	return metaSli, nil
}

func (w *WalletBase) AddAccounts(ctx context.Context, key string, chain string, start int64, count int64) ([]*AccountDetail, error) {
	ctx = context.WithValue(ctx, "chain", chain)
	Mnemonic, err := w.loadMnemonic(ctx, key)
	if err != nil {
		log.ZError(ctx, "AddAccounts loadMnemonic failed", err, "stage", "mnemonic_lookup")
		return nil, err
	}

	handler, err := w.selectHandler(chain)
	if err != nil {
		return nil, err
	}

	var AccountSli []*AccountDetail
	for i := start; i < start+count; i++ {
		addr, er := handler.AccountsFromMnemonicIndex(ctx, Mnemonic, int(i))
		if er != nil {
			return nil, er
		}
		t := &AccountDetail{
			Addr:  addr,
			Index: i,
			Chain: chain,
		}
		AccountSli = append(AccountSli, t)
	}

	return AccountSli, nil
}

func (w *WalletBase) AddDepositAddr(ctx context.Context, addr string, name string) error {
	exist := false
	w.DepositAddrStore.Range(func(key, value any) bool {
		keyStr, _ := key.(string)
		if strings.ToLower(keyStr) == strings.ToLower(addr) {
			exist = true
			return false
		}
		return true
	})
	if exist {
		err := fmt.Errorf("deposit address already exists")
		log.ZError(ctx, "deposit address already exists", err, "stage", "deposit_address_update", "result", "failure", "error_code", "duplicate_address")
		return err
	}

	w.DepositAddrStore.Store(addr, name)
	return nil
}

func (w *WalletBase) ListDepositAddr() ([]*DepositAddrMeta, error) {
	var depositSlice []*DepositAddrMeta
	w.DepositAddrStore.Range(func(key, value any) bool {
		keyStr, _ := key.(string)
		valueStr, _ := value.(string)
		depositSlice = append(depositSlice, &DepositAddrMeta{Addr: keyStr, Name: valueStr})
		return true
	})
	return depositSlice, nil
}

func (w *WalletBase) Sign(ctx context.Context, key string, chain string, chainId string, addr string, index int64, toAddr string, toAddrIndex int64, toAddrBase string, amount string, needSign []byte) ([]byte, error) {
	handler, err := w.selectHandler(chain)
	if err != nil {
		return nil, err
	}

	to, err := handler.SafetyCheck(ctx, addr, toAddr, amount, needSign)
	if err != nil {
		log.ZError(ctx, "Sign SafetyCheck exception", err)
		return nil, err
	}
	log.ZInfo(ctx, "transaction destination validated", "stage", "transaction_safety_check", "result", "success")
	// gas fee transfer Scene safety check , the withdraw addr is derived address
	if toAddrIndex > 0 && toAddrBase != "" {
		_, err = w.derivedAccountCheck(ctx, chain, to, toAddrIndex, toAddrBase)
		if err != nil {
			return nil, err
		}
	} else {
		if !w.withdrawalAddressCheck(to) {
			er := fmt.Errorf("sign data withdraw address is not in whitelist")
			return nil, er
		}
	}

	Mnemonic, err := w.loadMnemonic(ctx, key)
	if err != nil {
		log.ZError(ctx, "sign mnemonic lookup failed", err, "stage", "mnemonic_lookup", "result", "failure", "error_code", "mnemonic_lookup_failed")
		return nil, err
	}

	signRes, er := handler.SignTx(ctx, Mnemonic, addr, int(index), chainId, toAddr, amount, needSign)
	if er != nil {
		return nil, er
	}

	return signRes, nil
}

func (w *WalletBase) derivedAccountCheck(ctx context.Context, chain string, to string, toIndex int64, masterAddr string) (bool, error) {
	acc, err := w.AddAccounts(ctx, masterAddr, chain, toIndex, 1)
	if err != nil {
		log.ZError(ctx, "derivedAccountCheck exception", err)
		return false, err
	}
	for _, v := range acc {
		if v.Addr == to {
			return true, nil
		}
	}
	return false, fmt.Errorf("derived withdrawal address is invalid")
}

func (w *WalletBase) selectHandler(chain string) (WalletBehaviors, error) {
	if chain == ETH_TYP_ACCOUNT {
		return w.EthWallet, nil
	} else if chain == TRON_TYP_ACCOUNT {
		return w.TronWallet, nil
	} else if chain == SOLANA_TYP_ACCOUNT {
		return w.SolanaWallet, nil
	}
	return nil, fmt.Errorf("unsupported chain")
}

func (w *WalletBase) withdrawalAddressCheck(addr string) bool {
	_, isExist := w.DepositAddrStore.Load(addr)
	return isExist
}

// storeMnemonic
// key: master address
func (w *WalletBase) storeMnemonic(ctx context.Context, key, mnemonic string) error {
	encrypt, err := cipher.AESEncryptWithKey(mnemonic, w.aesKey)
	if err != nil {
		return err
	}
	w.KeyRamStore.Store(key, encrypt)
	return nil
}

// loadMnemonic
// key: master address
func (w *WalletBase) loadMnemonic(ctx context.Context, key string) (mnemonic string, err error) {
	encrypt, exist := w.KeyRamStore.Load(key)
	if !exist {
		err = fmt.Errorf("mnemonic not found")
		log.ZError(ctx, "mnemonic lookup failed", err, "stage", "mnemonic_lookup", "result", "failure", "error_code", "mnemonic_not_found")
		return "", err
	}

	MnemonicStr, ok := encrypt.(string)
	if !ok {
		err = fmt.Errorf("stored mnemonic has an invalid format")
		log.ZError(ctx, "stored mnemonic has an invalid format", err, "stage", "mnemonic_lookup", "result", "failure", "error_code", "invalid_mnemonic_record")
		return "", err
	}

	mnemonic, err = cipher.AESDecryptWithKey(MnemonicStr, w.aesKey)
	if err != nil {
		log.ZError(ctx, "KeyRamStore AESDecrypt failed", err)
		return "", err
	}
	return mnemonic, nil
}

type Account struct {
	BaseAddr string
	Chain    string
	List     []SubAccount
}
type SubAccount struct {
	Addr  string
	Index int
}

func (w *WalletBase) CheckAccounts(ctx context.Context, accounts []Account) error {
	var (
		err         error
		mnemonicMap = make(map[string]string)
	)

	for _, a := range accounts {
		chainCtx := context.WithValue(ctx, "chain", a.Chain)
		mnemonic, ok := mnemonicMap[a.BaseAddr]
		if !ok {
			mnemonic, err = w.loadMnemonic(chainCtx, a.BaseAddr)
			if err != nil {
				log.ZError(chainCtx, "CheckAccounts loadMnemonic failed", err, "stage", "mnemonic_lookup")
				return err
			}
			mnemonicMap[a.BaseAddr] = mnemonic
		}
		handler, err := w.selectHandler(a.Chain)
		if err != nil {
			return err
		}
		for j := range a.List {
			err = w.checkAccount(chainCtx, handler, mnemonic, &a.List[j])
			if err != nil {
				log.ZError(chainCtx, "account verification failed", err, "stage", "account_verification", "result", "failure", "error_code", "account_verification_failed")
				return err
			}
		}
	}

	return nil
}

func (w *WalletBase) checkAccount(ctx context.Context, handler WalletBehaviors, mnemonic string, account *SubAccount) error {
	addr, err := handler.AccountsFromMnemonicIndex(ctx, mnemonic, account.Index)
	if err != nil {
		return err
	}
	if account.Addr != addr {
		return fmt.Errorf("account address mismatch")
	}
	return nil
}
