package notifications

import (
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func checkParseSignificantChannel(t *testing.T, cm *ChannelManager, unparsedChannel *SignificantChannel) {
	parsedChannel := cm.sm.significantChannels[unparsedChannel.ChannelID]

	assert.Equal(t, unparsedChannel.Alias, parsedChannel.Alias)
	assert.Equal(t, unparsedChannel.ChannelID, parsedChannel.ChannelID)
	assert.Equal(t, unparsedChannel.MinRatio, strconv.FormatFloat(parsedChannel.ratios.min, 'f', 1, 64))
	assert.Equal(t, unparsedChannel.MaxRatio, strconv.FormatFloat(parsedChannel.ratios.max, 'f', 1, 64))
}

func TestParseSignificantChannels(t *testing.T) {
	significantChannels := []*SignificantChannel{
		{
			Alias:     "Test",
			ChannelID: 842390,
			MinRatio:  "0.1",
			MaxRatio:  "0.9",
		},
		{
			Alias:     "Boltz",
			ChannelID: 592857,
			MinRatio:  "0.4",
			MaxRatio:  "0.6",
		},
	}

	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:                  &MockLndClient{},
		notificationProvider: &MockDiscordClient{},
		nc:                   initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	channelManager.sm = initStateManager(channelManager, significantChannels)

	checkParseSignificantChannel(t, channelManager, significantChannels[0])
	checkParseSignificantChannel(t, channelManager, significantChannels[1])
}

func TestPopulateChannels(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:                  &MockLndClient{},
		notificationProvider: &MockDiscordClient{},
		nc:                   initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	significants := []*SignificantChannel{
		{
			Alias:     "Test",
			ChannelID: 842390,
			MinRatio:  "0.1",
			MaxRatio:  "0.9",
		},
		{
			Alias:     "NotFound",
			ChannelID: 413231,
			MinRatio:  "0.1",
			MaxRatio:  "0.9",
		},
	}
	sm := initStateManager(channelManager, significants)

	chs := &lnrpc.ListChannelsResponse{
		Channels: []*lnrpc.Channel{
			{
				ChanId: 123,
			},
			{
				ChanId: significants[0].ChannelID,
			},
		},
	}

	sm.populateChannels(chs)

	assert.Len(t, sm.channels, len(chs.Channels))

	assert.Len(t, sentMessages, 2)
	assert.Len(t, loggedMessages, 2)

	for _, substr := range []string{"imbalanced", "couldn't be found"} {
		assert.True(t, strings.Contains(loggedMessages[0], substr) || strings.Contains(loggedMessages[1], substr))
	}
}

func TestHandleOpen(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:                  &MockLndClient{},
		notificationProvider: &MockDiscordClient{},
		nc:                   initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	sm := initStateManager(channelManager, nil)

	channel := &lnrpc.Channel{
		ChanId:  123321123,
		Private: true,
	}

	sm.handleOpen(channel)

	assert.NotNil(t, sm.channels[channel.ChanId])
}

func TestHandleClose(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd: &MockLndClient{
			nodeAlias: "someNode",
		},
		notificationProvider: &MockDiscordClient{},
	}
	channelManager.nc = initNodeCache(channelManager.lnd, &utils.Clock{})

	significants := []*SignificantChannel{{
		Alias:     "Test",
		ChannelID: 842390,
		MinRatio:  "0.1",
		MaxRatio:  "0.9",
	}}
	sm := initStateManager(channelManager, significants)

	chs := &lnrpc.ListChannelsResponse{
		Channels: []*lnrpc.Channel{
			{
				ChanId: 123,
			},
			{
				ChanId: significants[0].ChannelID,
			},
		},
	}
	sm.populateChannels(chs)

	cleanUp()

	sm.handleClose(&lnrpc.ChannelCloseSummary{
		ChanId:    chs.Channels[0].ChanId,
		CloseType: lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE,
	})

	assert.Nil(t, sm.channels[chs.Channels[0].ChanId])
	assert.False(t, sm.imbalancedChannels[chs.Channels[0].ChanId])

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.True(t, strings.Contains(loggedMessages[0], "was closed"))

	cleanUp()

	sm.handleClose(&lnrpc.ChannelCloseSummary{
		ChanId:    chs.Channels[1].ChanId,
		CloseType: lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE,
	})

	assert.Nil(t, sm.channels[chs.Channels[1].ChanId])
	assert.False(t, sm.imbalancedChannels[chs.Channels[1].ChanId])

	assert.Len(t, sentMessages, 2)
	assert.Len(t, loggedMessages, 2)

	assert.True(t, strings.Contains(loggedMessages[0], "was closed"))
	assert.True(t, strings.Contains(loggedMessages[1], "couldn't be found"))

	cleanUp()
}

func TestHandleHtlc(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:                  &MockLndClient{},
		notificationProvider: &MockDiscordClient{},
		nc:                   initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	sm := initStateManager(channelManager, nil)

	sm.handleHtlc(0, false, 0)

	channel := &lnrpc.Channel{
		ChanId:        567234,
		LocalBalance:  50,
		RemoteBalance: 50,
		Capacity:      100,
	}
	sm.handleOpen(channel)
	sm.handleHtlc(channel.ChanId, true, 50000)

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.True(t, strings.Contains(loggedMessages[0], "imbalanced"))

	cleanUp()
}

func TestHandleSettledInvoice(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:                  &MockLndClient{},
		notificationProvider: &MockDiscordClient{},
		nc:                   initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	sm := initStateManager(channelManager, nil)

	sm.handleHtlc(0, false, 0)

	channel := &lnrpc.Channel{
		ChanId:        567234,
		LocalBalance:  50,
		RemoteBalance: 50,
		Capacity:      100,
	}
	sm.handleOpen(channel)

	sm.handleSettledInvoice(&lnrpc.Invoice{
		Htlcs: []*lnrpc.InvoiceHTLC{
			{
				ChanId:  channel.ChanId,
				AmtMsat: uint64(channel.RemoteBalance) * 1000,
			},
		},
	})

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.True(t, strings.Contains(loggedMessages[0], "imbalanced"))

	cleanUp()
}

func TestCheckChannel(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd: &MockLndClient{
			nodeAlias: "someNode",
		},
		notificationProvider: &MockDiscordClient{},
	}
	channelManager.nc = initNodeCache(channelManager.lnd, &utils.Clock{})

	significants := []*SignificantChannel{{
		Alias:     "Test",
		ChannelID: 842390,
		MinRatio:  "0.1",
		MaxRatio:  "0.9",
	}}
	sm := initStateManager(channelManager, significants)

	chs := &lnrpc.ListChannelsResponse{
		Channels: []*lnrpc.Channel{
			{
				ChanId: 123,
			},
			{
				ChanId: significants[0].ChannelID,
			},
		},
	}
	sm.populateChannels(chs)

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.True(t, strings.Contains(loggedMessages[0], "imbalanced"))

	cleanUp()

	sm.checkChannel(&lnrpc.Channel{
		ChanId:  123,
		Private: true,
	})

	sm.checkChannel(&lnrpc.Channel{
		ChanId:           123,
		UnsettledBalance: 12,
	})

	assert.Len(t, sentMessages, 0)
	assert.Len(t, loggedMessages, 0)

	cleanUp()

	sm.checkChannel(&lnrpc.Channel{
		ChanId:       significants[0].ChannelID,
		LocalBalance: 5,
		Capacity:     10,
	})

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 1)

	assert.True(t, strings.Contains(loggedMessages[0], "balanced"))
}

func TestUpdateChannelBalances(t *testing.T) {
	channel := &lnrpc.Channel{
		LocalBalance:  23234,
		RemoteBalance: 89554,
	}
	localBalanceBefore := channel.LocalBalance
	remoteBalanceBefore := channel.RemoteBalance

	amountMsat := uint64(1231)
	updateChannelBalances(channel, true, amountMsat)

	assert.Equal(t, localBalanceBefore+int64(amountMsat/1000), channel.LocalBalance)
	assert.Equal(t, remoteBalanceBefore-int64(amountMsat/1000), channel.RemoteBalance)

	localBalanceBefore = channel.LocalBalance
	remoteBalanceBefore = channel.RemoteBalance

	amountMsat = uint64(63212)
	updateChannelBalances(channel, false, amountMsat)

	assert.Equal(t, localBalanceBefore-int64(amountMsat/1000), channel.LocalBalance)
	assert.Equal(t, remoteBalanceBefore+int64(amountMsat/1000), channel.RemoteBalance)
}

func TestGetChannelRatio(t *testing.T) {
	channel := &lnrpc.Channel{
		Capacity:     1000,
		LocalBalance: 600,
	}

	assert.Equal(t, getChannelRatio(channel), 0.6)
}
