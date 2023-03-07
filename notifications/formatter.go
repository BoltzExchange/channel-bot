package notifications

import (
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"strconv"
)

func (sc *SignificantChannel) logBalance(discord discord.NotificationService, channel *lnrpc.Channel, isImbalanced bool) {
	var info string
	var emoji string

	if isImbalanced {
		info = "imbalanced"
		emoji = ":rotating_light:"
	} else {
		info = "balanced again"
		emoji = ":zap:"
	}

	message := emoji + " Channel **" + sc.Alias + "** `" + lnd.FormatChannelID(channel.ChanId) + "` is **" + info + "** " + emoji + " :\n"

	localBalance, remoteBalance := formatChannelBalances(channel)
	message += localBalance + "\n"
	message += "    Minimal: " + formatFloat(float64(channel.Capacity)*sc.ratios.min) + "\n"
	message += "    Maximal: " + formatFloat(float64(channel.Capacity)*sc.ratios.max) + "\n"
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

	message := "Channel `" + lnd.FormatChannelID(channel.ChanId) + "` to `" + manager.nc.getNodeName(channel.RemotePubkey) + "` is **" + info + "**:\n"

	localBalance, remoteBalance := formatChannelBalances(channel)
	message += localBalance + "\n" + remoteBalance

	logger.Info(message)
	_ = manager.discord.SendMessage(message)
}

func (sc *SignificantChannel) logSignificantNotFound(discord discord.NotificationService) {
	emoji := ":rotating_light:"
	message := emoji + " Channel **" + sc.Alias + "** `" + lnd.FormatChannelID(sc.ChannelID) + "` couldn't be found " + emoji

	logger.Info(message)
	_ = discord.SendMessage(message)
}

func formatChannelBalances(channel *lnrpc.Channel) (local string, remote string) {
	local = "  Local: " + strconv.FormatInt(channel.LocalBalance, 10)
	remote = "  Remote: " + strconv.FormatInt(channel.RemoteBalance, 10)

	return local, remote
}

func formatFloat(float float64) string {
	return strconv.FormatFloat(float, 'f', 0, 64)
}
