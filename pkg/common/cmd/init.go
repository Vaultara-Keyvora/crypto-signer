package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"lbe_crypto_signer/internal/logic"
	"lbe_crypto_signer/pkg/common/config"
	"lbe_crypto_signer/pkg/tools/log"
)

func InitConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Project init before start",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			param, err := parseSignerConfig(cmd)
			if err != nil {
				panic("fail to parse flags, error: " + err.Error())
			}

			err = logic.SignerInitIns.Init(context.Background(), param)
			if err != nil {
				panic("fail to parse flags, error: " + err.Error())
			}
			log.ZInfo(context.Background(), "init signer success")
		},
	}
	cmd.Flags().StringP(FlagConf, "c", "./config", "Path of config directory")

	for flag, usage := range logic.FlagUsageMap {
		shorthand := logic.FlagShorthandMap[flag]
		cmd.Flags().StringP(flag, shorthand, "", usage)
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}

	return cmd
}

func parseSignerConfig(cmd *cobra.Command) (*config.Signer, error) {
	var (
		ctx   = cmd.Context()
		flags = cmd.Flags()
		param = &config.Signer{}
		err   error
	)

	param.EthereumAccount, err = flags.GetString(logic.FlagEthereumAccount)
	if err != nil {
		log.ZError(ctx, "parseInitParam of account err", err)
		return nil, err
	}

	param.MerchantId, err = flags.GetString(logic.FlagMerchantId)
	if err != nil {
		log.ZError(ctx, "parseInitParam of merchantId err", err)
		return nil, err
	}

	param.MerchantUrl, err = flags.GetString(logic.FlagMerchantUrl)
	if err != nil {
		log.ZError(ctx, "parseInitParam of merchantUrl err", err)
		return nil, err
	}

	param.SignerUrl, err = flags.GetString(logic.FlagSignerUrl)
	if err != nil {
		log.ZError(ctx, "parseInitParam of signerUrl err", err)
		return nil, err
	}

	param.SecretKey, err = flags.GetString(logic.FlagSecretKey)
	if err != nil {
		log.ZError(ctx, "parseInitParam of secretKey err", err)
		return nil, err
	}
	return param, nil
}
