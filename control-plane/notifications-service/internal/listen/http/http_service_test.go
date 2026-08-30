package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/http/openapi/generated"
	"github.com/stretchr/testify/assert"
)

func TestNotifyBatch_TooManyRequests(t *testing.T) {
	s := HttpService{
		eventCh:    make(chan *events.Event),
		listenerId: "id",
	}

	res, err := s.NotifyBatch(context.Background(), []api.Notification{{
		TenantId:   "",
		EventType:  "",
		Properties: nil,
	}})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusTooManyRequests, res.Code)
}

func TestNotifyBatch_Success(t *testing.T) {
	s := HttpService{
		eventCh:    make(chan *events.Event, 10),
		listenerId: "id",
	}

	res, err := s.NotifyBatch(context.Background(), []api.Notification{{
		TenantId:   "",
		EventType:  "",
		Properties: nil,
	}})

	assert.Nil(t, err)
	assert.Equal(t, api.Response(http.StatusOK, api.NotifyBatch200Response{Succeed: 1, Failed: 0, Total: 1}), res)
}
