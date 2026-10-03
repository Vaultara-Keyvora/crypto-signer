package cmd

const (
	FlagConf = "config-path"
)

var (
	LogConfigFileName string
	ApiCfgFileName    string
	SignerCfgFileName string
)

func init() {
	LogConfigFileName = "log.yaml"
	ApiCfgFileName = "api.yaml"
	SignerCfgFileName = "signer.yaml"
}
