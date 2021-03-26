package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Configuration struct {
	cfgFile        string
	providerName   string
	providerConfig interface{}
}

func (config *Configuration) initConfig() error {
	if config.cfgFile != "" {
		viper.SetConfigFile(config.cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".trade")
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("can't read config: %s", err)
	}

	config.providerConfig = viper.Get(fmt.Sprint("providers.", config.providerName))
	return nil
}
