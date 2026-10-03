package config

import (
	"github.com/spf13/viper"
)

func LoadConfig(path string, config any) error {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(config); err != nil {
		return err
	}

	return nil
}
