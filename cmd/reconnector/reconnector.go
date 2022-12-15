package main

import (
	"github.com/BoltzExchange/channel-bot/config"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/google/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
	"sync"
	"time"
)

func main() {
	cfg := config.LoadConfig(&reconnectorConfig{
		BaseConfig: config.BaseConfig{
			LogFile:    "./reconnector.log",
			ConfigFile: "./reconnector.toml",
		},

		Interval: 10,
	})
	utils.InitLogger(cfg.BaseConfig.LogFile)
	config.LogConfig(cfg)

	_ = utils.InitLnd(cfg.Lnd)

	dur := time.Duration(cfg.Interval) * time.Minute
	tick := time.NewTicker(dur)

	logger.Info("Reconnecting peers every: " + dur.String())

	reconnect(cfg.Lnd)

	for {
		<-tick.C
		reconnect(cfg.Lnd)
	}
}

func logReconnectError(err error) {
	logger.Error("Could not reconnect peers: " + err.Error())
}

func reconnect(lnd *lnd.LND) {
	peers, err := lnd.ListPeers()
	if err != nil {
		logReconnectError(err)
		return
	}

	inactives, err := lnd.ListInactiveChannels()
	if err != nil {
		logReconnectError(err)
		return
	}

	peersMap := map[string]bool{}
	for _, peer := range peers.Peers {
		peersMap[peer.PubKey] = true
	}

	var peersToReconnect []string

	for _, cha := range inactives.Channels {
		if hasPeer := peersMap[cha.RemotePubkey]; !hasPeer {
			continue
		}

		peersToReconnect = append(peersToReconnect, cha.RemotePubkey)
	}

	if len(peersToReconnect) == 0 {
		return
	}

	logger.Infof("Found %d peers to reconnect", len(peersToReconnect))

	var wg sync.WaitGroup
	wg.Add(len(peersToReconnect))

	for _, peer := range peersToReconnect {
		go func(peer string) {
			defer wg.Done()

			errNodeInfo := reconnectPeer(lnd, peer)
			if errNodeInfo != nil {
				err = errNodeInfo
				return
			}
		}(peer)
	}

	wg.Wait()

	if err != nil {
		logger.Error("Could not reconnect peers: " + err.Error())
	}
}

func reconnectPeer(lnd *lnd.LND, peer string) error {
	peerInfo, err := lnd.GetNodeInfo(peer)
	if err != nil {
		return err
	}

	err = lnd.DisconnectPeer(peer)
	if err != nil {
		return err
	}

	for _, addr := range peerInfo.Node.Addresses {
		err = lnd.ConnectPeer(&lnrpc.LightningAddress{
			Pubkey: peer,
			Host:   addr.Addr,
		})

		// When the connection was established successfully, either other URIs don't have to be tried anymore
		if err != nil {
			return nil
		}
	}

	return nil
}
