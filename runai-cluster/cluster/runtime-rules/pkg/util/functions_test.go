package util

import (
	"fmt"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/logger"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const interactiveWorkloadKind = "InteractiveWorkload"

func TestRuntimeRules(t *testing.T) {
	logger.InitLogger(int(zap.DebugLevel))
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	t.Setenv("EVENT_LIST", strings.Trim(strings.Join(strings.Fields(fmt.Sprint([]int{3, 4})), ","), "[]"))

	RunSpecs(t, "functions Test Suite")
}

var _ = Describe("functions.go Test Suite", Ordered, func() {
	Context("getMessage function test", func() {
		It("Interactive Workload: check idle flow of events (naive)", func() {
			getEnv := GetEnvOrDefaultIntArray("EVENT_LIST", []int{})
			Expect(getEnv).To(Equal([]int{3, 4}))
		})
	})
	Context("GetWorkloadID function tests", func() {
		It("got annotation", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.UID = "uid"
			jobWorkload.Annotations = map[string]string{}
			jobWorkload.Annotations[constants.ControlPlaneWorkloadIdAnnotationForWorkload] = "test"
			workloadID, err := GetWorkloadID(jobWorkload)
			Expect(err).To(BeNil())
			Expect(workloadID).To(Equal(types.UID("test")))
		})
		It("no annotation", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.UID = "uid"
			workloadID, err := GetWorkloadID(jobWorkload)
			Expect(err).To(BeNil())
			Expect(workloadID).To(Equal(types.UID("uid")))
		})
		It("nil workload object", func() {
			workloadID, err := GetWorkloadID(nil)
			Expect(err).ToNot(BeNil())
			Expect(workloadID).To(Equal(types.UID("")))
		})
	})
	Context("IsJobPreemptible function tests", func() {
		It("nil workload object", func() {
			result := IsJobPreemptible(nil)
			Expect(result).To(BeFalse())
		})
		It("training workload - should be preemptible", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = "TrainingWorkload"
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeTrue())
		})
		It("distributed workload - should be preemptible", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = "DistributedWorkload"
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeTrue())
		})
		It("interactive workload with Preemptibility set to preemptible", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = interactiveWorkloadKind
			jobWorkload.Spec.Preemptibility = &fields.StringField{
				Value: ptr.To("preemptible"),
			}
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeTrue())
		})
		It("interactive workload with Preemptibility set to non-preemptible", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = interactiveWorkloadKind
			jobWorkload.Spec.Preemptibility = &fields.StringField{
				Value: ptr.To("non-preemptible"),
			}
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeFalse())
		})
		It("interactive workload with Preemptibility not initialized and Preemptible set to true", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = interactiveWorkloadKind
			jobWorkload.Spec.Preemptible = &fields.BoolField{ //nolint:staticcheck
				Value: ptr.To(true),
			}
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeTrue())
		})
		It("interactive workload with Preemptibility not initialized and Preemptible set to false", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = interactiveWorkloadKind
			jobWorkload.Spec.Preemptible = &fields.BoolField{ //nolint:staticcheck
				Value: ptr.To(false),
			}
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeFalse())
		})
		It("interactive workload with Preemptibility not initialized and Preemptible not set", func() {
			jobWorkload := &jobs.JobWorkload{}
			jobWorkload.Kind = interactiveWorkloadKind
			result := IsJobPreemptible(jobWorkload)
			Expect(result).To(BeFalse())
		})
	})
})
