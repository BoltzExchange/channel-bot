package main

import (
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/lnd"
)

type acceptorConfig struct {
	config.BaseConfig

	Lnd *lnd.LND `group:"LND Options"`
}

func (r *acceptorConfig) GetStruct() any {
	return r
}

func (r *acceptorConfig) GetHelp() *config.HelpOptions {
	return r.BaseConfig.Help
}

func (r *acceptorConfig) GetConfigFile() string {
	return r.BaseConfig.ConfigFile
}
