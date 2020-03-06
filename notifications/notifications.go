package notifications

import (
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"strconv"
	"time"
)

type ChannelManager struct {
	Interval int

	Lnd     *lnd.LND
	Discord *discord.Discord

	imbalancedChannels  map[uint64]bool
	significantChannels map[uint64]SignificantChannel

	startupHeight uint32
	// Map of closed channels for which notifications were sent already
	closedChannels map[uint64]bool
}

type ratios struct {
	min float64
	max float64
}

type SignificantChannel struct {
	Alias     string
	ChannelID uint64
	MinRatio  string
	MaxRatio  string

	ratios ratios
}

func (manager *ChannelManager) Init(significantChannels []*SignificantChannel) {
	logger.Info("Starting notification bot")

	// Balance notifications related initializations
	manager.imbalancedChannels = make(map[uint64]bool)
	manager.significantChannels = make(map[uint64]SignificantChannel)

	for _, significant := range significantChannels {
		maxRatio, _ := strconv.ParseFloat(significant.MaxRatio, 64)
		minRatio, _ := strconv.ParseFloat(significant.MinRatio, 64)

		ratios := ratios{
			max: maxRatio,
			min: minRatio,
		}

		significant.ratios = ratios

		manager.significantChannels[significant.ChannelID] = *significant
	}

	// Closed channel notifications related initializations
	manager.closedChannels = make(map[uint64]bool)

	nodeInfo, err := manager.Lnd.GetInfo()

	if err != nil {
		logger.Fatal("Could not get node info: " + err.Error())
		return
	}

	manager.startupHeight = nodeInfo.BlockHeight

	manager.check()

	ticker := time.NewTicker(time.Duration(manager.Interval) * time.Second)

	for range ticker.C {
		manager.check()
	}
}

func (manager *ChannelManager) check() {
	manager.checkBalances()
	manager.checkClosedChannels()
}

func getNodeName(lnd *lnd.LND, remotePubkey string) string {
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

func formatChannelID(channelId uint64) string {
	return strconv.FormatUint(channelId, 10)
}
