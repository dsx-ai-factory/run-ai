package legacycleanup

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Cluster safety checks", func() {
	It("does not require workloads-auth for legacy controller cleanup", func() {
		k8sClient := safetyClient()
		Expect(validateLegacyWebhookState(context.Background(), k8sClient, nil, false)).To(Succeed())
	})

	It("does not require ownership handoff webhooks when ownership protection is disabled", func() {
		k8sClient := safetyClient()
		Expect(validateLegacyWebhookState(
			context.Background(), k8sClient, []resourceDescriptor{legacyWorkloadResources[0]}, false,
		)).To(Succeed())
	})

	It("rejects a stale ownership webhook when ownership protection is disabled", func() {
		k8sClient := safetyClient(validWorkloadOwnershipValidator())
		Expect(validateLegacyWebhookState(
			context.Background(), k8sClient, []resourceDescriptor{legacyWorkloadResources[0]}, false,
		)).To(MatchError(ContainSubstring("ownership protection is disabled")))
	})

	It("requires ownership handoff webhooks when ownership protection is enabled", func() {
		k8sClient := safetyClient()
		Expect(validateLegacyWebhookState(
			context.Background(), k8sClient, []resourceDescriptor{legacyWorkloadResources[0]}, true,
		)).To(MatchError(ContainSubstring(workloadOwnershipValidationWebhook)))
	})

	It("accepts exact ownership handoff webhooks when ownership protection is enabled", func() {
		k8sClient := safetyClient(
			validWorkloadOwnershipValidator(), validWorkloadOwnershipMutator(),
		)
		Expect(validateLegacyWebhookState(
			context.Background(), k8sClient, []resourceDescriptor{legacyWorkloadResources[0]}, true,
		)).To(Succeed())
	})

	It("rejects a live retired webhook rule", func() {
		configuration := &admissionv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "legacy-workload-mutation"},
			Webhooks: []admissionv1.MutatingWebhook{{Rules: []admissionv1.RuleWithOperations{{
				Rule: admissionv1.Rule{APIGroups: []string{"run.ai"}, Resources: []string{"distributedworkloads"}},
			}}}},
		}
		k8sClient := safetyClient(configuration)
		Expect(validateLegacyWebhookState(context.Background(), k8sClient, nil, false)).To(
			MatchError(ContainSubstring("legacy-workload-mutation")))
	})
})

func safetyClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(admissionv1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}
