package apistruct

type SignerTxReq struct {
	Chain       string `json:"chain"`
	ChainID     string `json:"chain_id"`
	BaseAddr    string `json:"base_addr"`
	SignAddr    string `json:"sign_addr"`
	AddrIndex   int64  `json:"addr_index"`
	ToAddr      string `json:"to_addr"`
	ToAddrIndex int64  `json:"to_addr_index"`
	ToAddrBasic string `json:"to_addr_basic"`
	Amount      string `json:"amount"`
	SignData    []byte `json:"sign_data"`
}

type SignerTxResp struct {
	SignTx []byte `json:"sign_tx"`
}
