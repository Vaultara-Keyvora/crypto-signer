package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"lbe_crypto_signer/internal/api"
	"lbe_crypto_signer/internal/service"
	"lbe_crypto_signer/pkg/tools/system/program"
	"lbe_crypto_signer/version"
)

type LBESignerMachineCmd struct {
	*RootCmd
	ctx          context.Context
	activeConfig *service.ActiveConfig
}

func NewStartCmd() *LBESignerMachineCmd {
	var msgConfig service.ActiveConfig
	insBlockChain := &LBESignerMachineCmd{activeConfig: &msgConfig}
	configMap := map[string]any{
		ApiCfgFileName:    &msgConfig.ApiConf,
		SignerCfgFileName: &msgConfig.SignerConf,
	}
	insBlockChain.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(configMap))
	insBlockChain.ctx = context.WithValue(context.Background(), "version", version.Version)
	insBlockChain.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return insBlockChain.runE()
	}

	insBlockChain.RootCmd.Command.AddCommand(InitConfigCommand())
	return insBlockChain
}

func (s *LBESignerMachineCmd) Exec() error {
	return s.Execute()
}

func (s *LBESignerMachineCmd) runE() error {
	return api.Start(s.ctx, s.activeConfig)
}
