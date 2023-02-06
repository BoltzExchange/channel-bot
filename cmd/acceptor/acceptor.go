package main

import (
	"encoding/hex"
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func main() {
	cfg := config.LoadConfig(&acceptorConfig{
		BaseConfig: config.BaseConfig{
			LogFile:    "./acceptor.log",
			ConfigFile: "./acceptor.toml",
		},
	})
	utils.InitLogger(cfg.BaseConfig.LogFile)
	config.LogConfig(cfg)

	_ = utils.InitLnd(cfg.Lnd)

	accp, err := cfg.Lnd.Client.ChannelAcceptor(cfg.Lnd.Ctx)
	if err != nil {
		logger.Fatalf("Could not start ChannelAcceptor: %s", err.Error())
	}

	for {
		msg, err := accp.Recv()
		handleStreamError(err)
		logger.Infof("Got channel request from: %s", hex.EncodeToString(msg.NodePubkey))

		err = accp.Send(&lnrpc.ChannelAcceptResponse{
			PendingChanId:  msg.PendingChanId,
			Accept:         true,
			MinAcceptDepth: 0,
			ZeroConf:       true,
		})
		handleStreamError(err)
	}
}

func handleStreamError(err error) {
	if err != nil {
		logger.Fatalf("Error in ChannelAcceptor stream: %s", err.Error())
	}
}
