package validator

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testDistributedWorkloadName = "test-distributed-workload"
	testNamespace               = "runai-default"
	testNonSystemUser           = "test@run.ai"
	testSystemUser              = "system:serviceaccount:agent"
	testNewName                 = "new-name"
)

type DistributedAssetsTest struct {
	k8sClient       client.Client
	currentWorkload *dinfw.DistributedInferenceWorkload
}

// Mock validator for testing
type mockDistributedInferenceValidator struct {
	shouldFail bool
	errorMsg   string
}

func (m *mockDistributedInferenceValidator) Validate(_ context.Context, job *dinfw.DistributedInferenceWorkload) error {
	if m.shouldFail {
		return fmt.Errorf("%s", m.errorMsg)
	}
	return nil
}

var _ = Describe("DistributedInferenceValidatorHandler", func() {
	var (
		rawObject runtime.RawExtension
		ctx       context.Context
		scheme    *runtime.Scheme
	)
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	ctx = context.Background()

	Context("Constructor", func() {
		It("should create validator handler with validators", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			mockValidator := &mockDistributedInferenceValidator{}
			allowedServiceAccounts := []string{"test-service-account"}

			handler := NewDistributedInferenceValidator(k8sClient, allowedServiceAccounts, false, mockValidator)

			Expect(handler).ToNot(BeNil())
			Expect(handler.k8sClient).To(Equal(k8sClient))
			Expect(handler.allowedServiceAccounts).To(Equal(allowedServiceAccounts))
			Expect(len(handler.validators)).To(Equal(1))
		})
	})

	Context("Create workload", func() {
		DescribeTable("should allow to create workload with valid configuration", func(assetsTest DistributedAssetsTest) {
			mockValidator := &mockDistributedInferenceValidator{}
			handler := NewDistributedInferenceValidator(assetsTest.k8sClient, []string{}, false, mockValidator)

			dumpJson, err := json.Marshal(assetsTest.currentWorkload)
			Expect(err).ToNot(HaveOccurred())

			rawObject = runtime.RawExtension{Raw: dumpJson}
			response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		},
			Entry("Distributed Inference Workload", DistributedAssetsTest{
				k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				currentWorkload: createTestDistributedInferenceWorkload(),
			}),
		)

		It("should deny create workload when validator fails", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			mockValidator := &mockDistributedInferenceValidator{
				shouldFail: true,
				errorMsg:   "validation failed",
			}
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false, mockValidator)

			workload := createTestDistributedInferenceWorkload()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())

			rawObject = runtime.RawExtension{Raw: dumpJson}
			response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("validation failed"))
		})

		It("should allow create workload for non-distributed inference kind", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

			workload := createTestDistributedInferenceWorkload()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())

			rawObject = runtime.RawExtension{Raw: dumpJson}
			response := handler.HandleCreate(ctx, rawObject, "OtherKind", testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny create workload with malformed JSON", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

			rawObject = runtime.RawExtension{Raw: []byte("invalid json")}
			response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
		})

		It("should allow create distributed inference workload when runai job exists (same name)", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

			workload := createTestDistributedInferenceWorkload()
			job := resourceUtils.CreateDumbRunaiJob(workload.Name, workload.Namespace, "", "", runaiJobV1.JobSpec{})
			Expect(k8sClient.Create(ctx, job)).To(Succeed())

			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())

			rawObject = runtime.RawExtension{Raw: dumpJson}
			response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		Context("workload is run:ai control-plane workload (create)", func() {
			It("should allow system service account to create the workload", func() {
				k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
				handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

				workload := createTestDistributedInferenceWorkload()
				workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should not allow normal account to create the workload", func() {
				k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
				handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

				workload := createTestDistributedInferenceWorkload()
				workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))
			})
		})

		Context("when allowedServiceAccount is not empty", func() {
			var (
				allowedServiceAccount = "test-namespace"
			)

			It("should allow create workload when user is in allowedServiceAccount list", func() {
				k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
				handler := NewDistributedInferenceValidator(k8sClient, []string{allowedServiceAccount}, false)

				workload := createTestDistributedInferenceWorkload()
				workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, "test-namespace:test-sa")
				Expect(response.Allowed).To(BeTrue())
			})

			It("should allow create workload when user is system user", func() {
				k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
				handler := NewDistributedInferenceValidator(k8sClient, []string{allowedServiceAccount}, false)

				workload := createTestDistributedInferenceWorkload()
				workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should deny create workload when user is not in allowedServiceAccount list", func() {
				k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
				handler := NewDistributedInferenceValidator(k8sClient, []string{allowedServiceAccount}, false)

				workload := createTestDistributedInferenceWorkload()
				workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := handler.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, testNonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))
			})
		})
	})

	Context("Update workload", func() {
		var (
			handler         *DistributedInferenceValidatorHandler
			currentWorkload *dinfw.DistributedInferenceWorkload
			k8sClient       client.Client
		)

		BeforeEach(func() {
			k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
			handler = NewDistributedInferenceValidator(k8sClient, []string{}, false)
			currentWorkload = createTestDistributedInferenceWorkload()
		})

		It("should allow update workload with no changes", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should allow update workload with change in metadata labels", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.ObjectMeta.Labels = map[string]string{"test": "test"}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should allow update workload with change in metadata annotations", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.ObjectMeta.Annotations = map[string]string{"test": "test"}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should allow update workload with change in status", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.Status.Phase = "Running"
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny update workload with change in metadata namespace", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.ObjectMeta.Namespace = "new-namespace"
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("cannot update /metadata/namespace, field is immutable"))
		})

		It("should deny update workload with change in metadata kai.scheduler annotations", func() {
			currentWorkload.Annotations = map[string]string{"kai.scheduler/whatever": "name1"}
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.Annotations["kai.scheduler/whatever"] = "name2"
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("cannot update /metadata/annotations/kai.scheduler~1whatever, field is immutable"))
		})

		It("should deny update workload with change in metadata name", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			currentWorkload.ObjectMeta.Name = testNewName
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("cannot update /metadata/name, field is immutable"))
		})

		It("should deny update workload with change in spec", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			// Change something in spec (assuming spec changes are not allowed)
			currentWorkload.Spec.Workers = &flds.IntField{Value: ptr.To(5)}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("cannot update /spec/workers/value, field is immutable"))
		})

		It("should deny update workload when validator fails", func() {
			mockValidator := &mockDistributedInferenceValidator{
				shouldFail: true,
				errorMsg:   "validation failed",
			}
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false, mockValidator)

			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("validation failed"))
		})

		It("should deny update workload with malformed old JSON", func() {
			rawObjectOld := runtime.RawExtension{Raw: []byte("invalid json")}

			dumpJson, err := json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
		})

		It("should deny update workload with malformed new JSON", func() {
			workload := currentWorkload.DeepCopy()
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			rawObjectNew := runtime.RawExtension{Raw: []byte("invalid json")}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
			Expect(response.Allowed).To(BeFalse())
		})

		It("should allow changes to /spec/replicas if mnnvl is disabled", func() {
			workload := currentWorkload.DeepCopy()
			// Set MNNVL label to None (disabled)
			workload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.None}
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			// Change replicas
			currentWorkload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.None}
			currentWorkload.Spec.Replicas = &flds.IntField{Value: ptr.To(3)}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny changes to /spec/replicas if mnnvl is enabled and elastic is disabled", func() {
			workload := currentWorkload.DeepCopy()
			// Set MNNVL label to Required (enabled)
			workload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.Required}
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			// Change replicas
			currentWorkload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.Required}
			currentWorkload.Spec.Replicas = &flds.IntField{Value: ptr.To(3)}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("cannot update /spec/replicas/value, workload cannot be scaled when GPU network acceleration (MNNVL) is enabled"))
		})

		It("should allow changes to /spec/replicas if mnnvl is enabled and elastic is enabled", func() {
			// Create handler with elasticEnabled=true
			elasticHandler := NewDistributedInferenceValidator(k8sClient, []string{}, true)

			workload := currentWorkload.DeepCopy()
			// Set MNNVL label to Required (enabled)
			workload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.Required}
			dumpJson, err := json.Marshal(workload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectOld := runtime.RawExtension{Raw: dumpJson}

			// Change replicas
			currentWorkload.Labels = map[string]string{mnnvl.FeatureLabelKey: mnnvl.Required}
			currentWorkload.Spec.Replicas = &flds.IntField{Value: ptr.To(3)}
			dumpJson, err = json.Marshal(currentWorkload)
			Expect(err).ToNot(HaveOccurred())
			rawObjectNew := runtime.RawExtension{Raw: dumpJson}

			response := elasticHandler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testSystemUser)
			Expect(response.Allowed).To(BeTrue())
		})

		Context("workload is run:ai control-plane workload (update)", func() {
			BeforeEach(func() {
				currentWorkload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
			})

			It("should allow system service account to update the workload", func() {
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should not allow normal account to update the workload", func() {
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))
			})
		})

		Context("when allowedServiceAccount is not empty", func() {
			var (
				allowedServiceAccount = "test-namespace"
			)

			BeforeEach(func() {
				handler = NewDistributedInferenceValidator(k8sClient, []string{allowedServiceAccount}, false)
				currentWorkload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
			})

			It("should allow update workload when user is in allowedServiceAccount list", func() {
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, "test-namespace:test-sa")
				Expect(response.Allowed).To(BeTrue())
			})

			It("should allow update workload when user is system user", func() {
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should deny update workload when user is not in allowedServiceAccount list", func() {
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := handler.HandleUpdate(ctx, rawObjectOld, rawObjectNew, constants.KindDistributedInferenceWorkload, testNonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))
			})
		})
	})

	Context("isAllowedFieldChangeDistributedInference", func() {
		It("should allow changes to metadata fields except name and namespace", func() {
			allowedChanges := []string{
				"/metadata/labels",
				"/metadata/annotations",
				"/metadata/creationTimestamp",
				"/metadata/resourceVersion",
			}

			for _, change := range allowedChanges {
				patch := jsonpatch.Operation{Path: change}
				Expect(isAllowedFieldChangeDistributedInference(patch)).To(BeTrue(), "Expected %s to be allowed", change)
			}
		})

		It("should deny changes to metadata name and namespace", func() {
			deniedChanges := []string{
				"/metadata/name",
				"/metadata/namespace",
			}

			for _, change := range deniedChanges {
				patch := jsonpatch.Operation{Path: change}
				Expect(isAllowedFieldChangeDistributedInference(patch)).To(BeFalse(), "Expected %s to be denied", change)
			}
		})

		It("should allow changes to status fields and to /spec/replicas", func() {
			allowedChanges := []string{
				"/status/phase",
				"/status/conditions",
				"/status/message",
				"/spec/replicas",
			}

			for _, change := range allowedChanges {
				patch := jsonpatch.Operation{Path: change}
				Expect(isAllowedFieldChangeDistributedInference(patch)).To(BeTrue(), "Expected %s to be allowed", change)
			}
		})

		It("should deny changes to spec fields", func() {
			deniedChanges := []string{
				"/spec/workers",
				"/spec/servingPorts",
				"/spec/startupPolicy",
			}

			for _, change := range deniedChanges {
				patch := jsonpatch.Operation{Path: change}
				Expect(isAllowedFieldChangeDistributedInference(patch)).To(BeFalse(), "Expected %s to be denied", change)
			}
		})
	})

	Context("InjectDecoder", func() {
		It("should return nil", func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			handler := NewDistributedInferenceValidator(k8sClient, []string{}, false)

			err := handler.InjectDecoder(nil)
			Expect(err).To(BeNil())
		})
	})
})

func createTestDistributedInferenceWorkload() *dinfw.DistributedInferenceWorkload {
	return &dinfw.DistributedInferenceWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "workload.run.ai/v1alpha1",
			Kind:       constants.KindDistributedInferenceWorkload,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDistributedWorkloadName,
			Namespace: testNamespace,
			Labels:    map[string]string{},
		},
		Spec: dp.DistributedInferenceSpec{
			DistributedInferenceFields: field_groups.DistributedInferenceFields{
				Workers:  &flds.IntField{Value: ptr.To(1)},
				Replicas: &flds.IntField{Value: ptr.To(1)},
			},
		},
		Status: dinfw.DistributedInferenceWorkloadStatus{
			DistributedInferenceStatus: common.DistributedInferenceStatus{
				Phase: "Pending",
			},
		},
	}
}
