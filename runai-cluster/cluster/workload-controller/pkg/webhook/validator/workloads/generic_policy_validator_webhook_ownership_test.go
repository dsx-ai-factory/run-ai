package workloads

import (
	"context"
	"encoding/json"
	"fmt"
	workloadspolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads/policy"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	interactivepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const testUsername = "testUserName"

type FakeGetClient struct {
	client.Client
}

func (f *FakeGetClient) Get(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	return nil
}

var _ = Describe("Policy Validator CreateMergeStackForFrontend Suite", func() {
	var (
		ctx                  context.Context
		scheme               *runtime.Scheme
		admissionFakeRequest admission.Request
	)

	BeforeEach(func() {
		ctx = context.Background()

		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		os.Setenv(utils.EnvVarControlPlanePolicies, "true")
	})

	AfterEach(func() {
		os.Unsetenv(utils.EnvVarControlPlanePolicies)
	})

	Context("Ownership Protection", func() {
		It("control plane policies - should not be allowed to make this action due to ownership restriction", func() {

			policyMeta := resource.NewRunaiObject("minMaxOk")
			policyMeta.Labels = map[string]string{
				constants.PolicyLabel: constants.BoolTrueAsString,
			}

			validPolicy := interactivepolicy.InteractivePolicy{
				ObjectMeta: policyMeta,
				Spec: interactivepolicy.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: workloadspolicy.IntFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
			}

			admissionFakeRequest = admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					RequestKind: &metav1.GroupVersionKind{},
				}}

			admissionFakeRequest.RequestKind.Kind = constants.KindInteractivePolicy
			admissionFakeRequest.Namespace = constants.NamespaceRunAi
			admissionFakeRequest.Name = "interactive-policy"
			admissionFakeRequest.UserInfo.Username = testUsername
			admissionFakeRequest.Operation = admissionv1.Create

			dumpJson, err := json.Marshal(validPolicy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal validPolicy policy template")

			admissionFakeRequest.Object.Raw = dumpJson

			newK8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			newK8sClient = &FakeGetClient{newK8sClient}
			genericPolicyValidatorWithMockGet := workloadspolicy.NewPolicyValidator(newK8sClient, workloadspolicy.InitInteractivePolicyFromReq, []string{})

			admissionResponse := genericPolicyValidatorWithMockGet.Handle(ctx, admissionFakeRequest)
			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusUnauthorized, fmt.Errorf("Username %s is not authorized to perform this action on %s", testUsername, constants.KindInteractivePolicy))))
		})
		It("old policies - should not enforce ownership restriction", func() {
			validPolicy := interactivepolicy.InteractivePolicy{
				ObjectMeta: resource.NewRunaiObject("minMaxOk"),
				Spec: interactivepolicy.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: workloadspolicy.IntFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
			}

			admissionFakeRequest = admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					RequestKind: &metav1.GroupVersionKind{},
				}}

			admissionFakeRequest.RequestKind.Kind = constants.KindInteractivePolicy
			admissionFakeRequest.Namespace = constants.NamespaceRunAi
			admissionFakeRequest.Name = "interactive-policy"
			admissionFakeRequest.UserInfo.Username = testUsername
			admissionFakeRequest.Operation = admissionv1.Create

			dumpJson, err := json.Marshal(validPolicy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal validPolicy policy template")

			admissionFakeRequest.Object.Raw = dumpJson

			newK8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			newK8sClient = &FakeGetClient{newK8sClient}
			genericPolicyValidatorWithMockGet := workloadspolicy.NewPolicyValidator(newK8sClient, workloadspolicy.InitInteractivePolicyFromReq, []string{})

			admissionResponse := genericPolicyValidatorWithMockGet.Handle(ctx, admissionFakeRequest)
			Expect(admissionResponse.Allowed).To(BeTrue())
			Expect(admissionResponse.AdmissionResponse.Allowed).To(BeTrue())
			Expect(admissionResponse.Result.Code).To(Equal(int32(200)))
		})
	})
})
