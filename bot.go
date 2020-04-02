package main

import (
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"sync"
)

var lndInfo *lnrpc.GetInfoResponse

func main() {
	cfg := loadConfig()
	initLogger(cfg.LogFile)
	logConfig(cfg)

	initLnd(cfg)
	initDiscord(cfg)

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

func initLnd(cfg *config) {
	logger.Info("Initializing LND client")

	err := cfg.Lnd.Connect()
	checkError("LND", err)

	lndInfo, err = cfg.Lnd.GetInfo()
	checkError("LND", err)

	logger.Info("Initialized LND client: ", stringify(lndInfo))
}

func initDiscord(cfg *config) {
	logger.Info("Initializing Discord client")

	err := cfg.Discord.Init()
	checkError("Discord", err)

	err = cfg.Discord.SendMessage("Started channel bot with LND node: **" + lndInfo.Alias + "** (`" + lndInfo.IdentityPubkey + "`)")
	checkError("Discord", err)

	logger.Info("Initialized Discord client")
}

func checkError(service string, err error) {
	if err != nil {
		logger.Fatal("Could not initalize "+service+": ", err)
	}
}
