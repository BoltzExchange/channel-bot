package notifications

import (
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func (manager *ChannelManager) logClosedChannel(channel *lnrpc.ChannelCloseSummary) {
	closeType := "closed"

	if channel.CloseType != lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE {
		closeType = "**force closed** :rotating_light:"
	}

	message := "Channel `" + lnd.FormatChannelID(channel.ChanId) + "` to `" +
		manager.nc.getNodeName(channel.RemotePubkey) + "` was " + closeType

	logger.Info(message)
	_ = manager.discord.SendMessage(message)
}
