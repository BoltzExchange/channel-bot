package notifications

import (
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNodeName(t *testing.T) {
	lnd := &MockLndClient{}
	clock := &utils.Clock{
		MockTime: time.Now(),
	}

	nc := initNodeCache(lnd, clock)

	assert.Len(t, nc.cache, 0)

	assert.Equal(t, "pubkey", nc.getNodeName("pubkey"))

	assert.Len(t, nc.cache, 1)
	assert.Equal(t, "pubkey", nc.cache["pubkey"].name)
	assert.Equal(t, clock.MockTime, nc.cache["pubkey"].fetchedAt)

	clock.MockTime = time.Now().Add(-(nodeCacheExpiration + 1))
	lnd.nodeAlias = "someName"

	assert.Equal(t, "pubkey", nc.cache["pubkey"].name)

	assert.Equal(t, lnd.nodeAlias, nc.getNodeName("otherPubkey"))

	lnd.nodeAlias = "someNewName"
	assert.Equal(t, lnd.nodeAlias, nc.getNodeName("otherPubkey"))
}
