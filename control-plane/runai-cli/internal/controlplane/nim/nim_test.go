package nim_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type filteringPaginator struct {
	workloads    []openapi.Workload
	ignoreFilter bool
}

func (p *filteringPaginator) List(_ context.Context, _, _ int32, filters []string, _ bool) (*openapi.GetWorkloads200Response, error) {
	for _, value := range filters {
		if !p.ignoreFilter && value == "type=="+api.BackendNimServiceWorkloadType {
			var filtered []openapi.Workload
			for _, wl := range p.workloads {
				if wl.Type == api.BackendNimServiceWorkloadType {
					filtered = append(filtered, wl)
				}
			}
			return &openapi.GetWorkloads200Response{Workloads: filtered}, nil
		}
	}

	return &openapi.GetWorkloads200Response{Workloads: p.workloads}, nil
}

var _ = Describe("NIM workload lookup", func() {
	BeforeEach(func() {
		workload.SetPaginator(&filteringPaginator{workloads: []openapi.Workload{
			{Id: "inference-id", Name: "chatbot", Type: api.Inference.QueryParam()},
			{Id: "nim-id", Name: "chatbot", Type: api.BackendNimServiceWorkloadType},
		}})
		DeferCleanup(func() { workload.SetPaginator(nil) })
	})

	It("selects the NIM service when an inference workload has the same name", func() {
		fb := filter.NewBuilder().WithWorkloadType(api.Inference.QueryParam())

		out, err := nim.GetByNameAndProj(transaction.NewTxn(), "", "", "chatbot", "demo", "cluster-id", fb)

		Expect(err).NotTo(HaveOccurred())
		Expect(out.GetList()).To(ConsistOf(HaveField("Type", api.BackendNimServiceWorkloadType)))
	})

	It("reports workload not found when only a standard inference workload exists", func() {
		workload.SetPaginator(&filteringPaginator{workloads: []openapi.Workload{
			{Id: "inference-id", Name: "llama-repro", Type: api.Inference.QueryParam()},
		}})
		fb := filter.NewBuilder().WithWorkloadType(api.Inference.QueryParam())

		out, err := nim.GetByNameAndProj(transaction.NewTxn(), "", "", "llama-repro", "thea", "cluster-id", fb)

		Expect(err).To(MatchError("workload not found"))
		Expect(out).To(BeNil())
	})

	It("rejects a non-NIM workload returned by the backend", func() {
		workload.SetPaginator(&filteringPaginator{
			ignoreFilter: true,
			workloads: []openapi.Workload{
				{Id: "inference-id", Name: "chatbot", Type: api.Inference.QueryParam()},
			},
		})
		fb := filter.NewBuilder().WithWorkloadType(api.Inference.QueryParam())

		out, err := nim.GetByNameAndProj(transaction.NewTxn(), "", "", "chatbot", "demo", "cluster-id", fb)

		Expect(err).To(MatchError(`workload "chatbot" is of type Inference, not a NIMService workload`))
		Expect(out).To(BeNil())
	})
})
