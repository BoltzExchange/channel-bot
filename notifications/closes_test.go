package notifications

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCheckClosedChannels(t *testing.T) {
	initChannelManager()
	cleanUp()

	var closedChannelId uint64

	// Should send a notification
	closedChannels = []*lnrpc.ChannelCloseSummary{
		{
			ChanId:      123,
			CloseHeight: blockHeight + 1,
		},
	}
	closedChannelId = closedChannels[0].ChanId
	channelManager.checkClosedChannels()

	assert.Len(t, sentMessages, 1)
	assert.Len(t, loggedMessages, 2)
	assert.True(t, strings.HasSuffix(loggedMessages[0], "Checking closed channels\n"))

	assert.Len(t, channelManager.closedChannels, 1)
	assert.True(t, channelManager.closedChannels[closedChannels[0].ChanId])

	cleanUp()

	// Should not send a notification because that channel was closed before the bot was started
	closedChannels = []*lnrpc.ChannelCloseSummary{
		{
			ChanId:      321,
			CloseHeight: blockHeight - 1,
		},
	}
	channelManager.checkClosedChannels()

	assert.Len(t, sentMessages, 0)
	assert.False(t, channelManager.closedChannels[closedChannels[0].ChanId])

	cleanUp()

	// Should not send a notification because one was sent already for this channel
	closedChannels = []*lnrpc.ChannelCloseSummary{
		{
			ChanId:      closedChannelId,
			CloseHeight: blockHeight + 1,
		},
	}
	channelManager.checkClosedChannels()

	assert.Len(t, sentMessages, 0)
	assert.True(t, channelManager.closedChannels[closedChannels[0].ChanId])

	closedChannels = closedChannels[:0]
	cleanUp()
}

func checkForceCloseMessage(t *testing.T, message string, closedChannel *lnrpc.ChannelCloseSummary, closeType lnrpc.ChannelCloseSummary_ClosureType) {
	cleanUp()
	closedChannel.CloseType = closeType

	channelManager.logClosedChannel(closedChannel)

	assert.Equal(t, sentMessages[0], message)
	assert.True(t, strings.HasSuffix(loggedMessages[0], message+"\n"))
}

func TestLogClosedChannel(t *testing.T) {
	cleanUp()

	// Cooperatively closed channel
	closedChannel := &lnrpc.ChannelCloseSummary{
		ChanId:       321,
		RemotePubkey: "pubkey",
		CloseType:    lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE,
	}

	message := "Channel `321` to `pubkey` was closed"

	channelManager.logClosedChannel(closedChannel)

	assert.Equal(t, sentMessages[0], message)
	assert.True(t, strings.HasSuffix(loggedMessages[0], message+"\n"))

	// Force closed channels
	message = strings.ReplaceAll(message, "closed", "**force closed** :rotating_light:")

	checkForceCloseMessage(t, message, closedChannel, lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE)
	checkForceCloseMessage(t, message, closedChannel, lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE)
	checkForceCloseMessage(t, message, closedChannel, lnrpc.ChannelCloseSummary_BREACH_CLOSE)
	checkForceCloseMessage(t, message, closedChannel, lnrpc.ChannelCloseSummary_FUNDING_CANCELED)
	checkForceCloseMessage(t, message, closedChannel, lnrpc.ChannelCloseSummary_ABANDONED)

	cleanUp()
}
