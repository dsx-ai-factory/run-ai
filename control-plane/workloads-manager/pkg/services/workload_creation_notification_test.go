package services

import (
	"context"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

var _ = Describe("workload creation notification", func() {
	DescribeTable("preserves the public distributed framework at the serialized boundary",
		func(kind, version, expectedFramework string) {
			qRecord := distributedQueueRecordForNotificationTest(kind, version)

			notification := workloadCreationNotificationFromQueueRecord(qRecord)

			Expect(notification.Type).To(Equal(string(workloadsApi.WORKLOADTYPE_DISTRIBUTED)))
			Expect(notification.Group).NotTo(BeNil())
			Expect(*notification.Group).To(Equal("kubeflow.org"))
			Expect(notification.Version).NotTo(BeNil())
			Expect(*notification.Version).To(Equal(version))
			Expect(notification.DistributedFramework).NotTo(BeNil())
			Expect(*notification.DistributedFramework).To(Equal(expectedFramework))

			payload, err := json.Marshal(notification)
			Expect(err).NotTo(HaveOccurred())

			var decoded map[string]any
			Expect(json.Unmarshal(payload, &decoded)).To(Succeed())
			Expect(decoded).To(HaveKeyWithValue("type", string(workloadsApi.WORKLOADTYPE_DISTRIBUTED)))
			Expect(decoded).To(HaveKeyWithValue("distributedFramework", expectedFramework))
		},
		Entry("PyTorch v1", "PyTorchJob", "v1", "PyTorch"),
		Entry("TF v1", "TFJob", "v1", "TF"),
		Entry("XGBoost v1", "XGBoostJob", "v1", "XGBoost"),
		Entry("JAX v1", "JAXJob", "v1", "JAX"),
		Entry("MPI v2beta1", "MPIJob", "v2beta1", "MPI"),
	)

	It("omits an unknown distributed framework", func() {
		notification := workloadCreationNotificationFromQueueRecord(distributedQueueRecordForNotificationTest("UnknownJob", "v1"))

		Expect(notification.Type).To(Equal(string(workloadsApi.WORKLOADTYPE_DISTRIBUTED)))
		Expect(notification.DistributedFramework).To(BeNil())
	})
})

var _ = Describe("MPIJob v2beta1 suspend actions", func() {
	It("patches spec.runPolicy.suspend for suspend and resume", func() {
		manifest := dao_utils.JsonBMap{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"spec": map[string]any{
				"runPolicy": map[string]any{},
			},
		}
		componentDefinition := kartav1alpha1.ComponentDefinition{
			SuspendDefinition: &kartav1alpha1.SuspendDefinition{
				SuspendActions: []kartav1alpha1.SuspendAction{{Path: ".spec.runPolicy.suspend", Value: "true"}},
				ResumeActions:  []kartav1alpha1.SuspendAction{{Path: ".spec.runPolicy.suspend", Value: "false"}},
			},
		}

		suspended, apiErr := patchManifestSuspendActions(context.Background(), manifest, componentDefinition, true)
		Expect(apiErr).To(BeNil())
		expectRunPolicySuspend(suspended, true)

		resumed, apiErr := patchManifestSuspendActions(context.Background(), suspended, componentDefinition, false)
		Expect(apiErr).To(BeNil())
		expectRunPolicySuspend(resumed, false)
	})
})

func distributedQueueRecordForNotificationTest(kind, version string) *models.WorkloadsQueueRecord {
	return &models.WorkloadsQueueRecord{
		ID:               "workload-id",
		CreatedAt:        time.Now(),
		CreatedBy:        "test@run.ai",
		ClusterId:        "cluster-id",
		ProjectId:        "project-id",
		ProjectName:      "project-name",
		DepartmentId:     "department-id",
		DepartmentName:   "department-name",
		Name:             "distributed-workload",
		ResourceGroup:    "kubeflow.org",
		ResourceVersion:  version,
		ResourceKind:     kind,
		TranslatedFromV1: true,
		RawManifest: dao_utils.JsonBMap{
			"apiVersion": "kubeflow.org/" + version,
			"kind":       kind,
			"metadata": map[string]any{
				"name": "distributed-workload",
				"labels": map[string]any{
					ConvertedFromNativeLabel:    "true",
					constants.LabelWorkloadKind: string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
				},
			},
		},
	}
}

func expectRunPolicySuspend(manifest dao_utils.JsonBMap, expected bool) {
	spec, ok := manifest["spec"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	runPolicy, ok := spec["runPolicy"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	Expect(runPolicy).To(HaveKeyWithValue("suspend", expected))
}
