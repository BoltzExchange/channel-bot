package notifications

import (
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/BoltzExchange/channel-bot/utils"
	"time"
)

const nodeCacheExpiration = time.Hour * 24

type nodeInfo struct {
	name      string
	fetchedAt time.Time
}

type nodeCache struct {
	clock *utils.Clock
	lnd   lnd.LightningClient

	cache map[string]*nodeInfo
}

func initNodeCache(lnd lnd.LightningClient, clock *utils.Clock) *nodeCache {
	return &nodeCache{
		clock: clock,
		lnd:   lnd,
		cache: map[string]*nodeInfo{},
	}
}

func (nc *nodeCache) getNodeName(pubkey string) string {
	c := nc.cache[pubkey]

	if c == nil || time.Since(c.fetchedAt) > nodeCacheExpiration {
		c = &nodeInfo{
			name:      lnd.GetNodeName(nc.lnd, pubkey),
			fetchedAt: nc.clock.Now(),
		}
		nc.cache[pubkey] = c
	}

	return c.name
}
