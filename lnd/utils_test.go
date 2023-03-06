package lnd

import (
	"errors"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"strconv"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

const nodeAlias = "alias"

const failPublicKey = "fail"
const emptyPublicKey = "empty"

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

func (m *MockLndClient) GetNodeInfo(pubkey string) (*lnrpc.NodeInfo, error) {
	nodeInfo := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			Alias: nodeAlias,

			LastUpdate: 0,
			PubKey:     "",
			Color:      "",
			Addresses:  nil,
			Features:   nil,
		},
		Channels:      nil,
		NumChannels:   0,
		TotalCapacity: 0,
	}

	if pubkey == emptyPublicKey {
		nodeInfo.Node.Alias = ""
	}

	var err error = nil

	if pubkey == failPublicKey {
		err = errors.New("")
	}

	return nodeInfo, err
}

func (m *MockLndClient) GetChannelInfo(chanId uint64) (*lnrpc.ChannelEdge, error) {
	panic("")
}

func (m *MockLndClient) ForceCloseChannel(channelPoint string) (lnrpc.Lightning_CloseChannelClient, error) {
	panic("")
}

func (m *MockLndClient) ListInactiveChannels() (*lnrpc.ListChannelsResponse, error) {
	panic("")
}

func (m *MockLndClient) SubscribeInvoices(chan<- *lnrpc.Invoice, chan<- error) {
	panic("implement me")
}

func (m *MockLndClient) SubscribeHtlcEvents(chan<- *routerrpc.HtlcEvent, chan<- error) {
	panic("implement me")
}

func (m *MockLndClient) SubscribeChannelEvents(chan<- *lnrpc.ChannelEventUpdate, chan<- error) {
	panic("implement me")
}

func TestGetNodeName(t *testing.T) {
	client := &MockLndClient{}

	nodeName := GetNodeName(client, "")
	assert.Equal(t, nodeAlias, nodeName, "Node name is not queried alias")

	nodeName = GetNodeName(client, failPublicKey)
	assert.Equal(t, failPublicKey, nodeName, "Node name is not remote public key if alias cannot be queried")

	nodeName = GetNodeName(client, emptyPublicKey)
	assert.Equal(t, emptyPublicKey, nodeName, "Node name is not remote public key if alias is an empty string")
}

func TestFormatChannelID(t *testing.T) {
	channelIds := []uint64{
		158329674465281,
		118747255865345,
		145135534931969,
	}

	expectedResults := []string{
		"158329674465281",
		"118747255865345",
		"145135534931969",
	}

	for i, id := range channelIds {
		assert.Equal(
			t,
			expectedResults[i],
			FormatChannelID(id),
			"Channel ID "+strconv.FormatInt(int64(id), 10)+" is not formatted correctly",
		)
	}
}

func TestParseChannelPoint(t *testing.T) {
	channelPoints := []string{
		"446b399af44adbcddf85205d52438356437469fa7d44b0402a03e403318dc0a3:0",
		"787ac9eef35c6c8e96a64d823fe992481078306d85a69b1c279e08dd0a29ca68:1",
	}

	expectedResults := []lnrpc.ChannelPoint{
		{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
				FundingTxidStr: "446b399af44adbcddf85205d52438356437469fa7d44b0402a03e403318dc0a3",
			},
			OutputIndex: 0,
		},
		{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
				FundingTxidStr: "787ac9eef35c6c8e96a64d823fe992481078306d85a69b1c279e08dd0a29ca68",
			},
			OutputIndex: 1,
		},
	}

	for i, channelPoint := range channelPoints {
		assert.Equal(t, expectedResults[i].FundingTxid, parseChannelPoint(channelPoint).FundingTxid)
		assert.Equal(t, expectedResults[i].OutputIndex, parseChannelPoint(channelPoint).OutputIndex)
	}
}
