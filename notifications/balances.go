package notifications

import (
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

func (manager *ChannelManager) prepareBalanceCheck() {
	channels, err := manager.lnd.ListChannels()

	if err != nil {
		logger.Error("Could not get channels: " + err.Error())
		return
	}

	manager.sm.populateChannels(channels)
}

func (manager *ChannelManager) subscribeHtlcEvents() {
	logger.Info("Subscribing to HTLC events")

	htlcEvents := make(chan *routerrpc.HtlcEvent)
	htlcsErrChan := make(chan error)

	manager.lnd.SubscribeHtlcEvents(htlcEvents, htlcsErrChan)

	logger.Info("Subscribing to invoices")

	invoices := make(chan *lnrpc.Invoice)
	invoicesErrChan := make(chan error)

	manager.lnd.SubscribeInvoices(invoices, invoicesErrChan)

	manager.subs = &subscriptionChannels{
		htlcEvents:      htlcEvents,
		htlcErrChan:     htlcsErrChan,
		invoices:        invoices,
		invoicesErrChan: invoicesErrChan,
	}
}

func (manager *ChannelManager) handleHtlcEvents(
	openedChannels <-chan *lnrpc.Channel,
	closedChannels <-chan *lnrpc.ChannelCloseSummary,
) {
	hc := initHtlcStates(manager.sm)

	for {
		select {
		case opened := <-openedChannels:
			manager.sm.handleOpen(opened)
			break

		case closed := <-closedChannels:
			manager.sm.handleClose(closed)
			break

		case event := <-manager.subs.htlcEvents:
			if event.EventType == routerrpc.HtlcEvent_SEND || event.EventType == routerrpc.HtlcEvent_FORWARD {
				hc.handleEvent(event)
			}
			break

		case invoice := <-manager.subs.invoices:
			if invoice.State != lnrpc.Invoice_SETTLED {
				break
			}

			manager.sm.handleSettledInvoice(invoice)
			break

		case err := <-manager.subs.htlcErrChan:
			logger.Fatal("LND channel event subscription errored: " + err.Error())
			break

		case err := <-manager.subs.invoicesErrChan:
			logger.Fatal("LND invoice subscription errored: " + err.Error())
			break
		}
	}
}
