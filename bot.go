package main

import (
	"github.com/BoltzExchange/channel-bot/notifications/providers"
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/lightningnetwork/lnd/lnrpc"
	"sync"

	"github.com/google/logger"
)

func main() {
	cfg := loadConfig()
	initLogger(cfg.LogFile)
	logConfig(cfg)

	lndInfo := initLnd(cfg)
	provider := getNotificationProvider(cfg)

	err := provider.SendMessage("Started channel bot with LND node: **" + lndInfo.Alias + "** (`" + lndInfo.IdentityPubkey + "`)")
	checkError("Notification provider", err)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		cfg.Notifications.Init(cfg.SignificantChannels, cfg.Lnd, provider)
		wg.Done()
	}()

	go func() {
		cfg.ChannelCleaner.Init(cfg.Lnd, provider)
		wg.Done()
	}()

	wg.Wait()
	logger.Info("Shutting down")
}

func initLnd(cfg *config) *lnrpc.GetInfoResponse {
	logger.Info("Initializing LND client")

	err := cfg.Lnd.Connect()
	checkError("LND", err)

	lndInfo, err := cfg.Lnd.GetInfo()
	checkError("LND", err)

	lndInfo.Features = nil
	logger.Info("Initialized LND client: ", utils.Stringify(lndInfo))

	return lndInfo
}

func getNotificationProvider(cfg *config) providers.NotificationProvider {
	logger.Info("Initializing notification provider client")

	for _, provider := range []providers.NotificationProvider{
		cfg.Mattermost,
		cfg.Discord,
	} {
		initialized := initNotificationProvider(provider)
		if initialized != nil {
			return initialized
		}
	}

	return cfg.Discord
}

func initNotificationProvider(provider providers.NotificationProvider) providers.NotificationProvider {
	err := provider.Init()
	if err != nil {
		logger.Warningf("Could not init %s: %s\n", provider.Name(), err.Error())
		return nil
	}

	return provider
}

func checkError(service string, err error) {
	if err != nil {
		logger.Fatal("Could not initialize "+service+": ", err)
	}
}
