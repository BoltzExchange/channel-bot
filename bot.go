package main

import (
	"github.com/BoltzExchange/channel-bot/cleaner"
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/notifications"
	"github.com/BoltzExchange/channel-bot/utils"
	"sync"

	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func main() {
	cfg := config.LoadConfig(&channelBotConfig{
		BaseConfig: config.BaseConfig{
			LogFile:    "./channel-bot.log",
			ConfigFile: "./channel-bot.toml",
		},

		Notifications: &notifications.ChannelManager{
			Interval: 60,
		},

		ChannelCleaner: &cleaner.ChannelCleaner{
			Interval:           24,
			MaxInactive:        30,
			MaxInactivePrivate: 60,
		},
	})
	utils.InitLogger(cfg.BaseConfig.LogFile)
	config.LogConfig(cfg)

	lndInfo := utils.InitLnd(cfg.Lnd)
	initDiscord(cfg, lndInfo)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		cfg.Notifications.Init(cfg.SignificantChannels, cfg.Lnd, cfg.Discord)
		wg.Done()
	}()

	go func() {
		cfg.ChannelCleaner.Init(cfg.Lnd, cfg.Discord)
		wg.Done()
	}()

	wg.Wait()
	logger.Info("Shutting down")
}

func initDiscord(cfg *channelBotConfig, lndInfo *lnrpc.GetInfoResponse) {
	logger.Info("Initializing Discord client")

	err := cfg.Discord.Init()
	utils.CheckError("Discord", err)

	err = cfg.Discord.SendMessage("Started channel bot with LND node: **" + lndInfo.Alias + "** (`" + lndInfo.IdentityPubkey + "`)")
	utils.CheckError("Discord", err)

	logger.Info("Initialized Discord client")
}
