package notifications

import (
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

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

	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	cm := &ChannelManager{
		lnd:     &MockLndClient{},
		discord: &MockDiscordClient{},
		nc:      initNodeCache(&MockLndClient{}, &utils.Clock{}),
	}

	// Imbalanced
	message := "Channel `123` to `pubkey` is **imbalanced**:\n  Local: 32120398448\n  Remote: 123321123321"
	cm.logBalance(channel, true)

	checkLogs(t, message)

	cleanUp()

	// Balanced
	message = strings.Replace(message, "imbalanced", "balanced again", 1)
	cm.logBalance(channel, false)

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
