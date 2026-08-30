package describe

import (
	"sort"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

type WorkloadBuilder struct {
	wd Workload
}

func NewWorkloadBuilder() *WorkloadBuilder {
	return &WorkloadBuilder{
		Workload{
			General:  displayer.General{},
			Runtime:  displayer.Runtime{},
			Compute:  displayer.Compute{},
			Networks: []displayer.Network{},
			Events:   []displayer.Event{},
			PodsInfo: displayer.PodsInfo{},
			Spec:     displayer.Spec{},
			Issues:   []displayer.Issue{},
		},
	}
}

func (b *WorkloadBuilder) SetGeneral(general displayer.General) *WorkloadBuilder {
	b.wd.General = general
	return b
}

func (b *WorkloadBuilder) SetRuntime(runtime displayer.Runtime) *WorkloadBuilder {
	b.wd.Runtime = runtime
	return b
}

func (b *WorkloadBuilder) SetCompute(compute displayer.Compute) *WorkloadBuilder {
	b.wd.Compute = compute
	return b
}

func (b *WorkloadBuilder) SetStorage(storage *displayer.Storage) *WorkloadBuilder {
	b.wd.Storage = storage
	return b
}

func (b *WorkloadBuilder) SetMasterConfig(config *displayer.WorkloadRuntimeConfig) *WorkloadBuilder {
	b.wd.MasterConfig = config
	return b
}

func (b *WorkloadBuilder) SetNetworks(networks []displayer.Network) *WorkloadBuilder {
	b.wd.Networks = networks
	return b
}

func (b *WorkloadBuilder) SetNetworkTopology(networkTopology *displayer.NetworkTopology) *WorkloadBuilder {
	b.wd.NetworkTopology = networkTopology
	return b
}

func (b *WorkloadBuilder) SetPodsInfo(podsInfo displayer.PodsInfo) *WorkloadBuilder {
	b.wd.PodsInfo = podsInfo
	return b
}

func (b *WorkloadBuilder) SetEvents(events []displayer.Event) *WorkloadBuilder {
	b.wd.Events = events
	return b
}

func (b *WorkloadBuilder) SetSpec(spec displayer.Spec) *WorkloadBuilder {
	b.wd.Spec = spec
	return b
}

func (b *WorkloadBuilder) SetIssues(issues []displayer.Issue) *WorkloadBuilder {
	b.wd.Issues = issues
	return b
}

func (b *WorkloadBuilder) GPUNetworkAcceleration(used *bool) *WorkloadBuilder {
	b.wd.GPUAcceleration = used
	return b
}

func (b *WorkloadBuilder) Build() *Workload {
	sort.Slice(b.wd.PodsInfo.Pods, func(i, j int) bool {
		return strings.Compare(b.wd.PodsInfo.Pods[i].Name, b.wd.PodsInfo.Pods[j].Name) < 0
	})
	return &b.wd
}
