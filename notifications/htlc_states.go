package notifications

import (
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

type htlcHandler interface {
	handleHtlc(channelId uint64, isIncoming bool, amtMsat uint64)
}

type htlcStates struct {
	sm           htlcHandler
	pendingHtlcs map[string]*routerrpc.ForwardEvent
}

func initHtlcStates(sm htlcHandler) *htlcStates {
	return &htlcStates{
		sm:           sm,
		pendingHtlcs: map[string]*routerrpc.ForwardEvent{},
	}
}

func (s *htlcStates) handleEvent(event *routerrpc.HtlcEvent) {
	s.handleHtlcSide(event, true, event.IncomingChannelId, event.IncomingHtlcId)
	s.handleHtlcSide(event, false, event.OutgoingChannelId, event.OutgoingHtlcId)
}

func (s *htlcStates) handleHtlcSide(
	event *routerrpc.HtlcEvent,
	isIncoming bool,
	channelId, htlcId uint64,
) {
	if channelId == 0 {
		return
	}

	if fe := event.GetForwardEvent(); fe != nil {
		s.pendingHtlcs[concatChanIdHtlcId(channelId, htlcId)] = fe
		return
	}

	if se := event.GetSettleEvent(); se != nil {
		id := concatChanIdHtlcId(channelId, htlcId)

		htlc := s.pendingHtlcs[id]
		if htlc == nil {
			return
		}

		delete(s.pendingHtlcs, id)

		var amount uint64

		if isIncoming {
			amount = htlc.Info.IncomingAmtMsat
		} else {
			amount = htlc.Info.OutgoingAmtMsat
		}

		s.sm.handleHtlc(channelId, isIncoming, amount)

		return
	}

	// If it is not a new HTLC or a known HTLC settling, delete the HTLC id (in case we have it)
	delete(s.pendingHtlcs, concatChanIdHtlcId(channelId, htlcId))
}

func concatChanIdHtlcId(channelId, htlcId uint64) string {
	return fmt.Sprintf("%d/%d", channelId, htlcId)
}
