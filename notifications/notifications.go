package notifications

import (
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/BoltzExchange/channel-bot/notifications/providers"
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

type subscriptionChannels struct {
	htlcEvents  <-chan *routerrpc.HtlcEvent
	htlcErrChan <-chan error

	invoices        <-chan *lnrpc.Invoice
	invoicesErrChan <-chan error
}

type ChannelManager struct {
	lnd                  lnd.LightningClient
	notificationProvider providers.NotificationProvider

	logInsignificant bool

	nc *nodeCache
	sm *stateManager

	subs *subscriptionChannels
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

func (manager *ChannelManager) Init(
	significantChannels []*SignificantChannel,
	logInsignificant bool,
	lnd lnd.LightningClient,
	notificationProvider providers.NotificationProvider,
) {
	logger.Info("Starting notification bot")

	manager.lnd = lnd
	manager.logInsignificant = logInsignificant
	manager.notificationProvider = notificationProvider
	manager.nc = initNodeCache(manager.lnd, &utils.Clock{})
	manager.sm = initStateManager(manager, significantChannels)

	logger.Info("Subscribing to channel events")

	errChan := make(chan error)
	eventsChan := make(chan *lnrpc.ChannelEventUpdate)

	manager.lnd.SubscribeChannelEvents(eventsChan, errChan)

	manager.subscribeHtlcEvents()
	manager.prepareBalanceCheck()

	openedChannels := make(chan *lnrpc.Channel)
	closedChannels := make(chan *lnrpc.ChannelCloseSummary)

	go manager.handleHtlcEvents(openedChannels, closedChannels)

	for {
		select {
		case event := <-eventsChan:
			switch event.Type {
			case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
				openedChannels <- event.GetOpenChannel()

			case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
				closedChannels <- event.GetClosedChannel()
			}

			break

		case err := <-errChan:
			logger.Fatal("LND channel event subscription errored: " + err.Error())
			break
		}
	}
}
