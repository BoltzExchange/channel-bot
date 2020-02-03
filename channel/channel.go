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

	imbalancedChannels map[uint64]bool
}

func (manager *ChannelManager) Init() {
	logger.Info("Starting channel manager")

	manager.imbalancedChannels = make(map[uint64]bool)
	manager.checkChannels()

	ticker := time.NewTicker(time.Duration(manager.Interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			manager.checkChannels()
		}
	}
}

func (manager *ChannelManager) checkChannels() {
	logger.Info("Checking normal channels")
	channels, _ := manager.Lnd.ListChannels()

	for _, channel := range channels.Channels {
		if channel.Private {
			continue
		}

		channelRatio := float64(channel.LocalBalance) / float64(channel.Capacity)

		if channelRatio > 0.3 && channelRatio < 0.7 {
			if contains := manager.imbalancedChannels[channel.ChanId]; contains {
				logBalancedChannel(manager.Discord, channel)
				delete(manager.imbalancedChannels, channel.ChanId)
			}

			continue
		}

		if contains := manager.imbalancedChannels[channel.ChanId]; contains {
			continue
		}

		manager.imbalancedChannels[channel.ChanId] = true
		logImbalancedChannel(manager.Discord, channel)
	}
}

func logImbalancedChannel(discord *discord.Discord, channel *lnrpc.Channel) {
	message := "Channel `" + strconv.FormatUint(channel.ChanId, 10) + "` is **imbalanced**:\n"
	message += "  Local: " + strconv.FormatInt(channel.LocalBalance, 10) + "\n"
	message += "  Remote: " + strconv.FormatInt(channel.RemoteBalance, 10)

	logger.Info(message)
	discord.SendMessage(message)
}

func logBalancedChannel(discord *discord.Discord, channel *lnrpc.Channel) {
	message := "Channel `" + strconv.FormatUint(channel.ChanId, 10) + "` is **balanced again**:\n"
	message += "  Local: " + strconv.FormatInt(channel.LocalBalance, 10) + "\n"
	message += "  Remote: " + strconv.FormatInt(channel.RemoteBalance, 10)

	logger.Info(message)
	discord.SendMessage(message)
}
