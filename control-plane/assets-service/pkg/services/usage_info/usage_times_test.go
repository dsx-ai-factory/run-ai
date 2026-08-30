package usage_info

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

var _ = Describe("Test usage times logic", func() {
	Context("no patching info", func() {
		It("should return with no errors", func() {
			err := PatchUsageTimes(context.Background(), 1, auth.WorkloadTypeTraining, nil)
			Expect(err).To(BeNil())
		})
	})

	Context("empty patching info", func() {
		It("should return with no errors", func() {
			err := PatchUsageTimes(context.Background(), 1, auth.WorkloadTypeTraining, make(map[string]response_patching.ResponsePatchingPtrs))
			Expect(err).To(BeNil())
		})
	})

	Context("missing project", func() {
		It("should return with no error after doing no work", func() {
			patchInfo := map[string]response_patching.ResponsePatchingPtrs{
				"dummyId": {},
			}
			err := PatchUsageTimes(context.Background(), 0, auth.WorkloadTypeTraining, patchInfo)
			Expect(err).To(BeNil())
		})
	})

	Context("missing workload type", func() {
		It("should return with no error after doing no work", func() {
			patchInfo := map[string]response_patching.ResponsePatchingPtrs{
				"dummyId": {},
			}
			err := PatchUsageTimes(context.Background(), 0, "", patchInfo)
			Expect(err).To(BeNil())
		})
	})
})
