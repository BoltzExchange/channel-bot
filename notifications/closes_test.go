package notifications

import (
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func checkForceCloseMessage(
	t *testing.T,
	cm *ChannelManager,
	message string,
	closedChannel *lnrpc.ChannelCloseSummary,
	closeType lnrpc.ChannelCloseSummary_ClosureType,
) {
	cleanUp()
	closedChannel.CloseType = closeType

	cm.logClosedChannel(closedChannel)

	assert.Equal(t, sentMessages[0], message)
	assert.True(t, strings.HasSuffix(loggedMessages[0], message+"\n"))
}

func TestLogClosedChannel(t *testing.T) {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	channelManager := &ChannelManager{
		lnd:     &MockLndClient{},
		discord: &MockDiscordClient{},
		nc:      initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}
	channelManager.sm = initStateManager(channelManager, nil)

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

	checkForceCloseMessage(t, channelManager, message, closedChannel, lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE)
	checkForceCloseMessage(t, channelManager, message, closedChannel, lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE)
	checkForceCloseMessage(t, channelManager, message, closedChannel, lnrpc.ChannelCloseSummary_BREACH_CLOSE)
	checkForceCloseMessage(t, channelManager, message, closedChannel, lnrpc.ChannelCloseSummary_FUNDING_CANCELED)
	checkForceCloseMessage(t, channelManager, message, closedChannel, lnrpc.ChannelCloseSummary_ABANDONED)

	cleanUp()
}
