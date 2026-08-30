package metrics

import "github.com/run-ai/runai/runai-common-packages/metrics"

const (
	MessageSentByteName     = "message_sent_byte"
	MessageReceivedByteName = "message_received_byte"
)

// byte size histogram
var messageSizeHistogram = []float64{128, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}

func MessageSentReport(endpointMeta *EndpointMeta, value uint64) {
	if endpointMeta == nil {
		return
	}

	metrics.HistogramVec(MessageSentByteName).Get().WithLabels(endpointMeta.LabelMap()).Observe(float64(value))
}

func MessageReceivedReport(endpointMeta *EndpointMeta, value uint64) {
	if endpointMeta == nil {
		return
	}

	metrics.HistogramVec(MessageReceivedByteName).Get().WithLabels(endpointMeta.LabelMap()).Observe(float64(value))
}

func ConnectionMessageRegister() {
	endpointMeta := &EndpointMeta{}
	metrics.HistogramVec(MessageSentByteName).Register(
		"Total number of messages sent by the server", endpointMeta.Labels(), messageSizeHistogram)
	metrics.HistogramVec(MessageReceivedByteName).Register(
		"Total number of messages received from clients", endpointMeta.Labels(), messageSizeHistogram)
}
