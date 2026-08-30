package workloads

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	fake "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace                 = "space"
	existingPvcInNamespace    = "pvc-in-namespace"
	existingPvcNotInNamespace = "pvc-not-in-namespace"
	nonExitingPvc             = "blabla"
)

type TestPvc struct {
	Name      string
	Namespace string
	Existing  bool
}

var _ = Describe("ExistingResourcesValidator", func() {
	var (
		k8sClient client.Client
		validator *ExistingPvcsValidator
		ctx       context.Context
		scheme    *runtime.Scheme
	)

	scheme = runtime.NewScheme()
	register.InitSchema(scheme)
	ctx = context.Background()
	k8sClient = fake.InitFakeClientWithObjectNoStatus(scheme, getTestPvcs()...)
	validator = NewExistingPvcsValidator(k8sClient)

	DescribeTable("workload existing resources validation",
		func(testPvcs []TestPvc, expectedErrorMsg string) {
			workload := getWorkloadWithPvcs(testPvcs)

			err := validator.Validate(ctx, workload)
			if len(expectedErrorMsg) > 0 {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedErrorMsg))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("resources exist (no pvcs)", []TestPvc{}, ""),
		Entry("pvc exists not in namespace", []TestPvc{getTestPvc(existingPvcNotInNamespace, namespace, true)}, fmt.Sprintf(pvcNotFoundTemplate, existingPvcNotInNamespace, namespace)),
		Entry("pvc does not exist", []TestPvc{getTestPvc(nonExitingPvc, namespace, true)}, fmt.Sprintf(pvcNotFoundTemplate, nonExitingPvc, namespace)),
	)
})

func getTestPvc(name string, namespace string, existing bool) TestPvc {
	return TestPvc{
		Name:      name,
		Namespace: namespace,
		Existing:  existing,
	}
}

func getWorkloadWithPvcs(testPvcs []TestPvc) *jobs.JobWorkload {
	pvcs := make(map[string]field_types.PvcsItemAndRules)
	for _, testPvc := range testPvcs {
		pvcs[testPvc.Name] = field_types.PvcsItemAndRules{
			Value: &field_types.PvcsItemValue{
				ExistingPvc: ptr.To(testPvc.Existing),
				ClaimName:   ptr.To(testPvc.Name),
			},
		}
	}

	return &jobs.JobWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "w",
			Namespace: namespace,
		},
		Spec: jobs.JobSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &pvcs,
					},
				},
			},
		},
	}
}

func getTestPvcs() []client.Object {
	pvcInNamespace := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      existingPvcInNamespace,
			Namespace: namespace,
		},
	}

	pvcNotInNamespace := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      existingPvcNotInNamespace,
			Namespace: "other-namespace",
		},
	}

	return []client.Object{pvcInNamespace, pvcNotInNamespace}
}
