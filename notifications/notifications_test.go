package notifications

import (
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"time"
)

type MockWriter struct{}

var loggedMessages []string

func (m *MockWriter) Write(p []byte) (n int, err error) {
	loggedMessages = append(loggedMessages, string(p))
	return 0, nil
}

type MockDiscordClient struct{}

var sentMessages []string

func (m MockDiscordClient) SendMessage(message string) error {
	sentMessages = append(sentMessages, message)
	return nil
}

type MockLndClient struct{}

const blockHeight uint32 = 534

func (m MockLndClient) GetInfo() (*lnrpc.GetInfoResponse, error) {
	return &lnrpc.GetInfoResponse{
		BlockHeight: blockHeight,
	}, nil
}

var channels []*lnrpc.Channel

func (m MockLndClient) ListChannels() (*lnrpc.ListChannelsResponse, error) {
	return &lnrpc.ListChannelsResponse{
		Channels: channels,
	}, nil
}

var closedChannels []*lnrpc.ChannelCloseSummary

func (m MockLndClient) ClosedChannels() (*lnrpc.ClosedChannelsResponse, error) {
	return &lnrpc.ClosedChannelsResponse{
		Channels: closedChannels,
	}, nil
}

func (m MockLndClient) GetNodeInfo(pubkey string) (*lnrpc.NodeInfo, error) {
	return &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			Alias: "",
		},
	}, nil
}

func (m MockLndClient) GetChannelInfo(chanId uint64) (*lnrpc.ChannelEdge, error) {
	panic("")
}

func (m MockLndClient) ListInactiveChannels() (*lnrpc.ListChannelsResponse, error) {
	panic("")
}

func (m MockLndClient) ForceCloseChannel(channelPoint string) (lnrpc.Lightning_CloseChannelClient, error) {
	panic("")
}

func initChannelManager() {
	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	lnd := &MockLndClient{}
	discord := &MockDiscordClient{}

	go func() {
		channelManager.Init(
			[]*SignificantChannel{},
			lnd,
			discord,
		)
	}()

	time.Sleep(time.Duration(10) * time.Millisecond)
	channelManager.ticker.Stop()
}

func cleanUp() {
	sentMessages = sentMessages[:0]
	loggedMessages = loggedMessages[:0]
}

var channelManager = ChannelManager{
	Interval: 10,
}

func TestInit(t *testing.T) {
	cleanUp()

	initChannelManager()

	assert.NotNil(t, channelManager.lnd)
	assert.NotNil(t, channelManager.discord)
	assert.NotNil(t, channelManager.closedChannels)
	assert.NotNil(t, channelManager.imbalancedChannels)
	assert.NotNil(t, channelManager.significantChannels)
	assert.NotNil(t, channelManager.ticker, "Did not start ticker")

	assert.Equal(t, channelManager.startupHeight, blockHeight)

	assert.Len(t, sentMessages, 0)

	assert.True(t, strings.HasSuffix(loggedMessages[0], "Starting notification bot\n"))
	assert.True(t, strings.HasSuffix(loggedMessages[1], "Checking significant channel balances\n"))
	assert.True(t, strings.HasSuffix(loggedMessages[2], "Checking normal channel balances\n"))
	assert.True(t, strings.HasSuffix(loggedMessages[3], "Checking closed channels\n"))

	cleanUp()
}

func checkParseSignificantChannel(t *testing.T, unparsedChannel *SignificantChannel) {
	parsedChannel := channelManager.significantChannels[unparsedChannel.ChannelID]

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

	channelManager.parseSignificantChannels(significantChannels)

	checkParseSignificantChannel(t, significantChannels[0])
	checkParseSignificantChannel(t, significantChannels[1])
}
