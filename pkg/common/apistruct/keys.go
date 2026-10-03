package apistruct

type KeysAddMnemonicReq struct {
	Mnemonic    string `json:"mnemonic" binding:"required"`
	Name        string `json:"name"  binding:"required"`
	Fingerprint string `json:"fingerprint"  binding:"required"` // 指纹
}

type KeysAddMnemonicResp struct {
	Accounts []AccountBasic `json:"accounts"`
}

type AccountBasic struct {
	Addr  string `json:"addr"`
	Chain string `json:"chain"`
}

type KeysListReq struct{}

type KeysListResp struct {
	KeySlice []KeyMeta `json:"keySlice"`
}

type KeyMeta struct {
	KeyFLag  string         `json:"keyFLag"`
	Name     string         `json:"name"`
	Accounts []AccountBasic `json:"accounts"`
}

type AddAccountsReq struct {
	BaseAddr   string `json:"baseAddr"  binding:"required"`
	Chain      string `json:"chain"  binding:"required"`
	StartIndex int64  `json:"startIndex"  binding:"required"`
	Count      int64  `json:"count"  binding:"required"`
}

type AddAccountsResp struct {
	BaseAddr string           `json:"baseAddr"`
	Adds     []*AccountDetail `json:"adds"`
}

type AddDepositAddrReq struct {
	Addr string `json:"addr"  binding:"required"`
	Name string `json:"name"  binding:"required"`
}

type AddDepositAddrResp struct {
	Addr string `json:"baseAddr"`
}

type ListDepositAddrReq struct{}

type ListDepositAddrResp struct {
	DepositAccounts []*DepositAddrMeta `json:"depositAccounts"`
}

type DepositAddrMeta struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type AccountDetail struct {
	Addr  string `json:"addr"`
	Index int64  `json:"index"`
	Chain string `json:"chain"`
}

type CheckAccountsReq struct {
	Accounts []Accounts `json:"accounts"`
}

type Accounts struct {
	BaseAddr string       `json:"baseAddr"  binding:"required"`
	Chain    string       `json:"chain"  binding:"required"`
	List     []SubAccount `json:"list"`
}

type SubAccount struct {
	Addr  string `json:"addr"  binding:"required"`
	Index uint64 `json:"index"`
}

type CheckAccountsResp struct {
}
