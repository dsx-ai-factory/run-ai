//go:build integration
// +build integration

package tests

import (
	"context"
	"encoding/json"
	"io"

	"github.com/run-ai/runai/runai-common-packages/logging"

	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	pubsub_config "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/kafka"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/nats"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
	"github.com/stretchr/testify/assert"
)

type integrationTestEvent struct {
	Data string
}

type TestServiceDlq struct {
	pevent.DlqMessage
}

const (
	waitForWantedStateSeconds        int = 20
	waitForStateStabilizationSeconds int = 2

	subscribeRetries = 2
)

func setUpIntegrationTest(t *testing.T) {
	setEnvs(t)

	logging.Init(nil)
	runaipubsub.SetLogger(logging.Actual())
	err := runaipubsub.Init(func(e error) {
		fmt.Printf("Error: %v\n", e)
	})
	if err != nil {
		fmt.Printf("Error when initializing runaipubsub: %v\n", err)
		os.Exit(1)
	}
}

func setEnvs(t *testing.T) {
	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "kafka"))
	t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "localhost:9092"))
	// t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "nats"))
	// t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "nats://localhost:4222"))
	t.Setenv("PUBSUB_KAFKA_AUTH_TYPE", "plain")
	t.Setenv("PUBSUB_ENABLE_TLS", "false")
	t.Setenv("PUBSUB_REPLICAS", "1")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("PUBSUB_SUBSCRIBE_RETRIES", fmt.Sprintf("%d", subscribeRetries))
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../config/secrets/"))
	t.Setenv("PUBSUB_WAIT_FOR_ACK_SEC", "1")
	t.Setenv("PUBSUB_ENABLE_DLQ", "false")
}

func tearDownIntegrationTest() {
	defer config.FlushAll()
	err := runaipubsub.Close()
	if err != nil {
		fmt.Printf("Error when closing runaipubsub: %v\n", err)
		os.Exit(1)
	}
}

// Please keep this test a the top of the file.
// It must run first because it creates a new consumer group, which causes it to consume all messages of previous tests.
func TestSubscribeTwoGroupsAndThenPublish(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Sub-->Pub"
	count := 0

	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestSubscribeTwoGroupsAndThenPublish: First group: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	cfg := config.Get[*pubsub_config.Pubsub]()
	pb := createPubsuber()
	err = pb.Connect(cfg.Url, cfg.User, cfg.Password)
	assert.NoError(t, err)
	defer pb.Close()

	cfg.ServiceName = "test-service-2"
	config.Set(cfg)
	defer func() {
		cfg.ServiceName = "test-service"
		config.Set(cfg)
	}()
	err = pb.Subscribe(context.Background(), "integrationTestEvent", func(event []byte) error {
		count += 1
		fmt.Println("TestSubscribeTwoGroupsAndThenPublish: Second group: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	ev := &integrationTestEvent{
		Data: eventString,
	}
	err = runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 2
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 2
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestSubscribeAndThenPublish(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Sub-->Pub"
	count := 0
	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestSubscribeAndThenPublish: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	ev := &integrationTestEvent{
		Data: eventString,
	}

	err = runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestSubscribeAsyncAndThenPublish(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	eventString := "Async Subscribe Test"
	count := 0

	runaipubsub.SubscribeAsync(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestSubscribeAsyncAndThenPublish: time: ", time.Now().String(), " count: ", count)
		return nil
	})

	ev := &integrationTestEvent{
		Data: eventString,
	}
	assert.Eventually(t, func() bool {
		return runaipubsub.Publish(context.Background(), ev) == nil
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestPublishAndThenSubscribe(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Pub-->Sub"
	ev := &integrationTestEvent{
		Data: eventString,
	}
	count := 0

	err := runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestPublishAndThenSubscribe: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestNoResponseRedelivery(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Redelivery"
	ev := &integrationTestEvent{
		Data: eventString,
	}
	count := 0

	err := runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		count += 1
		fmt.Println("TestNoResponseRedelivery: time: ", time.Now().String(), " count: ", count)
		time.Sleep(2500 * time.Millisecond)
		return nil
	})
	assert.NoError(t, err)

	expectedTries := subscribeRetries + 1
	assert.Eventually(t, func() bool {
		return count == expectedTries
	}, time.Duration(waitForWantedStateSeconds*2)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > expectedTries
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestRejectNoRedelivery(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "NoRedelivery"
	ev := &integrationTestEvent{
		Data: eventString,
	}

	err := runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	count := 0
	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		count++
		fmt.Println("TestRejectNoRedelivery: time: ", time.Now().String(), " count: ", count)
		return fmt.Errorf("Nak %d", count)
	})
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestRejectRedelivery(t *testing.T) {
	t.Setenv("PUBSUB_RETRY_ON_REJECT", "true")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Redelivery"
	ev := &integrationTestEvent{
		Data: eventString,
	}

	err := runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)
	count := 0
	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		count++
		fmt.Println("TestRejectRedelivery time: ", time.Now().String(), " count: ", count)
		return fmt.Errorf("Nak %d", count)
	})
	assert.NoError(t, err)
	expectedTries := subscribeRetries + 1
	assert.Eventually(t, func() bool {
		return count == expectedTries
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > expectedTries
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestPanicRedelivery(t *testing.T) {
	t.Setenv("PUBSUB_RETRY_ON_REJECT", "true")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	eventString := "Redelivery"
	ev := &integrationTestEvent{
		Data: eventString,
	}

	err := runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	count := 0
	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		count++
		fmt.Println("TestPanicRedelivery: time: ", time.Now().String(), " count: ", count)
		panic("Panic")
	})
	assert.NoError(t, err)
	expectedTries := subscribeRetries + 1
	assert.Eventually(t, func() bool {
		return count == expectedTries
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > expectedTries
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestWithAdditionalPubSuber_AdditionalMessage(t *testing.T) {
	t.Setenv("PUBSUB_ADDITIONAL_SUBSCRIBER_TYPE", "redis")
	t.Setenv("PUBSUB_ADDITIONAL_SUBSCRIBER_URL", getEnvOrDefault("REDIS_PUBSUB_URL", "localhost:6379"))
	t.Setenv("PUBSUB_ENABLE_ADDITIONAL_SUBSCRIBER", "true")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	redisPubsuber := runaipubsub.GetAdditionalPubSuber()
	assert.NotNil(t, redisPubsuber)

	eventString := "Redis Only PubSub Message"
	count := 0
	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestWithAdditionalPubSuber_AdditionalMessage: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	ev := pevent.Event[integrationTestEvent]{
		Payload: &integrationTestEvent{
			Data: eventString,
		},
	}
	evBytes, err := json.Marshal(ev)
	assert.NoError(t, err)
	err = redisPubsuber.Publish("integrationTestEvent", evBytes)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestWithAdditionalPubSuber_OriginalMessage(t *testing.T) {
	t.Setenv("PUBSUB_ADDITIONAL_SUBSCRIBER_TYPE", "redis")
	t.Setenv("PUBSUB_ADDITIONAL_SUBSCRIBER_URL", getEnvOrDefault("REDIS_PUBSUB_URL", "localhost:6379"))
	t.Setenv("PUBSUB_ENABLE_ADDITIONAL_SUBSCRIBER", "true")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	eventString := "Original"

	count := 0
	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestWithAdditionalPubSuber_OriginalMessage: time: ", time.Now().String(), " count: ", count)
		return nil
	})
	assert.NoError(t, err)

	ev := &integrationTestEvent{
		Data: eventString,
	}
	err = runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
}

func TestSubscribe_FailedProcessing_DlqEnabled(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	cfg := config.Get[*pubsub_config.Pubsub]()
	cfg.EnableDLQ = true
	cfg.WaitForAckSec = 10
	cfg.RetryOnReject = true
	config.Set(cfg)

	eventString := "Event message"

	count := 0
	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		return fmt.Errorf("Failed processing")
	})
	assert.NoError(t, err)

	dlqCount := 0
	err = runaipubsub.Subscribe(func(ctx context.Context, dlqMessage *TestServiceDlq) error {
		assert.Equal(t, "test-service", dlqMessage.Consumer)
		assert.Equal(t, "Failed processing", dlqMessage.Error)
		assert.Equal(t, "integrationTestEvent", dlqMessage.OriginTopic)

		var runaiEvent pevent.Event[integrationTestEvent]
		err := json.Unmarshal(dlqMessage.Body, &runaiEvent)
		assert.NoError(t, err)
		assert.Equal(t, eventString, runaiEvent.Payload.Data)

		dlqCount++
		return nil
	})

	assert.NoError(t, err)

	ev := &integrationTestEvent{
		Data: eventString,
	}
	err = runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == subscribeRetries+1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Eventually(t, func() bool {
		return dlqCount == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
}

func TestSubscribe_FailedProcessing_DlqDisabled(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	cfg := config.Get[*pubsub_config.Pubsub]()
	cfg.WaitForAckSec = 10
	cfg.RetryOnReject = true
	config.Set(cfg)

	eventString := "Event message"

	count := 0
	err := runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		return fmt.Errorf("Failed processing")
	})
	assert.NoError(t, err)

	ev := &integrationTestEvent{
		Data: eventString,
	}
	err = runaipubsub.Publish(context.Background(), ev)
	assert.NoError(t, err)

	dlqCount := 0
	err = runaipubsub.Subscribe(func(ctx context.Context, dlqMessage *TestServiceDlq) error {
		dlqCount += 1
		return fmt.Errorf("shouldn't be called")
	})
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return count == subscribeRetries+1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)

	time.Sleep(2 * time.Second)
	assert.Equal(t, dlqCount, 0)
}

func TestSubscribe_HandlesDeprecatedEventField(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	eventString := "Event message"
	ev := pevent.New[integrationTestEvent]()
	ev.Event = &integrationTestEvent{
		Data: eventString,
	}

	jsonPayload, err := json.Marshal(ev)
	assert.NoError(t, err)

	err = runaipubsub.GetPubSuber().Publish("integrationTestEvent", jsonPayload)
	assert.NoError(t, err)

	done := make(chan bool)
	err = runaipubsub.Subscribe(func(ctx context.Context, event *integrationTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		done <- true
		return nil
	})
	assert.NoError(t, err)

	<-done
}

func TestRetryConnect(t *testing.T) {
	defer tearDownIntegrationTest()
	connectRetries := 1

	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "nats"))
	t.Setenv("PUBSUB_URL", "invalid-url")
	t.Setenv("PUBSUB_CONNECT_RETRIES", fmt.Sprintf("%d", connectRetries))
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../config/secrets/"))

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = originalStdout
	}()

	runaipubsub.SetLogger(&fmtLogger{})
	err := runaipubsub.Init(func(e error) {})
	assert.Error(t, err)

	w.Close()
	out, _ := io.ReadAll(r)

	capturedOutput := string(out)

	count := strings.Count(capturedOutput, "Failed to connect to pubsub")
	assert.Equal(t, connectRetries+1, count) // +1 for the initial connection attempt
}

func TestIsInitialized_OnInitReturnsTrue(t *testing.T) {
	defer tearDownIntegrationTest()

	assert.False(t, runaipubsub.IsInitialized())

	setUpIntegrationTest(t)

	assert.True(t, runaipubsub.IsInitialized())
}

func getEnvOrDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}

func createPubsuber() runaipubsub.PubSuber {
	cfg := config.Get[*pubsub_config.Pubsub]()
	if cfg.Type == "kafka" {
		return kafka.NewPubSuber(logger.NewLogger(logging.Actual(), func(e error) {
			fmt.Printf("Error: %v\n", e)
		}))
	}
	return nats.NewPubSuber(logger.NewLogger(logging.Actual(), func(e error) {
		fmt.Printf("Error: %v\n", e)
	}))
}

type fmtLogger struct{}

func (*fmtLogger) Debugf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (*fmtLogger) Infof(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (*fmtLogger) Warnf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (*fmtLogger) Errorf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}
