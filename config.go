package main

import (
	"github.com/BoltzExchange/channel-bot/cleaner"
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/BoltzExchange/channel-bot/notifications"
)

type channelBotConfig struct {
	config.BaseConfig

	Notifications  *notifications.ChannelManager `group:"Notification Options"`
	ChannelCleaner *cleaner.ChannelCleaner       `group:"Channel Cleaner Options"`

	Lnd     *lnd.LND         `group:"LND Options"`
	Discord *discord.Discord `group:"Discord Options"`

	// This option is only parsed in the TOML config file
	SignificantChannels []*notifications.SignificantChannel
}

func (c *channelBotConfig) GetStruct() any {
	return c
}

func (c *channelBotConfig) GetHelp() *config.HelpOptions {
	return c.BaseConfig.Help
}

func (c *channelBotConfig) GetConfigFile() string {
	return c.BaseConfig.ConfigFile
}
