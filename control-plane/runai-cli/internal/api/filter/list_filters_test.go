package filter_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	apifilter "github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
)

var _ = Describe("GetFrameworkFilters", func() {
	Context("FilterBuilder", func() {
		It("Set created at to milliseconds", func() {
			fb := apifilter.NewBuilder()
			fb.WithCreatedAt(time.Date(1, time.January, 1, 1, 1, 1, 1, time.Local))
			fbList := fb.StringList()
			Expect(fbList).To(HaveLen(1))
			Expect(fbList[0]).To(ContainSubstring("T01:01:01.000000001"))
		})
	})

	Context("GetFrameworkFilters", func() {
		It("returns filters for distributed frameworks", func() {
			frameworks := map[api.Framework]string{
				api.MPI:     "MPI||MPIJob",
				api.PyTorch: "PyTorch||PyTorchJob",
				api.TF:      "TF||TFJob",
				api.XGBoost: "XGBoost||XGBoostJob",
				api.Jax:     "JAX||JAXJob",
			}
			for framework, expectedValue := range frameworks {
				filters := apifilter.GetFrameworkFilters(framework)
				Expect(filters).To(HaveLen(1))
				Expect(filters[0].Type()).To(Equal(apifilter.DistributedFramework))
				Expect(filters[0].Value()).To(Equal(expectedValue))
				Expect(filters[0].Get()).To(Equal(string(apifilter.DistributedFramework) + "==" + expectedValue))
			}
		})

		It("excludes both short and legacy distributed framework aliases from standard training", func() {
			filters := apifilter.GetFrameworkFilters(api.Runai)
			Expect(filters).To(HaveLen(10))

			expectedFilters := []string{
				"distributedFramework!=MPI",
				"distributedFramework!=MPIJob",
				"distributedFramework!=PyTorch",
				"distributedFramework!=PyTorchJob",
				"distributedFramework!=JAX",
				"distributedFramework!=JAXJob",
				"distributedFramework!=TF",
				"distributedFramework!=TFJob",
				"distributedFramework!=XGBoost",
				"distributedFramework!=XGBoostJob",
			}
			actualFilters := make([]string, 0, len(filters))
			for _, filter := range filters {
				actualFilters = append(actualFilters, filter.Get())
			}
			Expect(actualFilters).To(ConsistOf(expectedFilters))
		})

		It("returns empty list for any-distributed frameworks", func() {
			frameworks := []api.Framework{api.AnyDistributed}
			for _, framework := range frameworks {
				filters := apifilter.GetFrameworkFilters(framework)
				Expect(filters).To(BeEmpty())
			}
		})
	})
})
