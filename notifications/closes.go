package notifications

import (
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func (manager *ChannelManager) checkClosedChannels() {
	logger.Info("Checking closed channels")

	closedChannels, err := manager.Lnd.ClosedChannels()

	if err != nil {
		logger.Error("Could not get closed channels: " + err.Error())
		return
	}

	for _, channel := range closedChannels.Channels {
		// Do not send notifications for channels that were closed before the bot was started
		if channel.CloseHeight < manager.startupHeight {
			continue
		}

		// Do not send notifications for a closed channel more than once
		if sentAlready := manager.closedChannels[channel.ChanId]; sentAlready {
			continue
		}

		manager.logClosedChannel(channel)
		manager.closedChannels[channel.ChanId] = true
	}
}

func (manager *ChannelManager) logClosedChannel(channel *lnrpc.ChannelCloseSummary) {
	nodeName := channel.RemotePubkey

	nodeInfo, err := manager.Lnd.GetNodeInfo(channel.RemotePubkey)

	// Use the alias if it can be queried
	if err == nil {
		nodeName = nodeInfo.Node.Alias
	}

	closeType := "closed"

	if channel.CloseType != lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE {
		closeType = "**force closed** :rotating_light:"
	}

	message := "Channel `" + formatChannelID(channel.ChanId) + "` to `" + nodeName + "` was " + closeType

	logger.Info(message)
	_ = manager.Discord.SendMessage(message)
}
