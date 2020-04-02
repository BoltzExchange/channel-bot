package cleaner

import (
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"math"
	"strconv"
	"time"
)

// TODO: should significant channels be ignored?
type ChannelCleaner struct {
	Interval int `long:"cleaner.interval" description:"Interval in hours at which inactive channels should be checked and possibly closed. Set to 0 to disable this feature"`

	MaxInactive        int `long:"cleaner.maxinactive" description:"After how many days of inactivity a public channel should be force closed"`
	MaxInactivePrivate int `long:"cleaner.maxinactiveprivate" description:"After how many days of inactivity a private channel should be force closed"`

	lnd     lnd.LightningClient
	discord discord.NotificationService

	ticker *time.Ticker
}

func (cleaner *ChannelCleaner) Init(lnd lnd.LightningClient, discord discord.NotificationService) {
	if cleaner.Interval == 0 {
		return
	}

	logger.Info("Starting channel cleaner")

	cleaner.lnd = lnd
	cleaner.discord = discord

	cleaner.forceCloseChannels()

	cleaner.ticker = time.NewTicker(time.Duration(cleaner.Interval) * time.Hour)

	for range cleaner.ticker.C {
		cleaner.forceCloseChannels()
	}
}

func (cleaner *ChannelCleaner) forceCloseChannels() {
	logger.Info("Cleaning inactive channels")

	channels, err := cleaner.lnd.ListInactiveChannels()

	if err != nil {
		logger.Error("Could not get channels: " + err.Error())
		return
	}

	now := time.Now()

	maxInactivePublic := time.Duration(cleaner.MaxInactive) * time.Hour * 24
	maxInactivePrivate := time.Duration(cleaner.MaxInactivePrivate) * time.Hour * 24

	for _, channel := range channels.Channels {
		// Get the channel info from LND to find out the last time the channel was active
		channelInfo, err := cleaner.lnd.GetChannelInfo(channel.ChanId)

		if err != nil {
			logger.Error("Could not get channel info: " + err.Error())
			return
		}

		lastUpdate := channelInfo.Node1Policy.LastUpdate

		if lastUpdate < channelInfo.Node2Policy.LastUpdate {
			lastUpdate = channelInfo.Node2Policy.LastUpdate
		}

		lastUpdateTime := time.Unix(int64(lastUpdate), 0)

		var shouldBeClosed bool

		if channel.Private {
			shouldBeClosed = lastUpdateTime.Add(maxInactivePrivate).Before(now)
		} else {
			shouldBeClosed = lastUpdateTime.Add(maxInactivePublic).Before(now)
		}

		if !shouldBeClosed {
			continue
		}

		cleaner.logClosingChannels(channel, lastUpdateTime)

		// TODO: handle close client
		_, err = cleaner.lnd.ForceCloseChannel(channel.ChannelPoint)

		if err != nil {
			logger.Error("Could not close channel " + lnd.FormatChannelID(channel.ChanId) + ": " + err.Error())
			return
		}
	}
}

func (cleaner *ChannelCleaner) logClosingChannels(channel *lnrpc.Channel, lastUpdate time.Time) {
	channelType := "public"

	if channel.Private {
		channelType = "private"
	}

	lastUpdateDelta := int(math.Round(time.Since(lastUpdate).Hours() / 24))

	message := "Force closing " + channelType + " channel `" + lnd.FormatChannelID(channel.ChanId) + "` to `" + lnd.GetNodeName(cleaner.lnd, channel.RemotePubkey) +
		"` because it was inactive for " + strconv.Itoa(lastUpdateDelta) + " days"

	logger.Info(message)
	_ = cleaner.discord.SendMessage(message)
}
