package metrics

import "github.com/run-ai/runai/runai-common-packages/metrics"

const (
	NameConnectionDuration = "connections_duration_seconds"
)

type LifecycleDuration struct {
	meta *EndpointMeta
}

func NewLifecycleDuration(info *EndpointMeta) *LifecycleDuration {
	return &LifecycleDuration{meta: info}
}

func (ld *LifecycleDuration) Name() string {
	return NameConnectionDuration
}

func (ld *LifecycleDuration) Description() string {
	return "Duration of interactive sessions"
}

func (ld *LifecycleDuration) Labels() []string {
	return GetLabels(ld.LabelMap("", ""))
}

func (ld *LifecycleDuration) LabelMap(status, reason string) map[string]string {
	dm := ld.meta.LabelMap()
	dm[LabelsStatus] = status
	dm[LabelsReason] = reason

	return dm
}

func (ld *LifecycleDuration) Report(status, reason string, duration float64) {
	metrics.HistogramVec(ld.Name()).Get().WithLabels(ld.LabelMap(status, reason)).Observe(duration)
}

func (ld *LifecycleDuration) Register() {
	// histogram value: 0.1, 0.5, 1, 5, 10, 10m, 1h, 1d, above
	buckets := []float64{0.1, 0.5, 1, 5, 10, 600, 3600, 86400}
	metrics.HistogramVec(ld.Name()).Register(ld.Description(), ld.Labels(), buckets)
}
