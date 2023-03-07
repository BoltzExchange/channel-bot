package notifications

import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/stretchr/testify/assert"
	"testing"
)

type handledHtlc struct {
	channelId  uint64
	isIncoming bool
	amtMsat    uint64
}

type mockHtlcHandler struct {
	handledHtlcs []*handledHtlc
}

func (m *mockHtlcHandler) handleHtlc(channelId uint64, isIncoming bool, amtMsat uint64) {
	m.handledHtlcs = append(m.handledHtlcs, &handledHtlc{
		channelId:  channelId,
		isIncoming: isIncoming,
		amtMsat:    amtMsat,
	})
}

var hc = &mockHtlcHandler{}
var hs = initHtlcStates(hc)

var event = &routerrpc.HtlcEvent{
	IncomingChannelId: 987,
	OutgoingChannelId: 123,
	IncomingHtlcId:    56,
	OutgoingHtlcId:    3,
	EventType:         routerrpc.HtlcEvent_SEND,
	Event: &routerrpc.HtlcEvent_ForwardEvent{
		ForwardEvent: &routerrpc.ForwardEvent{
			Info: &routerrpc.HtlcInfo{
				IncomingTimelock: 753125,
				OutgoingTimelock: 753113,
				IncomingAmtMsat:  65012,
				OutgoingAmtMsat:  65000,
			},
		},
	},
}

func TestConcatChanIdHtlcId(t *testing.T) {
	assert.Equal(t, "123/321", concatChanIdHtlcId(123, 321))
	assert.Equal(t, "243909086723/45890", concatChanIdHtlcId(243909086723, 45890))
}

func TestHandleEventForward(t *testing.T) {
	hs.handleEvent(event)

	assert.Len(t, hs.pendingHtlcs, 2)
	assert.Equal(
		t,
		event.GetForwardEvent(),
		hs.pendingHtlcs[concatChanIdHtlcId(event.IncomingChannelId, event.IncomingHtlcId)],
	)
	assert.Equal(
		t,
		event.GetForwardEvent(),
		hs.pendingHtlcs[concatChanIdHtlcId(event.OutgoingChannelId, event.OutgoingHtlcId)],
	)
}

func TestHandleEventSettle(t *testing.T) {
	settleEvent := &routerrpc.HtlcEvent{
		IncomingChannelId: event.IncomingChannelId,
		IncomingHtlcId:    event.IncomingHtlcId,
		EventType:         routerrpc.HtlcEvent_SEND,
		Event: &routerrpc.HtlcEvent_SettleEvent{
			SettleEvent: &routerrpc.SettleEvent{},
		},
	}

	hs.handleEvent(settleEvent)

	assert.Len(t, hs.pendingHtlcs, 1)
	assert.Equal(
		t,
		event.GetForwardEvent(),
		hs.pendingHtlcs[concatChanIdHtlcId(event.OutgoingChannelId, event.OutgoingHtlcId)],
	)

	assert.Len(t, hc.handledHtlcs, 1)
	assert.Equal(t, hc.handledHtlcs[0], &handledHtlc{
		channelId:  settleEvent.IncomingChannelId,
		isIncoming: true,
		amtMsat:    event.GetForwardEvent().Info.IncomingAmtMsat,
	})

	settleEvent = &routerrpc.HtlcEvent{
		OutgoingChannelId: event.OutgoingChannelId,
		OutgoingHtlcId:    event.OutgoingHtlcId,
		EventType:         routerrpc.HtlcEvent_SEND,
		Event: &routerrpc.HtlcEvent_SettleEvent{
			SettleEvent: &routerrpc.SettleEvent{},
		},
	}

	hs.handleEvent(settleEvent)

	assert.Len(t, hs.pendingHtlcs, 0)

	assert.Len(t, hc.handledHtlcs, 2)
	assert.Equal(t, hc.handledHtlcs[1], &handledHtlc{
		channelId:  settleEvent.OutgoingChannelId,
		isIncoming: false,
		amtMsat:    event.GetForwardEvent().Info.OutgoingAmtMsat,
	})
}

func TestHandleEventFailure(t *testing.T) {
	hs.handleEvent(event)

	failEvent := &routerrpc.HtlcEvent{
		IncomingChannelId: 987,
		OutgoingChannelId: 123,
		IncomingHtlcId:    56,
		OutgoingHtlcId:    3,
	}

	assert.Len(t, hs.pendingHtlcs, 2)

	hs.handleEvent(failEvent)
	assert.Len(t, hs.pendingHtlcs, 0)
}

func TestIgnoreUnknownSettle(t *testing.T) {
	htlcsHandled := hc.handledHtlcs

	settleEvent := &routerrpc.HtlcEvent{
		OutgoingChannelId: event.OutgoingChannelId,
		OutgoingHtlcId:    event.OutgoingHtlcId,
		EventType:         routerrpc.HtlcEvent_SEND,
		Event: &routerrpc.HtlcEvent_SettleEvent{
			SettleEvent: &routerrpc.SettleEvent{},
		},
	}

	hs.handleEvent(settleEvent)
	assert.Equal(t, htlcsHandled, hc.handledHtlcs)
}

func TestHandleEventSideUnpopulated(t *testing.T) {
	hc := &mockHtlcHandler{}
	hs := initHtlcStates(hc)

	event := &routerrpc.HtlcEvent{
		OutgoingChannelId: 123,
		OutgoingHtlcId:    3,
		EventType:         routerrpc.HtlcEvent_SEND,
		Event: &routerrpc.HtlcEvent_ForwardEvent{
			ForwardEvent: &routerrpc.ForwardEvent{
				Info: &routerrpc.HtlcInfo{
					OutgoingAmtMsat:  65000,
					OutgoingTimelock: 753113,
				},
			},
		},
	}

	hs.handleEvent(event)

	assert.Len(t, hs.pendingHtlcs, 1)
	assert.Equal(
		t,
		event.GetForwardEvent(),
		hs.pendingHtlcs[concatChanIdHtlcId(event.OutgoingChannelId, event.OutgoingHtlcId)],
	)
}
