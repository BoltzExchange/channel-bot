package lnd

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type LightningClient interface {
	GetInfo() (*lnrpc.GetInfoResponse, error)
	GetNodeInfo(pubkey string) (*lnrpc.NodeInfo, error)
	ListChannels() (*lnrpc.ListChannelsResponse, error)
	ClosedChannels() (*lnrpc.ClosedChannelsResponse, error)
	GetChannelInfo(chanId uint64) (*lnrpc.ChannelEdge, error)
	ListInactiveChannels() (*lnrpc.ListChannelsResponse, error)
	ForceCloseChannel(channelPoint string) (lnrpc.Lightning_CloseChannelClient, error)

	SubscribeInvoices(events chan<- *lnrpc.Invoice, errChan chan<- error)
	SubscribeHtlcEvents(events chan<- *routerrpc.HtlcEvent, errChan chan<- error)
	SubscribeChannelEvents(events chan<- *lnrpc.ChannelEventUpdate, errChan chan<- error)
}

type LND struct {
	Host            string `long:"lnd.host" description:"gRPC host of the LND node"`
	Port            int    `long:"lnd.port" description:"gRPC port of the LND node"`
	Macaroon        string `long:"lnd.macaroon" description:"Path to a macaroon file of the LND node"`
	Certificate     string `long:"lnd.certificate" description:"Path to a certificate file of the LND node"`
	TlsNameOverride string `long:"lnd.tls-name-override" description:"Override the TLS name of the LND node"`

	ctx    context.Context
	client lnrpc.LightningClient
	router routerrpc.RouterClient
}

func (lnd *LND) Connect() error {
	creds, err := credentials.NewClientTLSFromFile(lnd.Certificate, lnd.TlsNameOverride)

	if err != nil {
		return errors.New(fmt.Sprint("could not read LND certificate: ", err))
	}

	con, err := grpc.Dial(lnd.Host+":"+strconv.Itoa(lnd.Port), grpc.WithTransportCredentials(creds))

	if err != nil {
		return errors.New(fmt.Sprint("could not create gRPC client: ", err))
	}

	lnd.client = lnrpc.NewLightningClient(con)
	lnd.router = routerrpc.NewRouterClient(con)

	if lnd.ctx == nil {
		macaroonFile, err := os.ReadFile(lnd.Macaroon)

		if err != nil {
			return errors.New(fmt.Sprint("could not read LND macaroon: ", err))
		}

		macaroon := metadata.Pairs("macaroon", hex.EncodeToString(macaroonFile))
		lnd.ctx = metadata.NewOutgoingContext(context.Background(), macaroon)
	}

	return nil
}

func (lnd *LND) GetInfo() (*lnrpc.GetInfoResponse, error) {
	return lnd.client.GetInfo(lnd.ctx, &lnrpc.GetInfoRequest{})
}

func (lnd *LND) ListChannels() (*lnrpc.ListChannelsResponse, error) {
	return lnd.client.ListChannels(lnd.ctx, &lnrpc.ListChannelsRequest{})
}

func (lnd *LND) ListInactiveChannels() (*lnrpc.ListChannelsResponse, error) {
	return lnd.client.ListChannels(lnd.ctx, &lnrpc.ListChannelsRequest{
		InactiveOnly: true,
	})
}

func (lnd *LND) ClosedChannels() (*lnrpc.ClosedChannelsResponse, error) {
	return lnd.client.ClosedChannels(lnd.ctx, &lnrpc.ClosedChannelsRequest{})
}

func (lnd *LND) GetNodeInfo(pubkey string) (*lnrpc.NodeInfo, error) {
	return lnd.client.GetNodeInfo(lnd.ctx, &lnrpc.NodeInfoRequest{
		PubKey: pubkey,
	})
}

func (lnd *LND) GetChannelInfo(chanId uint64) (*lnrpc.ChannelEdge, error) {
	return lnd.client.GetChanInfo(lnd.ctx, &lnrpc.ChanInfoRequest{
		ChanId: chanId,
	})
}

func (lnd *LND) ForceCloseChannel(channelPoint string) (lnrpc.Lightning_CloseChannelClient, error) {
	channel := parseChannelPoint(channelPoint)

	return lnd.client.CloseChannel(lnd.ctx, &lnrpc.CloseChannelRequest{
		ChannelPoint: &channel,
		Force:        true,
	})
}

func (lnd *LND) SubscribeChannelEvents(events chan<- *lnrpc.ChannelEventUpdate, errChan chan<- error) {
	client, err := lnd.client.SubscribeChannelEvents(lnd.ctx, &lnrpc.ChannelEventSubscription{})
	if err != nil {
		errChan <- err
		return
	}

	go handleSubscription(client.Recv, events, errChan)
}

func (lnd *LND) SubscribeInvoices(events chan<- *lnrpc.Invoice, errChan chan<- error) {
	client, err := lnd.client.SubscribeInvoices(lnd.ctx, &lnrpc.InvoiceSubscription{})
	if err != nil {
		errChan <- err
		return
	}

	go handleSubscription(client.Recv, events, errChan)
}

func (lnd *LND) SubscribeHtlcEvents(events chan<- *routerrpc.HtlcEvent, errChan chan<- error) {
	client, err := lnd.router.SubscribeHtlcEvents(lnd.ctx, &routerrpc.SubscribeHtlcEventsRequest{})
	if err != nil {
		errChan <- err
		return
	}

	go handleSubscription(client.Recv, events, errChan)
}

func handleSubscription[T any](recv func() (T, error), events chan<- T, errChan chan<- error) {
	for {
		event, err := recv()
		if err != nil {
			errChan <- err
			break
		}

		events <- event
	}
}
