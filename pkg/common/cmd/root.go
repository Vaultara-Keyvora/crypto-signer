package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"lbe_crypto_signer/pkg/common/config"
	"lbe_crypto_signer/pkg/tools/errs"
	"lbe_crypto_signer/pkg/tools/log"
	"lbe_crypto_signer/version"
	"path/filepath"
)

type RootCmd struct {
	ctx         context.Context
	Command     cobra.Command
	processName string
	log         config.Log
}

func NewRootCmd(processName string, opts ...func(*CmdOpts)) *RootCmd {
	rootCmd := &RootCmd{processName: processName}
	cmd := cobra.Command{
		Use:  "Start",
		Long: fmt.Sprintf(`Start %s `, processName),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.scanStartPreRun(cmd, opts...)
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.Flags().StringP(FlagConf, "c", "./config", "Path of config directory")
	rootCmd.Command = cmd
	return rootCmd
}

type CmdOpts struct {
	loggerPrefixName string
	configMap        map[string]any
}

func WithConfigMap(configMap map[string]any) func(*CmdOpts) {
	return func(opts *CmdOpts) {
		opts.configMap = configMap
	}
}

func (r *RootCmd) scanStartPreRun(cmd *cobra.Command, opts ...func(*CmdOpts)) error {
	cmdOpts := r.applyOptions(opts...)
	if err := r.initializeConfiguration(cmd, cmdOpts); err != nil {
		return err
	}
	if err := r.initializeLogger(cmdOpts); err != nil {
		return errs.WrapMsg(err, "failed to initialize logger")
	}
	return nil
}

func (r *RootCmd) applyOptions(opts ...func(*CmdOpts)) *CmdOpts {
	cmdOpts := defaultCmdOpts()
	for _, opt := range opts {
		opt(cmdOpts)
	}

	return cmdOpts
}

func (r *RootCmd) initializeLogger(cmdOpts *CmdOpts) error {
	if err := log.InitFromConfig(
		"lbe signer-active",
		r.processName,
		r.log.RemainLogLevel,
		r.log.IsStdout,
		r.log.IsJson,
		r.log.StorageLocation,
		r.log.RemainRotationCount,
		r.log.RotationTime,
		version.Version,
		r.log.IsSimplify,
	); err != nil {
		return err
	}
	return errs.Wrap(log.InitConsoleLogger(r.processName, r.log.RemainLogLevel, r.log.IsJson, version.Version))
}

func (r *RootCmd) initializeConfiguration(cmd *cobra.Command, opts *CmdOpts) error {
	configDirectory, err := cmd.Flags().GetString(FlagConf)
	if err != nil {
		return err
	}
	for configFileName, configStruct := range opts.configMap {
		err := config.LoadConfig(filepath.Join(configDirectory, configFileName), configStruct)
		if err != nil {
			return err
		}
	}
	return config.LoadConfig(filepath.Join(configDirectory, LogConfigFileName), &r.log)
}

func (r *RootCmd) Execute() error {
	return r.Command.Execute()
}

func defaultCmdOpts() *CmdOpts {
	return &CmdOpts{
		loggerPrefixName: "lbe-signer-active-log",
	}
}
