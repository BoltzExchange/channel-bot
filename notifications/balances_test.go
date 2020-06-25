package notifications

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCheckBalances(t *testing.T) {
	cleanUp()
	initChannelManager()
	channelManager.imbalancedChannels = map[uint64]bool{}
	channelManager.significantChannels = map[uint64]SignificantChannel{}

	channels = []*lnrpc.Channel{
		{
			ChanId:           123,
			UnsettledBalance: 0,
			LocalBalance:     71,
			Capacity:         100,
		},
	}

	// Should send a notification because the channel has too much outbound and no notification has been sent yet
	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 1)
	assert.True(t, channelManager.imbalancedChannels[channels[0].ChanId])

	delete(channelManager.imbalancedChannels, channels[0].ChanId)

	cleanUp()

	// Should send a notification because the channel has too much inbound an no notification has been sent yet
	channels[0].LocalBalance = 29
	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 1)
	assert.True(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should not send a notification although the channel is imbalanced because a notification has been sent already
	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 0)

	cleanUp()

	// Should send a notification when the channel is balanced again
	channels[0].LocalBalance = 50
	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 1)
	assert.False(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should send the notification for a channel that is balanced again only once
	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 0)
	assert.False(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should ignore channels with unsettled balances
	channels[0].LocalBalance = 0
	channels[0].UnsettledBalance = 1

	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 0)

	channels[0].UnsettledBalance = 0
	cleanUp()

	// Should ignore private channels
	channels[0].Private = true

	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 0)

	channels[0].Private = false
	cleanUp()

	// Should ignore channels that are are significant
	channelManager.significantChannels[channels[0].ChanId] = SignificantChannel{
		ratios: ratios{
			min: -1,
			max: 1,
		},
	}

	channelManager.checkBalances(false)

	assert.Len(t, sentMessages, 0)

	cleanUp()

	// Should only check significant channels on startup
	channelManager.significantChannels = map[uint64]SignificantChannel{}

	channelManager.checkBalances(true)

	assert.Len(t, sentMessages, 0)
	assert.True(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()
}

func TestCheckSignificantChannelBalances(t *testing.T) {
	cleanUp()
	channelManager.imbalancedChannels = map[uint64]bool{}
	channelManager.significantChannels = map[uint64]SignificantChannel{}

	channels := []*lnrpc.Channel{
		{
			ChanId:           123,
			Capacity:         100,
			LocalBalance:     70,
			UnsettledBalance: 0,
		},
	}
	channelManager.significantChannels[123] = SignificantChannel{
		ChannelID: 123,
		ratios: ratios{
			max: 0.6,
			min: 0.4,
		},
	}

	// Should send a notification because the channel has too much outbound and no notification has been sent yet
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 1)
	assert.True(t, channelManager.imbalancedChannels[channels[0].ChanId])

	delete(channelManager.imbalancedChannels, channels[0].ChanId)

	cleanUp()

	// Should send a notification because the channel has too much inbound an no notification has been sent yet
	channels[0].LocalBalance = 30
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 1)
	assert.True(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should not send a notification although the channel is imbalanced because a notification has been sent already
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 0)

	cleanUp()

	// Should send a notification when the channel is balanced again
	channels[0].LocalBalance = 50
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 1)
	assert.False(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should send the notification for a channel that is balanced again only once
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 0)
	assert.False(t, channelManager.imbalancedChannels[channels[0].ChanId])

	cleanUp()

	// Should ignore channels with unsettled balances
	channels[0].LocalBalance = 0
	channels[0].UnsettledBalance = 1

	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 0)

	channels[0].UnsettledBalance = 0
	cleanUp()

	// Should ignore channels that are not significant
	channels[0].LocalBalance = 50
	channels = append(channels, &lnrpc.Channel{
		ChanId:           321,
		Capacity:         100,
		LocalBalance:     70,
		UnsettledBalance: 0,
	})

	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 0)

	cleanUp()

	// Should send a notification in case a significant channel that cannot be found
	channels[0].ChanId = 321
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 1)
	assert.True(t, channelManager.notFoundSignificantChannel[123])

	cleanUp()

	// Should not send a notification for a significant channel that cannot be found twice
	channelManager.checkSignificantChannelBalances(channels)

	assert.Len(t, sentMessages, 0)

	cleanUp()
}

func TestGetChannelRatio(t *testing.T) {
	channel := &lnrpc.Channel{
		Capacity:     1000,
		LocalBalance: 600,
	}

	assert.Equal(t, getChannelRatio(channel), 0.6)
}

func checkLogs(t *testing.T, message string) {
	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.Equal(t, sentMessages[0], message)
	assert.True(t, strings.HasSuffix(loggedMessages[0], message+"\n"))
}

func TestLogSignificantBalance(t *testing.T) {
	cleanUp()

	significantChannel := &SignificantChannel{
		Alias: "Boltz",
		ratios: ratios{
			min: 0.2,
			max: 0.8,
		},
	}

	channel := &lnrpc.Channel{
		ChanId:        123,
		RemotePubkey:  "pubkey",
		Capacity:      155441521769,
		LocalBalance:  32120398448,
		RemoteBalance: 123321123321,
	}

	// Imbalanced
	message := ":rotating_light: Channel **Boltz** `123` is **imbalanced** :rotating_light: :\n  Local: 32120398448\n    Minimal: 31088304354\n    Maximal: 124353217415\n  Remote: 123321123321"
	significantChannel.logBalance(&MockDiscordClient{}, channel, true)

	checkLogs(t, message)

	cleanUp()

	// Balanced
	message = strings.Replace(message, "imbalanced", "balanced again", 1)
	message = strings.Replace(message, ":rotating_light:", ":zap:", 2)

	significantChannel.logBalance(&MockDiscordClient{}, channel, false)

	checkLogs(t, message)

	cleanUp()
}

func TestLogBalance(t *testing.T) {
	cleanUp()

	channel := &lnrpc.Channel{
		ChanId:        123,
		RemotePubkey:  "pubkey",
		LocalBalance:  32120398448,
		RemoteBalance: 123321123321,
	}

	// Imbalanced
	message := "Channel `123` to `pubkey` is **imbalanced**:\n  Local: 32120398448\n  Remote: 123321123321"
	channelManager.logBalance(channel, true)

	checkLogs(t, message)

	cleanUp()

	// Balanced
	message = strings.Replace(message, "imbalanced", "balanced again", 1)
	channelManager.logBalance(channel, false)

	checkLogs(t, message)

	cleanUp()
}

func TestLogSignificantNotFound(t *testing.T) {
	cleanUp()

	significantChannel := &SignificantChannel{
		Alias:     "Boltz",
		ChannelID: 123,
		ratios: ratios{
			min: 0.2,
			max: 0.8,
		},
	}

	message := ":rotating_light: Channel **Boltz** `123` couldn't be found :rotating_light:"
	significantChannel.logSignificantNotFound(&MockDiscordClient{})

	checkLogs(t, message)

	cleanUp()
}

func TestFormatChannelBalances(t *testing.T) {
	channel := &lnrpc.Channel{
		LocalBalance:  321,
		RemoteBalance: 123,
	}

	local, remote := formatChannelBalances(channel)

	assert.Equal(t, local, "  Local: 321")
	assert.Equal(t, remote, "  Remote: 123")
}

func TestFormatFloat(t *testing.T) {
	assert.Equal(t, formatFloat(230948), "230948")

	// Should ignore decimals
	assert.Equal(t, formatFloat(0.1), "0")
}
