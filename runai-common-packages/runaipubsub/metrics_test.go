package runaipubsub

import (
	"errors"
	"testing"

	metrics_testing "github.com/run-ai/runai/runai-common-packages/metrics/testing"
	"github.com/stretchr/testify/require"
)

func TestNewPubsubMetrics(t *testing.T) {
	psMetrics := newPubsubMetrics()
	require.NotNil(t, psMetrics)
}

func TestRegister(t *testing.T) {
	manager := metrics_testing.SetMock(t)
	manager.CounterVec(pubsubEventCount).Register("Count of pubsub events processed by subscriber", metricLabels)

	psMetrics := newPubsubMetrics()
	psMetrics.register()
}

func TestRecordEvent(t *testing.T) {
	manager := metrics_testing.SetMock(t)
	manager.CounterVec(pubsubEventCount).Register("Count of pubsub events processed by subscriber", metricLabels)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "false",
	}).Add(0)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "true",
	}).Add(0)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "false",
	}).Inc()

	psMetrics := newPubsubMetrics()
	psMetrics.register()
	psMetrics.registerEvent("test_topic")
	psMetrics.recordEvent("test_topic", nil)
}

func TestRecordEvent_Error(t *testing.T) {
	manager := metrics_testing.SetMock(t)
	manager.CounterVec(pubsubEventCount).Register("Count of pubsub events processed by subscriber", metricLabels)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "false",
	}).Add(0)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "true",
	}).Add(0)
	manager.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  "test_topic",
		errorLabel: "true",
	}).Inc()

	psMetrics := newPubsubMetrics()
	psMetrics.register()
	psMetrics.registerEvent("test_topic")
	psMetrics.recordEvent("test_topic", errors.New("error"))
}
