package notify

import (
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/require"
)

func TestNewSink_HistoryDisabled_ReturnsNoOpHistory(t *testing.T) {
	sc := &sinks.SinkConfig{History: &sinks.HistorySinkConfig{HistoryEnabled: false}}

	s, err := newSink("my_history", sc)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, sinks.HistorySinkType, s.Type())

	_, ok := s.(*sinks.HistorySink)
	require.True(t, ok)

	require.NoError(t, s.Verify())
	require.NoError(t, s.Sink(&events.Event{UID: "uid-1"}))
}
