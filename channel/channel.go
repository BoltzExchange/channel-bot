package channel

import (
	"strconv"
	"time"

	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type ChannelManager struct {
	Interval int

	Lnd     *lnd.LND
	Discord *discord.Discord

	imbalancedChannels  map[uint64]bool
	significantChannels map[uint64]SignificantChannel
}

type ratios struct {
	min float64
	max float64
}

type SignificantChannel struct {
	Alias     string
	ChannelID uint64
	MinRatio  string
	MaxRatio  string

	ratios ratios
}

func (manager *ChannelManager) Init(significantChannels []*SignificantChannel) {
	logger.Info("Starting channel manager")

	manager.significantChannels = make(map[uint64]SignificantChannel)

	for _, significant := range significantChannels {
		maxRatio, _ := strconv.ParseFloat(significant.MaxRatio, 64)
		minRatio, _ := strconv.ParseFloat(significant.MinRatio, 64)

		ratios := ratios{
			max: maxRatio,
			min: minRatio,
		}

		significant.ratios = ratios

		manager.significantChannels[significant.ChannelID] = *significant
	}

	manager.imbalancedChannels = make(map[uint64]bool)
	manager.checkChannels()

	ticker := time.NewTicker(time.Duration(manager.Interval) * time.Second)

	for range ticker.C {
		manager.checkChannels()
	}
}

func (manager *ChannelManager) checkChannels() {
	logger.Info("Checking significant channels")

	channels, err := manager.Lnd.ListChannels()

	if err != nil {
		logger.Error("Could not get channels: " + err.Error())
		return
	}

	manager.checkSignificantChannels(channels)

	logger.Info("Checking normal channels")

	for _, channel := range channels.Channels {
		_, isSignificant := manager.significantChannels[channel.ChanId]
		if channel.Private || isSignificant {
			continue
		}

		channelRatio := getChannelRatio(channel)

		if channelRatio > 0.3 && channelRatio < 0.7 {
			if contains := manager.imbalancedChannels[channel.ChanId]; contains {
				manager.logChannel(channel, false)
				delete(manager.imbalancedChannels, channel.ChanId)
			}

			continue
		}

		if contains := manager.imbalancedChannels[channel.ChanId]; contains {
			continue
		}

		manager.imbalancedChannels[channel.ChanId] = true
		manager.logChannel(channel, true)
	}
}

func (manager *ChannelManager) checkSignificantChannels(channels *lnrpc.ListChannelsResponse) {
	for _, channel := range channels.Channels {
		significantChannel, isSignificant := manager.significantChannels[channel.ChanId]

		if !isSignificant {
			continue
		}

		channelRatio := getChannelRatio(channel)

		if channelRatio > significantChannel.ratios.max && channelRatio < significantChannel.ratios.min {
			if contains := manager.imbalancedChannels[channel.ChanId]; contains {
				significantChannel.logChannel(manager.Discord, channel, false)
				delete(manager.imbalancedChannels, channel.ChanId)
			}

			continue
		}

		if contains := manager.imbalancedChannels[channel.ChanId]; contains {
			continue
		}

		manager.imbalancedChannels[channel.ChanId] = true
		significantChannel.logChannel(manager.Discord, channel, true)
	}
}

func getChannelRatio(channel *lnrpc.Channel) float64 {
	return float64(channel.LocalBalance) / float64(channel.Capacity)
}

func (significantChannel *SignificantChannel) logChannel(discord *discord.Discord, channel *lnrpc.Channel, isImbalanced bool) {
	var info string
	var emoji string

	if isImbalanced {
		info = "imbalanced"
		emoji = ":rotating_light:"
	} else {
		info = "balanced again"
		emoji = ":zap:"
	}

	message := emoji + " Channel " + significantChannel.Alias + " `" + formatChannelID(channel) + "` is **" + info + "** " + emoji + " :\n"

	localBalance, remoteBalance := formatChannelBalances(channel)
	message += localBalance + "\n"
	message += "    Minimal: " + formatFloat(float64(channel.LocalBalance)*significantChannel.ratios.min) + "\n"
	message += "    Maximal: " + formatFloat(float64(channel.LocalBalance)*significantChannel.ratios.max) + "\n"
	message += remoteBalance

	logger.Info(message)
	_ = discord.SendMessage(message)
}

func (manager *ChannelManager) logChannel(channel *lnrpc.Channel, isImbalanced bool) {
	var info string

	if isImbalanced {
		info = "imbalanced"
	} else {
		info = "balanced again"
	}

	// Let's just ignore that error
	nodeInfo, _ := manager.Lnd.GetNodeInfo(channel.RemotePubkey)

	message := "Channel `" + formatChannelID(channel) + "` to `" + nodeInfo.Node.Alias + "` is **" + info + "**:\n"

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

func formatChannelID(channel *lnrpc.Channel) string {
	return strconv.FormatUint(channel.ChanId, 10)
}
