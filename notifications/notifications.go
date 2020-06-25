package notifications

import (
	"github.com/BoltzExchange/channel-bot/discord"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/google/logger"
	"strconv"
	"time"
)

type ChannelManager struct {
	Interval int `short:"i" long:"notifications.interval" description:"Interval in seconds at which the channel balances and closed channels should be checked. Set to 0 to disable this feature"`

	lnd     lnd.LightningClient
	discord discord.NotificationService

	imbalancedChannels  map[uint64]bool
	significantChannels map[uint64]SignificantChannel

	startupHeight uint32
	// Map of closed channels for which notifications were sent already
	closedChannels map[uint64]bool

	// Map of significant channels that couldn't be found
	notFoundSignificantChannel map[uint64]bool

	ticker *time.Ticker
}

type ratios struct {
	min float64
	max float64
}

type SignificantChannel struct {
	Alias     string
	ChannelID uint64

	// These values are just for parsing
	MinRatio string
	MaxRatio string

	// This struct is actually used for comparisons
	ratios ratios
}

func (manager *ChannelManager) Init(significantChannels []*SignificantChannel, lnd lnd.LightningClient, discord discord.NotificationService) {
	if manager.Interval == 0 {
		return
	}

	logger.Info("Starting notification bot")

	manager.lnd = lnd
	manager.discord = discord

	// Balance notification related initializations
	manager.imbalancedChannels = make(map[uint64]bool)
	manager.notFoundSignificantChannel = make(map[uint64]bool)

	manager.parseSignificantChannels(significantChannels)

	// Closed channel notifications related initializations
	manager.closedChannels = make(map[uint64]bool)

	nodeInfo, err := manager.lnd.GetInfo()

	if err != nil {
		logger.Fatal("Could not get node info: " + err.Error())
		return
	}

	manager.startupHeight = nodeInfo.BlockHeight

	manager.check(true)

	manager.ticker = time.NewTicker(time.Duration(manager.Interval) * time.Second)

	for range manager.ticker.C {
		manager.check(false)
	}
}

func (manager *ChannelManager) check(isStartup bool) {
	manager.checkBalances(isStartup)
	manager.checkClosedChannels()
}

func (manager *ChannelManager) parseSignificantChannels(significantChannels []*SignificantChannel) {
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
}
