package apistruct

type AdminLoginReq struct {
	Address string `json:"address"`
	Seeds   string `json:"seeds"`
	Sign    string `json:"sign"`
}

type AdminLoginResp struct {
	Token       string `json:"token"`
	Fingerprint string `json:"fingerprint"` // 指纹
	PublicKey   string `json:"public_key"`  // 公钥
}

type AdminLoginOutReq struct {
	Token string `json:"token"`
}

type AdminLoginOutResp struct{}
