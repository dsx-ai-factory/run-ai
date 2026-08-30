package runaipubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
)

var (
	resubscribeHandler         = NewResubscribeHandler()
	psMetrics                  = newPubsubMetrics()
	once                       sync.Once
	log                        logger.Loggerer = &dummyLogger{}
	pubSuberInstance           PubSuber
	additionalPubSuberInstance PubSuber
	pubSubConnector            PubsubConnectorer
	pubsubEnabled              bool
	closeContext               context.Context = context.Background()
	cancelCloseContext         context.CancelFunc
	errChan                    chan error = make(chan error, 10)
)

// SetLogger should be called before init
// else will use default logger without logs
func SetLogger(logging logger.Loggerer) {
	if logging == nil {
		return
	}

	log = logging
}

func Init(errorCallback func(e error)) error {
	log := getLogger(errorCallback)
	psMetrics.register()

	var err error
	once.Do(func() {
		closeContext, cancelCloseContext = context.WithCancel(context.Background())
		if err = config_models.InitConfig(getOnConfigChangeCallback(log)); err != nil {
			log.Errorf("failed to init configuration, err %+v", err)
			return
		}
		if pubsubEnabled = ptr.ToBool(config.Get[*config_models.Pubsub]().Enabled); !pubsubEnabled {
			return
		}
		pubSubConnector = newPubsubConnector(log)
		err = pubSubConnector.Connect()
	})

	return err
}

func InitAsync(errorCallback func(e error)) (chan error, error) {
	log := getLogger(errorCallback)
	psMetrics.register()

	var configErr error
	once.Do(func() {
		closeContext, cancelCloseContext = context.WithCancel(context.Background())
		if err := config_models.InitConfig(getOnConfigChangeCallback(log)); err != nil {
			log.Errorf("failed to init configuration, err %+v", err)
			configErr = err
			return
		}
		if pubsubEnabled = ptr.ToBool(config.Get[*config_models.Pubsub]().Enabled); !pubsubEnabled {
			return
		}
		pubSubConnector = newPubsubConnector(log)
		errChan = pubSubConnector.ConnectAsync(closeContext)
	})

	return errChan, configErr
}

func Publish[T any](ctx context.Context, event *T) error {
	if !pubsubEnabled {
		return nil
	}
	if !IsInitialized() {
		return fmt.Errorf("pubsub not initialized")
	}

	runaiEvent := pevent.New[T]().WithCtx(ctx).WithPayload(event)
	jsonPayload, err := json.Marshal(runaiEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event : %v", err)
	}
	return pubSuberInstance.Publish(runaiEvent.GetTopicName(), jsonPayload)
}

func Subscribe[T any](callback func(ctx context.Context, event *T) error) error {
	if !pubsubEnabled {
		return nil
	}
	if !IsInitialized() {
		return fmt.Errorf("pubsub not initialized")
	}

	topic := pevent.New[T]().GetTopicName()
	psMetrics.registerEvent(topic)
	finalCallback := getSubscribeCallback(topic, callback)

	err := pubSuberInstance.Subscribe(closeContext, topic, finalCallback)
	if err != nil {
		return err
	}
	log.Infof("Subscribed to topic %s", topic)

	subscribeToAdditionalPubSuber(closeContext, topic, finalCallback)
	resubscribeHandler.Add(closeContext, topic, finalCallback)

	return nil
}

func SubscribeAsync[T any](callback func(ctx context.Context, event *T) error) <-chan error {
	if !pubsubEnabled {
		return make(chan error)
	}
	topic := pevent.New[T]().GetTopicName()
	psMetrics.registerEvent(topic)
	finalCallback := getSubscribeCallback(topic, callback)
	go func() {
		if shouldSubscribe(closeContext, func() bool { return pubSuberInstance == nil }) {
			return
		}

		pubSuberInstance.SubscribeAsync(closeContext, topic, finalCallback, errChan)
	}()

	go func() {
		if enabled, _ := initializeAdditionalPubsuber(config.Get[*config_models.Pubsub]()); !enabled {
			return
		}

		if shouldSubscribe(closeContext, func() bool { return additionalPubSuberInstance == nil }) {
			return
		}
		additionalPubSuberInstance.SubscribeAsync(closeContext, topic, finalCallback, errChan)
	}()

	resubscribeHandler.Add(closeContext, topic, finalCallback)
	return errChan
}

func Close() error {
	once = sync.Once{}
	if cancelCloseContext != nil {
		cancelCloseContext()
	}

	if !pubsubEnabled {
		return nil
	}

	if additionalPubSuberInstance != nil {
		_ = additionalPubSuberInstance.Close()
		additionalPubSuberInstance = nil
	}

	if pubSuberInstance == nil {
		return nil
	}
	err := pubSuberInstance.Close()
	pubSuberInstance = nil
	return err
}

func IsInitialized() bool {
	return pubSuberInstance != nil
}

func SetMock(pubSuber PubSuber) {
	pubSuberInstance = pubSuber
	pubsubEnabled = true
}

func GetPubSuber() PubSuber {
	return pubSuberInstance
}

func GetAdditionalPubSuber() PubSuber {
	return additionalPubSuberInstance
}

func defaultErrorCallback(_ error) {
	// pass
}

func getOnConfigChangeCallback(log *logger.Logger) func() {
	return func() {
		if err := pubSubConnector.Connect(); err != nil {
			log.Errorf("Failed to reconnect pubsubers: %w", err)
			return
		}

		if err := resubscribeHandler.Resubscribe(); err != nil {
			log.Errorf("Failed to resubscribe: %w", err)
		}
	}
}

func subscribeToAdditionalPubSuber(ctx context.Context, topic string, finalCallback func(payloadJson []byte) error) {
	if additionalPubSuberInstance == nil {
		return
	}

	err := additionalPubSuberInstance.Subscribe(ctx, topic, finalCallback)
	if err != nil {
		log.Warnf("Failed to subscribe to additional pubsuber: %v", err)
	} else {
		log.Debugf("Subscribed to additional pubsuber on topic %s", topic)
	}
}

func getSubscribeCallback[T any](topic string, callback func(ctx context.Context, event *T) error) func(payloadJson []byte) error {
	return func(payloadJson []byte) error {
		log.Debugf("Received event on topic %s: %s", topic, string(payloadJson))
		runaiEvent := pevent.New[T]()
		err := json.Unmarshal(payloadJson, runaiEvent)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: %v", err)
		}
		ctx := runaiEvent.ExtractCtx()
		err = callback(ctx, runaiEvent.GetPayload())
		psMetrics.recordEvent(topic, err)
		return err
	}
}

func getLogger(errorCallback func(e error)) *logger.Logger {
	if errorCallback != nil {
		return logger.NewLogger(log, errorCallback)
	}

	return logger.NewLogger(log, defaultErrorCallback)
}

func shouldSubscribe(ctx context.Context, cond func() bool) (terminate bool) {
	for cond() {
		select {
		case <-ctx.Done():
			return true
		case <-time.After(time.Second):
			continue
		}
	}
	return false
}
