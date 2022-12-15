package utils

import (
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func InitLnd(lnd *lnd.LND) *lnrpc.GetInfoResponse {
	logger.Info("Initializing LND client")

	err := lnd.Connect()
	CheckError("LND", err)

	lndInfo, err := lnd.GetInfo()
	CheckError("LND", err)

	logger.Info("Initialized LND client: ", Stringify(lndInfo))
	return lndInfo
}
