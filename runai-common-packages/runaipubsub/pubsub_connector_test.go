package runaipubsub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	logger_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/logger/mocks"
	mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConnect_AdditionalPubsuberNotInitialized(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type: config_models.PSTRedis,
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Nil(t, err)

	assert.Equal(t, primaryPubSuberMock, pubSuberInstance)
	assert.Nil(t, additionalPubSuberInstance)
	factoryMock.AssertNotCalled(t, "Create", log, config_models.PSTNats)
}

func TestConnect_AdditionalPubsuberInitialized(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Nil(t, err)

	assert.Equal(t, primaryPubSuberMock, pubSuberInstance)
	assert.Equal(t, additionalPubSuberMock, additionalPubSuberInstance)
}

func TestConnect_AdditionalPubsuberDisabled(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: false,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Nil(t, err)

	assert.Equal(t, primaryPubSuberMock, pubSuberInstance)
	assert.Nil(t, additionalPubSuberInstance)
}

func TestConnect_FailedToConnectToPrimaryPubsuber(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Error(t, err)

	assert.Nil(t, pubSuberInstance)
	assert.Equal(t, additionalPubSuberMock, additionalPubSuberInstance)
}

func TestConnect_ConnectedAfterMultipleAttempts_WontReturnError(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: false,
		ConnectRetries:             3,
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect")).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect again")).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.NoError(t, err)
}

func TestConnect_FailedAfterMultipleAttempts_ReturnError(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: false,
		ConnectRetries:             2,
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect")).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect again")).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect again")).Once()

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Error(t, err)
}

func TestConnect_SucceededToBoth(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	pubSuberTemp, additionalPubSuberTemp := mocks.NewPubSuber(t), mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberTemp
	additionalPubSuberInstance = additionalPubSuberTemp
	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	pubSuberTemp.On("Close").Return(nil).Once()
	additionalPubSuberTemp.On("Close").Return(nil).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Nil(t, err)

	assert.Equal(t, primaryPubSuberMock, pubSuberInstance)
	assert.Equal(t, additionalPubSuberMock, additionalPubSuberInstance)
}

func TestConnect_FailedToConnectToBothPubsuber(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.Error(t, err)
	fmt.Printf("errors: %+v", err)
	assert.Contains(t, err.Error(), "failed to connect pubsub instance, is_additional: false")
}

func TestConnect_AdditionalSameUrl(t *testing.T) {
	logs := make(chan interface{}, 10)
	log := logger_mocks.NewLoggerer(t)
	loggerStub := func(args mock.Arguments) {
		logs <- args[0]
	}

	log.On("Warnf", mock.Anything, mock.Anything).Run(loggerStub)
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := getPubsubConnector(log, factoryMock).Connect()
	assert.NoError(t, err)
	l := <-logs
	assert.Contains(t, l.(string), "failed to validate additional subscriber config")
	require.Eventually(t, func() bool {
		return pubSuberInstance == primaryPubSuberMock
	}, 3*time.Second, 100*time.Millisecond, "pubSuberInstance did not match primaryPubSuberMock within timeout")
	require.Eventually(t, func() bool {
		return additionalPubSuberInstance == nil
	}, 3*time.Second, 100*time.Millisecond, "additional with invalid config")
}

func TestConnectAsync_AdditionalPubsuberInitialized(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	pubSuberTemp, additionalPubSuberTemp := mocks.NewPubSuber(t), mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberTemp
	additionalPubSuberInstance = additionalPubSuberTemp
	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	pubSuberTemp.On("Close").Return(nil).Once()
	additionalPubSuberTemp.On("Close").Return(nil).Once()
	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	getPubsubConnector(log, factoryMock).ConnectAsync(context.Background())
	require.Eventually(t, func() bool {
		return pubSuberInstance == primaryPubSuberMock
	}, 3*time.Second, 100*time.Millisecond, "pubSuberInstance did not match primaryPubSuberMock within timeout")
	require.Eventually(t, func() bool {
		return additionalPubSuberMock == additionalPubSuberInstance
	}, 3*time.Second, 100*time.Millisecond, "pubSuberInstance did not match primaryPubSuberMock within timeout")
}

func TestConnectAsync_AdditionalPubsuberDisabled(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: false,
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	getPubsubConnector(log, factoryMock).ConnectAsync(ctx)
	require.Eventually(t, func() bool {
		return pubSuberInstance == primaryPubSuberMock
	}, 3*time.Second, 100*time.Millisecond, "pubSuberInstance did not match primaryPubSuberMock")
	assert.Nil(t, additionalPubSuberInstance)
}

func TestConnectAsync_FailedToConnectToPrimaryPubsuber(t *testing.T) {
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost_else",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	errs := getPubsubConnector(log, factoryMock).ConnectAsync(ctx)
	err := <-errs
	assert.Error(t, err)

	assert.Nil(t, pubSuberInstance)
	require.Eventually(t, func() bool {
		return additionalPubSuberMock == additionalPubSuberInstance
	}, 3*time.Second, 100*time.Millisecond, "pubSuberInstance did not match primaryPubSuberMock within timeout")
	cancel()
}

func TestConnectAsync_FailedToConnectToAllPubsuber(t *testing.T) {
	factoryMock := NewMockPubSubFactoryer(t)
	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)
	log := logger.NewLogger(&dummyLogger{}, func(e error) {})
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to connect"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := getPubsubConnector(log, factoryMock).ConnectAsync(ctx)
	mainErr, additionalErr := false, false
	for done := false; !done; {
		select {
		case err := <-errs:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to connect pubsub instance, is_additional: ")
			if strings.Contains(err.Error(), "is_additional: true") {
				additionalErr = true
			}
			if strings.Contains(err.Error(), "is_additional: false") {
				mainErr = true
			}

			done = mainErr && additionalErr
		case <-time.After(3 * time.Second):
			done = true
		}
	}

	assert.True(t, mainErr)
	assert.True(t, additionalErr)
	assert.Nil(t, pubSuberInstance)
	assert.Nil(t, additionalPubSuberInstance)
}

func TestConnectAsync_AllFailedCtxCanceled(t *testing.T) {
	called := make(chan interface{}, 10)
	log := logger_mocks.NewLoggerer(t)
	loggerStub := func(args mock.Arguments) {
		called <- args[0]
	}
	log.On("Infof", mock.Anything, mock.Anything, mock.Anything).Run(loggerStub)
	log.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost1",
	}
	config.Set(conf)
	defer reset()

	primaryPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberMock := mocks.NewPubSuber(t)

	factoryMock := NewMockPubSubFactoryer(t)
	factoryMock.On("Create", log, config_models.PSTRedis).Return(primaryPubSuberMock, nil)
	factoryMock.On("Create", log, config_models.PSTNats).Return(additionalPubSuberMock, nil)

	primaryPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed"))
	additionalPubSuberMock.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("additional failed too"))

	ctx, cancel := context.WithCancel(context.Background())
	errs := getPubsubConnector(log, factoryMock).ConnectAsync(ctx)

	for mainErr, additionalErr := false, false; !mainErr || !additionalErr; {
		err := <-errs
		mainErr = mainErr || strings.Contains(err.Error(), "is_additional: true")
		additionalErr = additionalErr || strings.Contains(err.Error(), "is_additional: false")
	}
	cancel()
	counter := 0
	for counter < 2 {
		l := <-called
		if strings.Contains(l.(string), "terminating") {
			counter += 1
		}
	}
	require.Equal(t, counter, 2)
}

func TestValidateAdditionalSubscriberConfig(t *testing.T) {
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTNats,
		AdditionalSubscriberUrl:    "localhost_other",
	}

	err := validateAdditionalSubscriberConfig(conf)
	require.NoError(t, err)
}

func TestValidateAdditionalSubscriberConfig_TypeEmpty(t *testing.T) {
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   "",
		AdditionalSubscriberUrl:    "localhost_other",
	}

	err := validateAdditionalSubscriberConfig(conf)
	require.Equal(t, err.Error(), "additional subscriber type is empty")
}

func TestValidateAdditionalSubscriberConfig_EmptyUrl(t *testing.T) {
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTRedis,
		AdditionalSubscriberUrl:    "",
	}
	err := validateAdditionalSubscriberConfig(conf)
	require.Equal(t, err.Error(), "additional subscriber url is empty")
}

func TestValidateAdditionalSubscriberConfig_SameUrl(t *testing.T) {
	conf := &config_models.Pubsub{
		Type:                       config_models.PSTRedis,
		Url:                        "localhost",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   config_models.PSTRedis,
		AdditionalSubscriberUrl:    "localhost",
	}

	err := validateAdditionalSubscriberConfig(conf)
	require.Equal(t, err.Error(), "additional subscriber url is the same as the main subscriber url")
}

func getPubsubConnector(l logger.Loggerer, factoryMock PubSubFactoryer) *pubsubConnector {
	return &pubsubConnector{
		pubSubFactory: factoryMock,
		logger:        l,
		errors:        make(chan error, 10),
	}
}
