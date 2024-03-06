package notifications

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"strconv"
)

type stateManager struct {
	manager *ChannelManager

	significantChannels map[uint64]*SignificantChannel
	imbalancedChannels  map[uint64]bool

	channels map[uint64]*lnrpc.Channel
}

var defaultRatios = ratios{
	max: 0.7,
	min: 0.3,
}

func initStateManager(manager *ChannelManager, significantChannels []*SignificantChannel) *stateManager {
	sm := &stateManager{
		manager: manager,

		significantChannels: map[uint64]*SignificantChannel{},
		imbalancedChannels:  map[uint64]bool{},

		channels: map[uint64]*lnrpc.Channel{},
	}

	for _, significant := range significantChannels {
		maxRatio, _ := strconv.ParseFloat(significant.MaxRatio, 64)
		minRatio, _ := strconv.ParseFloat(significant.MinRatio, 64)

		significant.ratios = ratios{
			max: maxRatio,
			min: minRatio,
		}

		sm.significantChannels[significant.ChannelID] = significant
	}

	return sm
}

func (sm *stateManager) populateChannels(channels *lnrpc.ListChannelsResponse) {
	for _, channel := range channels.Channels {
		sm.channels[channel.ChanId] = channel
	}

	for _, signi := range sm.significantChannels {
		if channel := sm.channels[signi.ChannelID]; channel != nil {
			sm.checkChannel(channel)
		} else {
			signi.logSignificantNotFound(sm.manager.notificationProvider)
		}
	}
}

func (sm *stateManager) handleOpen(channel *lnrpc.Channel) {
	sm.channels[channel.ChanId] = channel
	sm.checkChannel(channel)
}

func (sm *stateManager) handleClose(closed *lnrpc.ChannelCloseSummary) {
	delete(sm.channels, closed.ChanId)
	delete(sm.imbalancedChannels, closed.ChanId)

	sm.manager.logClosedChannel(closed)

	if signi := sm.significantChannels[closed.ChanId]; signi != nil {
		signi.logSignificantNotFound(sm.manager.notificationProvider)
	}
}

func (sm *stateManager) handleHtlc(channelId uint64, isIncoming bool, amtMsat uint64) {
	channel := sm.channels[channelId]
	if channel == nil {
		return
	}

	updateChannelBalances(channel, isIncoming, amtMsat)
	sm.checkChannel(channel)
}

func (sm *stateManager) handleSettledInvoice(invoice *lnrpc.Invoice) {
	touchedChannels := map[uint64]*lnrpc.Channel{}

	for _, htlc := range invoice.Htlcs {
		channel := sm.channels[htlc.ChanId]
		touchedChannels[htlc.ChanId] = channel

		updateChannelBalances(channel, true, htlc.AmtMsat)
	}

	for _, channel := range touchedChannels {
		sm.checkChannel(channel)
	}
}

func (sm *stateManager) checkChannel(channel *lnrpc.Channel) {
	signi, isSignificant := sm.significantChannels[channel.ChanId]

	if !isSignificant && (channel.UnsettledBalance != 0 || channel.Private) {
		return
	}

	channelRatio := getChannelRatio(channel)

	var checkRatio ratios

	if isSignificant {
		checkRatio = signi.ratios
	} else {
		checkRatio = defaultRatios
	}

	if channelRatio > checkRatio.min && channelRatio < checkRatio.max {
		if contains := sm.imbalancedChannels[channel.ChanId]; contains {
			if isSignificant {
				signi.logBalance(sm.manager.notificationProvider, channel, false)
			} else {
				sm.manager.logBalance(channel, false)
			}
			delete(sm.imbalancedChannels, channel.ChanId)
		}

		return
	}

	if contains := sm.imbalancedChannels[channel.ChanId]; contains {
		return
	}

	sm.imbalancedChannels[channel.ChanId] = true

	if isSignificant {
		signi.logBalance(sm.manager.notificationProvider, channel, true)
	} else {
		sm.manager.logBalance(channel, true)
	}
}

func updateChannelBalances(channel *lnrpc.Channel, isIncoming bool, amountMsat uint64) {
	if !isIncoming {
		amountMsat = -amountMsat
	}

	amountSat := int64(amountMsat) / 1000

	channel.LocalBalance += amountSat
	channel.RemoteBalance -= amountSat
}

func getChannelRatio(channel *lnrpc.Channel) float64 {
	return float64(channel.LocalBalance) / float64(channel.Capacity)
}
