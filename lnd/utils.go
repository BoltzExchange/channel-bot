package lnd

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"strconv"
	"strings"
)

func GetNodeName(lnd LightningClient, remotePubkey string) string {
	nodeName := remotePubkey

	nodeInfo, err := lnd.GetNodeInfo(remotePubkey)

	// Use the alias if it can be queried and is not empty
	if err == nil {
		if nodeInfo.Node.Alias != "" {
			nodeName = nodeInfo.Node.Alias
		}
	}

	return nodeName
}

func FormatChannelID(channelId uint64) string {
	return strconv.FormatUint(channelId, 10)
}

func parseChannelPoint(channelPoint string) lnrpc.ChannelPoint {
	split := strings.Split(channelPoint, ":")
	outputIndex, _ := strconv.Atoi(split[1])

	return lnrpc.ChannelPoint{
		FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
			FundingTxidStr: split[0],
		},
		OutputIndex: uint32(outputIndex),
	}
}
