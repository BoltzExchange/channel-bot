package main

import (
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/lnd"
)

type reconnectorConfig struct {
	config.BaseConfig

	Lnd *lnd.LND `group:"LND Options"`

	Interval int `short:"i" long:"interval" description:"Interval in minutes at which peers with inactive channels should be reconnected"`
}

func (r *reconnectorConfig) GetStruct() any {
	return r
}

func (r *reconnectorConfig) GetHelp() *config.HelpOptions {
	return r.BaseConfig.Help
}

func (r *reconnectorConfig) GetConfigFile() string {
	return r.BaseConfig.ConfigFile
}
