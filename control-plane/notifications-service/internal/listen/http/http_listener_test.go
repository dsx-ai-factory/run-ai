package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/http/openapi/generated"
	"github.com/stretchr/testify/assert"
)

func TestNotify_OK(t *testing.T) {
	eventCh, cancelCtx, server := startListener()
	defer cancelCtx()

	req := createRequest(t, `{"tenant_id": "123", "event_type": "type", "properties": {}}`, server.URL)
	client := http.Client{}
	response, err := client.Do(req)
	assert.Nil(t, err)

	expectedEvent := &events.Event{
		RawEvent:   &api.Notification{TenantId: "123", EventType: "type", Properties: make(map[string]interface{})},
		UID:        events.UID("123"),
		ListenerId: "h_listener",
		Properties: make(map[string]interface{}),
	}
	assert.Equal(t, response.StatusCode, http.StatusOK)
	event := <-eventCh
	assert.Equal(t, expectedEvent.ListenerId, event.ListenerId)
	assert.Equal(t, expectedEvent.Properties, event.Properties)
	assert.Equal(t, expectedEvent.RawEvent, event.RawEvent)
}

func TestNotify_QueueIsFull(t *testing.T) {
	_, cancelCtx, server := startListener()
	defer cancelCtx()

	client := http.Client{}
	req := createRequest(t, `{"tenant_id": "123", "event_type": "type", "properties": {}}`, server.URL)
	client.Do(req)
	req = createRequest(t, `{"tenant_id": "123", "event_type": "type", "properties": {}}`, server.URL)
	response, err := client.Do(req)

	assert.Nil(t, err)
	assert.Equal(t, response.StatusCode, http.StatusTooManyRequests)
}

func TestNotify_Invalid(t *testing.T) {
	_, cancelCtx, server := startListener()
	defer cancelCtx()

	req := createRequest(t, `{"tenant_id": "123"}`, server.URL)
	client := http.Client{}
	response, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
}

func TestListen_StopsOnCtxCanceled(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	eventCh := make(chan *events.Event, 1)

	testListener := NewHttpListener("h_listener", eventCh, &HttpListenerConfig{Port: 0})
	called := make(chan struct{})
	testListener.listenAndServeFunc = func() error {
		called <- struct{}{}
		return nil
	}
	testListener.closeServer = func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	}

	go testListener.Listen(ctx)

	time.Sleep(30 * time.Millisecond)
	cancelCtx()

	waitForCall(called)
	waitForCall(called)
	assert.Equal(t, 0, len(called))
}

func waitForCall(called chan struct{}) bool {
	select {
	case <-called:
		return true
	case <-time.After(1 * time.Second):
		return false
	}
}

func createRequest(t *testing.T, eventStr string, url string) *http.Request {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notify", url), strings.NewReader(eventStr))
	req.Header.Set("Content-Type", "application/json")
	assert.Nil(t, err)
	return req
}

func startListener() (chan *events.Event, context.CancelFunc, *httptest.Server) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	eventCh := make(chan *events.Event, 1)

	testListener := NewHttpListener("h_listener", eventCh, &HttpListenerConfig{Port: 0})
	server := httptest.NewServer(createRouter("h_listener", eventCh))

	testListener.listenAndServeFunc = func() error {
		return nil
	}

	testListener.closeServer = func(ctx context.Context) error {
		return nil
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancelCtx()
	}()

	testListener.Listen(ctx)

	return eventCh, cancelCtx, server
}
