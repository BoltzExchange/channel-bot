package cleaner

import (
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
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

func (m *MockDiscordClient) SendMessage(message string) error {
	sentMessages = append(sentMessages, message)
	return nil
}

type MockLndClient struct{}

func (m *MockLndClient) GetInfo() (*lnrpc.GetInfoResponse, error) {
	panic("")
}

func (m *MockLndClient) ListChannels() (*lnrpc.ListChannelsResponse, error) {
	panic("")
}

func (m *MockLndClient) ClosedChannels() (*lnrpc.ClosedChannelsResponse, error) {
	panic("")
}

const nodeAlias = "alias"

func (m *MockLndClient) GetNodeInfo(pubkey string) (*lnrpc.NodeInfo, error) {
	return &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			Alias: nodeAlias,
		},
	}, nil
}

var channelInfo = &lnrpc.ChannelEdge{
	Node1Policy: &lnrpc.RoutingPolicy{
		LastUpdate: uint32(time.Now().Unix()),
	},
	Node2Policy: &lnrpc.RoutingPolicy{
		LastUpdate: uint32(time.Now().Unix()),
	},
}

func (m *MockLndClient) GetChannelInfo(chanId uint64) (*lnrpc.ChannelEdge, error) {
	return channelInfo, nil
}

var forceClosedChannels []string

func (m *MockLndClient) ForceCloseChannel(channelPoint string) (lnrpc.Lightning_CloseChannelClient, error) {
	forceClosedChannels = append(forceClosedChannels, channelPoint)
	return nil, nil
}

var inactiveChannelsResponse = &lnrpc.ListChannelsResponse{}

func (m *MockLndClient) ListInactiveChannels() (*lnrpc.ListChannelsResponse, error) {
	return inactiveChannelsResponse, nil
}

func cleanUp() {
	sentMessages = sentMessages[:0]
	loggedMessages = loggedMessages[:0]
	forceClosedChannels = forceClosedChannels[:0]
}

var cleaner = ChannelCleaner{
	Interval:               1,
	MaxInactiveTime:        30,
	MaxInactiveTimePrivate: 60,
}

func TestInit(t *testing.T) {
	cleanUp()

	mockWriter := &MockWriter{}
	logger.Init("", false, false, mockWriter)

	lnd := &MockLndClient{}
	discord := &MockDiscordClient{}

	go func() {
		cleaner.Init(lnd, discord)
	}()

	time.Sleep(time.Duration(10) * time.Millisecond)

	assert.NotNil(t, cleaner.ticker, "Does not start ticker for periodical channel cleanup")

	assert.True(t, strings.HasSuffix(loggedMessages[0], "Starting channel cleaner\n"), "Does not send log message on startup")
	assert.True(t, strings.HasSuffix(loggedMessages[1], "Cleaning inactive channels\n"), "Does not execute cleaning routine on startup")

	cleaner.ticker.Stop()
	cleanUp()
}

func testForceClose(t *testing.T) {
	// Should force close channel
	cleaner.forceCloseChannels()

	assert.Equal(t, inactiveChannelsResponse.Channels[0].ChannelPoint, forceClosedChannels[0], "Did not force close channel that has not been updated for longer than the max inactive time")
	assert.True(t, len(sentMessages) == 1 && len(loggedMessages) == 2, "Did not log channel closure")

	// Should not force close because the last update of node 2 is not old enough
	channelInfo.Node2Policy.LastUpdate = uint32(time.Now().Unix())

	cleaner.forceCloseChannels()

	assert.True(t, len(forceClosedChannels) == 1 && len(sentMessages) == 1, "Did force close channel although the node 2 update is not old enough")

	cleanUp()
}

func TestForceCloseChannels(t *testing.T) {
	cleanUp()

	tooOldPublic := uint32(time.Now().AddDate(0, 0, -(cleaner.MaxInactiveTime + 1)).Unix())
	tooOldPrivate := uint32(time.Now().AddDate(0, 0, -(cleaner.MaxInactiveTimePrivate + 1)).Unix())

	// Public channel
	inactiveChannelsResponse.Channels = []*lnrpc.Channel{
		{
			ChanId:       1,
			Private:      false,
			RemotePubkey: "pub1",
			ChannelPoint: "public:1",
		},
	}

	channelInfo.Node1Policy.LastUpdate = tooOldPublic
	channelInfo.Node2Policy.LastUpdate = tooOldPublic

	testForceClose(t)

	// Private channel
	inactiveChannelsResponse.Channels = []*lnrpc.Channel{
		{
			ChanId:       2,
			Private:      true,
			RemotePubkey: "pub2",
			ChannelPoint: "private:2",
		},
	}

	channelInfo.Node1Policy.LastUpdate = tooOldPrivate
	channelInfo.Node2Policy.LastUpdate = tooOldPrivate

	testForceClose(t)

	// In this example the max inactive time for private channels is longer,
	// hence private channels should not be closed if they the have been inactive
	// for longer than the max inactive time of public channel but less long
	// than the on of a private channel
	channelInfo.Node1Policy.LastUpdate = tooOldPublic
	channelInfo.Node2Policy.LastUpdate = tooOldPublic

	cleaner.forceCloseChannels()

	assert.Len(t, forceClosedChannels, 0, "Did force private because max timeout of public channels was used")

	cleanUp()

	// Sanity check to make sure the loop works as it should.
	// The first channel should be skipped because it is a
	// private one (those have longer timeouts in the test setup)
	// and the second channel should be closed because it
	// is a public channel
	inactiveChannelsResponse.Channels = []*lnrpc.Channel{
		{
			ChanId:       3,
			Private:      true,
			RemotePubkey: "pub3",
			ChannelPoint: "private:3",
		},
		{
			ChanId:       4,
			Private:      false,
			RemotePubkey: "pub4",
			ChannelPoint: "public:4",
		},
	}

	cleaner.forceCloseChannels()

	assert.Equal(t, inactiveChannelsResponse.Channels[1].ChannelPoint, forceClosedChannels[0], "Loop was cancelled after first inactive channel that was not force closed")

	cleanUp()
}

func TestLogClosingChannels(t *testing.T) {
	cleanUp()

	channel := &lnrpc.Channel{
		Private:      false,
		ChanId:       145135534931969,
		RemotePubkey: "03793e5deff6c3acc0558440bf04ffd6ea2adebd8eb50246b98a8d27abbf79539a",
	}

	daysAgo := 90
	lastUpdate := time.Now().AddDate(0, 0, -daysAgo)

	cleaner.logClosingChannels(channel, lastUpdate)

	expectedMessage := "Force closing public channel `145135534931969` to `alias` because it was inactive for 90 days"

	assert.Equal(t, expectedMessage, sentMessages[0], "Message sent before closing is invalid: "+sentMessages[0])
	assert.True(t, strings.HasSuffix(loggedMessages[0], sentMessages[0]+"\n"))

	channel.Private = true
	expectedMessage = strings.Replace(expectedMessage, "public", "private", 1)

	cleaner.logClosingChannels(channel, lastUpdate)

	assert.Equal(t, expectedMessage, sentMessages[1], "Message sent before closing is invalid: "+sentMessages[0])
	assert.True(t, strings.HasSuffix(loggedMessages[1], sentMessages[1]+"\n"))

	cleanUp()
}
