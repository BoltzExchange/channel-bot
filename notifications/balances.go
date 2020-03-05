package notifications

import (
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"strconv"
)

func (manager *ChannelManager) checkBalances() {
	logger.Info("Checking significant channel balances")

	channels, err := manager.Lnd.ListChannels()

	if err != nil {
		logger.Error("Could not get channels: " + err.Error())
		return
	}

	manager.checkSignificantChannelBalances(channels)

	logger.Info("Checking normal channel balances")

	for _, channel := range channels.Channels {
		_, isSignificant := manager.significantChannels[channel.ChanId]

		if channel.Private || isSignificant {
			continue
		}

		channelRatio := getChannelRatio(channel)

		if channelRatio > 0.3 && channelRatio < 0.7 {
			if contains := manager.imbalancedChannels[channel.ChanId]; contains {
				manager.logBalance(channel, false)
				delete(manager.imbalancedChannels, channel.ChanId)
			}

			continue
		}

		if contains := manager.imbalancedChannels[channel.ChanId]; contains {
			continue
		}

		manager.imbalancedChannels[channel.ChanId] = true
		manager.logBalance(channel, true)
	}
}

func (manager *ChannelManager) checkSignificantChannelBalances(channels *lnrpc.ListChannelsResponse) {
	for _, channel := range channels.Channels {
		significantChannel, isSignificant := manager.significantChannels[channel.ChanId]

		if !isSignificant {
			continue
		}

		channelRatio := getChannelRatio(channel)

		if channelRatio > significantChannel.ratios.max && channelRatio < significantChannel.ratios.min {
			if contains := manager.imbalancedChannels[channel.ChanId]; contains {
				significantChannel.logBalance(manager.Discord, channel, false)
				delete(manager.imbalancedChannels, channel.ChanId)
			}

			continue
		}

		if contains := manager.imbalancedChannels[channel.ChanId]; contains {
			continue
		}

		manager.imbalancedChannels[channel.ChanId] = true
		significantChannel.logBalance(manager.Discord, channel, true)
	}
}

func getChannelRatio(channel *lnrpc.Channel) float64 {
	return float64(channel.LocalBalance) / float64(channel.Capacity)
}

func (significantChannel *SignificantChannel) logBalance(discord *discord.Discord, channel *lnrpc.Channel, isImbalanced bool) {
	var info string
	var emoji string

	if isImbalanced {
		info = "imbalanced"
		emoji = ":rotating_light:"
	} else {
		info = "balanced again"
		emoji = ":zap:"
	}

	message := emoji + " Channel " + significantChannel.Alias + " `" + formatChannelID(channel.ChanId) + "` is **" + info + "** " + emoji + " :\n"

	localBalance, remoteBalance := formatChannelBalances(channel)
	message += localBalance + "\n"
	message += "    Minimal: " + formatFloat(float64(channel.LocalBalance)*significantChannel.ratios.min) + "\n"
	message += "    Maximal: " + formatFloat(float64(channel.LocalBalance)*significantChannel.ratios.max) + "\n"
	message += remoteBalance

	logger.Info(message)
	_ = discord.SendMessage(message)
}

func (manager *ChannelManager) logBalance(channel *lnrpc.Channel, isImbalanced bool) {
	var info string

	if isImbalanced {
		info = "imbalanced"
	} else {
		info = "balanced again"
	}

	nodeInfo, err := manager.Lnd.GetNodeInfo(channel.RemotePubkey)

	if err != nil {
		logger.Error("Could not get node info: " + err.Error())
		return
	}

	message := "Channel `" + formatChannelID(channel.ChanId) + "` to `" + nodeInfo.Node.Alias + "` is **" + info + "**:\n"

	localBalance, remoteBalance := formatChannelBalances(channel)
	message += localBalance + "\n" + remoteBalance

	logger.Info(message)
	_ = manager.Discord.SendMessage(message)
}

func formatFloat(float float64) string {
	return strconv.FormatFloat(float, 'f', 0, 64)
}

func formatChannelBalances(channel *lnrpc.Channel) (local string, remote string) {
	local = "  Local: " + strconv.FormatInt(channel.LocalBalance, 10)
	remote = "  Remote: " + strconv.FormatInt(channel.RemoteBalance, 10)

	return local, remote
}
