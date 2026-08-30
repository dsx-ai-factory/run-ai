package runaipubsub

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/redis"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	conf := &config_models.Pubsub{
		Type: config_models.PSTRedis,
	}
	config.Set(conf)
	factory := newPubSubFactory()
	logger := logger.NewLogger(&dummyLogger{}, func(e error) {})
	pubsuber, err := factory.Create(logger, conf.Type)
	assert.Nil(t, err)
	assert.Equal(t, redis.NewRedisPubSuber(logger), pubsuber)
}

func TestFactoryUnknownType(t *testing.T) {
	var kafkaPST config_models.PubSubType = "unknown"
	conf := &config_models.Pubsub{
		Type: kafkaPST,
	}
	config.Set(conf)
	factory := newPubSubFactory()
	pubsuber, err := factory.Create(logger.NewLogger(&dummyLogger{}, func(e error) {}), conf.Type)
	assert.NotNil(t, err)
	assert.Nil(t, pubsuber)
}
