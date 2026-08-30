package workloads

import (
	"context"
	"fmt"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	fake "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	existingNodepool1   = "default"
	existingNodepool2   = "testpool"
	nonExistingNodepool = "blabla"
)

var _ = Describe("ExistingResourcesValidator", func() {
	var (
		k8sClient client.Client
		validator *ExistingNodepoolsValidator
		ctx       context.Context
		scheme    *runtime.Scheme
	)

	scheme = runtime.NewScheme()
	register.InitSchema(scheme)
	ctx = context.Background()
	k8sClient = fake.InitFakeClientWithObjectNoStatus(scheme, getTestNodepools()...)
	validator = NewExistingNodepoolsValidator(k8sClient)

	DescribeTable("workload existing resources validation",
		func(nodepool string, nodepools []string, expectedErrorMsg string) {
			workload := getWorkloadWithNodepools(nodepool, nodepools)

			err := validator.Validate(ctx, workload)
			if len(expectedErrorMsg) > 0 {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedErrorMsg))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("resources exist (NodePools)", "", []string{existingNodepool1, existingNodepool2}, ""),
		Entry("resources exist (NodePool)", existingNodepool1, []string{}, ""),
		Entry("NodePool does not exist", nonExistingNodepool, []string{}, fmt.Sprintf(nodepoolNotFoundTemplate, nonExistingNodepool)),
		Entry("one of NodePools does not exist", "", []string{existingNodepool1, nonExistingNodepool}, fmt.Sprintf(nodepoolNotFoundTemplate, nonExistingNodepool)),
	)
})

func getWorkloadWithNodepools(nodepool string, nodepools []string) *jobs.JobWorkload {
	return &jobs.JobWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "w",
			Namespace: "namespace",
		},
		Spec: jobs.JobSpec{
			RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
				NodePools: field_types.NewStringField(strings.Join(nodepools, nodepoolsSpecSeparator)),
				NodePool:  field_types.NewStringField(nodepool),
			},
		},
	}
}

func getTestNodepools() []client.Object {
	nodepool1 := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: existingNodepool1,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}

	nodepool2 := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: existingNodepool2,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}

	// kai.resources twins of the fixtures, so these specs resolve the same node pools
	// under both values of the compile-time UseKaiNodePools flag.
	// TODO: drop the legacy nodepool1/nodepool2 once clusterconstants.UseKaiNodePools=true
	kaiNodepool1 := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: existingNodepool1,
		},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}

	kaiNodepool2 := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: existingNodepool2,
		},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}

	return []client.Object{nodepool1, nodepool2, kaiNodepool1, kaiNodepool2}
}
