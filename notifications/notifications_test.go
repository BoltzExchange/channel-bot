package notifications

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
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

func cleanUp() {
	sentMessages = sentMessages[:0]
	loggedMessages = loggedMessages[:0]
}

type MockLndClient struct {
	nodeAlias string
}

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

func (m MockLndClient) GetNodeInfo(string) (*lnrpc.NodeInfo, error) {
	return &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			Alias: m.nodeAlias,
		},
	}, nil
}

func (m MockLndClient) GetChannelInfo(uint64) (*lnrpc.ChannelEdge, error) {
	panic("")
}

func (m MockLndClient) ListInactiveChannels() (*lnrpc.ListChannelsResponse, error) {
	panic("")
}

func (m MockLndClient) ForceCloseChannel(string) (lnrpc.Lightning_CloseChannelClient, error) {
	panic("")
}

func (m MockLndClient) SubscribeInvoices(chan<- *lnrpc.Invoice, chan<- error) {
	panic("implement me")
}

func (m MockLndClient) SubscribeHtlcEvents(chan<- *routerrpc.HtlcEvent, chan<- error) {
	panic("implement me")
}

func (m MockLndClient) SubscribeChannelEvents(chan<- *lnrpc.ChannelEventUpdate, chan<- error) {
	panic("implement me")
}
