package metrics

import (
	"github.com/run-ai/runai/runai-common-packages/metrics"
)

const NameConnectionActive = "connections_active"

type Active struct {
	meta *EndpointMeta
}

func NewActive(info *EndpointMeta) *Active {
	return &Active{meta: info}
}

func (ca *Active) Name() string {
	return NameConnectionActive
}

func (ca *Active) Description() string {
	return "Number of currently open connections"
}

func (ca *Active) Labels() []string {
	return ca.meta.Labels()
}

func (ca *Active) LabelMap() map[string]string {
	return ca.meta.LabelMap()
}

func (ca *Active) Inc() {
	metrics.GaugeVec(ca.Name()).Get().WithLabels(ca.LabelMap()).Inc()
}

func (ca *Active) Dec() {
	metrics.GaugeVec(ca.Name()).Get().WithLabels(ca.LabelMap()).Dec()
}

func (ca *Active) Register() {
	metrics.GaugeVec(ca.Name()).Register(ca.Description(), ca.Labels())
}
