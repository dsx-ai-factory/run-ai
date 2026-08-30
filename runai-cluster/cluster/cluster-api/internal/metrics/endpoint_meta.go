package metrics

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
)

type (
	Interactivity string
	Path          string
)

const (
	LabelPath              = "path"
	LabelInteractive       = "interactive"
	LabelWorkloadType      = "type"
	LabelWorkloadFramework = "framework"

	LabelsStatus = "status"
	LabelsReason = "reason"

	FullInteractive Interactivity = "full"
	HalfInteractive Interactivity = "half"
	NoneInteractive Interactivity = "none"
)

type EndpointMeta struct {
	Path              string
	Interactivity     Interactivity
	WorkloadType      api.WorkloadType
	WorkloadFramework api.JobType
	TenantId          string
}

func (i *EndpointMeta) LabelMap() map[string]string {
	return map[string]string{
		LabelPath:              i.Path,
		LabelInteractive:       string(i.Interactivity),
		LabelWorkloadType:      string(i.WorkloadType),
		LabelWorkloadFramework: string(i.WorkloadFramework),
	}
}

func (i *EndpointMeta) Labels() []string {
	return GetLabels(i.LabelMap())
}

func GetLabels(labelMap map[string]string) []string {
	labels := make([]string, 0)
	for key := range labelMap {
		labels = append(labels, key)
	}

	return labels
}
