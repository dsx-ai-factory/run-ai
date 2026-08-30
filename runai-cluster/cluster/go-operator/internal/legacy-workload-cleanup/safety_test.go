package legacycleanup

import (
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	inferencecontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	jobcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
)

var _ = Describe("Legacy workload cleanup safety", func() {
	DescribeTable("classifies the atomic controller state",
		func(workload, inference, job *bool, expected controllerState, errorSubstring string) {
			config := &runaiv1.RunaiConfig{Spec: runaiv1.RunaiConfigSpec{
				WorkloadController:          &workloadcontroller.WorkloadController{Enabled: workload},
				InferenceWorkloadController: &inferencecontroller.InferenceWorkloadController{Enabled: inference},
				JobController:               &jobcontroller.JobController{Enabled: job},
			}}
			state, err := classifyControllerState(config)
			if errorSubstring != "" {
				Expect(err).To(MatchError(ContainSubstring(errorSubstring)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(state).To(Equal(expected))
		},
		Entry("all enabled", ptr.To(true), ptr.To(true), ptr.To(true), controllersEnabled, ""),
		Entry("all disabled", ptr.To(false), ptr.To(false), ptr.To(false), controllersDisabled, ""),
		Entry("mixed", ptr.To(false), ptr.To(true), ptr.To(true), controllerState(0), "not atomic"),
		Entry("missing value", nil, ptr.To(true), ptr.To(true), controllerState(0), "is missing"),
	)

	It("removes only the exact allowlisted finalizer", func() {
		remaining, changed := withoutFinalizer(
			[]string{"foreign.example/finalizer", legacyWorkloadFinalizer, "run.ai.workload/finalizer-extra"},
			legacyWorkloadFinalizer,
		)
		Expect(changed).To(BeTrue())
		Expect(remaining).To(Equal([]string{"foreign.example/finalizer", "run.ai.workload/finalizer-extra"}))
	})

	It("is idempotent when the allowlisted finalizer is absent", func() {
		finalizers := []string{"foreign.example/finalizer"}
		remaining, changed := withoutFinalizer(finalizers, legacyWorkloadFinalizer)
		Expect(changed).To(BeFalse())
		Expect(remaining).To(Equal(finalizers))
	})

	It("allows only the four ownership webhook configurations with their exact service and operations", func() {
		webhook := admissionv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipValidationWebhook},
			Webhooks: []admissionv1.ValidatingWebhook{{
				Name: "ownership.run.ai",
				ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{
					Namespace: "runai", Name: "anyworkload-controller",
				}},
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
					Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v2alpha1"}, Resources: []string{"trainingworkloads"}},
				}},
			}},
		}
		Expect(validateValidatingWebhooks([]admissionv1.ValidatingWebhookConfiguration{webhook})).To(Succeed())
	})

	DescribeTable("rejects overbroad rules in an approved ownership webhook",
		func(mutate func(*admissionv1.Rule)) {
			webhook := validWorkloadOwnershipValidator()
			mutate(&webhook.Webhooks[0].Rules[0].Rule)
			Expect(validateValidatingWebhooks([]admissionv1.ValidatingWebhookConfiguration{*webhook})).To(
				MatchError(ContainSubstring("not an exact ownership handoff rule")))
		},
		Entry("wildcard resource", func(rule *admissionv1.Rule) { rule.Resources = []string{"*"} }),
		Entry("RunaiJob", func(rule *admissionv1.Rule) { rule.Resources = []string{"runaijobs"} }),
		Entry("policy CRD", func(rule *admissionv1.Rule) { rule.Resources = []string{"trainingpolicies"} }),
		Entry("wrong API version", func(rule *admissionv1.Rule) { rule.APIVersions = []string{"v1alpha1"} }),
		Entry("extra API group", func(rule *admissionv1.Rule) { rule.APIGroups = []string{"run.ai", "apps"} }),
		Entry("workload subresource", func(rule *admissionv1.Rule) {
			rule.Resources = []string{"trainingworkloads/status"}
		}),
	)

	It("rejects a retired workload rule on an unapproved webhook", func() {
		webhook := admissionv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "legacy-workload-validation"},
			Webhooks: []admissionv1.ValidatingWebhook{{
				Name: "legacy.run.ai",
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Create},
					Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, Resources: []string{"trainingworkloads"}},
				}},
			}},
		}
		Expect(validateValidatingWebhooks([]admissionv1.ValidatingWebhookConfiguration{webhook})).To(
			MatchError(ContainSubstring("legacy-workload-validation")))
	})

	It("recognizes retired workload subresources in webhook rules", func() {
		Expect(targetsRetiredResource(admissionv1.Rule{
			APIGroups: []string{"run.ai"}, Resources: []string{"trainingworkloads/status"},
		})).To(BeTrue())
	})

	It("allows an update-only ownership mutator targeting AnyWorkload", func() {
		webhook := admissionv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipMutationWebhook},
			Webhooks: []admissionv1.MutatingWebhook{{
				Name: "ownership.run.ai",
				ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{
					Namespace: "runai", Name: "anyworkload-controller",
				}},
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Update},
					Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v2alpha1"}, Resources: []string{"trainingworkloads"}},
				}},
			}},
		}
		Expect(validateMutatingWebhooks([]admissionv1.MutatingWebhookConfiguration{webhook})).To(Succeed())
	})

	It("rejects create on a retained ownership mutator", func() {
		webhook := admissionv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipMutationWebhook},
			Webhooks: []admissionv1.MutatingWebhook{{
				ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{
					Namespace: "runai", Name: "anyworkload-controller",
				}},
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update},
					Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, Resources: []string{"trainingworkloads"}},
				}},
			}},
		}
		Expect(validateMutatingWebhooks([]admissionv1.MutatingWebhookConfiguration{webhook})).To(
			MatchError(ContainSubstring("operations")))
	})

	It("rejects duplicate operations that omit a required ownership operation", func() {
		Expect(sameOperations(
			[]admissionv1.OperationType{admissionv1.Update, admissionv1.Update},
			[]admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
		)).To(BeFalse())
	})

	It("requires both workload ownership handoff webhooks when workload APIs are served", func() {
		served := []resourceDescriptor{legacyWorkloadResources[0]}
		Expect(validateOwnershipHandoff(served, nil, nil)).To(
			MatchError(ContainSubstring(workloadOwnershipValidationWebhook)))
	})

	It("does not require inference ownership handoff when inference APIs are absent", func() {
		served := []resourceDescriptor{legacyWorkloadResources[0]}
		validating := []admissionv1.ValidatingWebhookConfiguration{{
			ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipValidationWebhook},
			Webhooks: []admissionv1.ValidatingWebhook{{
				ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{Namespace: "runai", Name: "anyworkload-controller"}},
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
					Rule: admissionv1.Rule{
						APIGroups:   []string{"run.ai"},
						APIVersions: []string{"v2alpha1"},
						Resources:   []string{"trainingworkloads"},
					},
				}},
			}},
		}}
		mutating := []admissionv1.MutatingWebhookConfiguration{{
			ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipMutationWebhook},
			Webhooks: []admissionv1.MutatingWebhook{{
				ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{Namespace: "runai", Name: "anyworkload-controller"}},
				Rules: []admissionv1.RuleWithOperations{{
					Operations: []admissionv1.OperationType{admissionv1.Update},
					Rule: admissionv1.Rule{
						APIGroups:   []string{"run.ai"},
						APIVersions: []string{"v2alpha1"},
						Resources:   []string{"trainingworkloads"},
					},
				}},
			}},
		}}
		Expect(validateOwnershipHandoff(served, validating, mutating)).To(Succeed())
	})

	It("requires ownership handoff coverage for every served workload resource", func() {
		served := []resourceDescriptor{legacyWorkloadResources[0], legacyWorkloadResources[1]}
		Expect(validateOwnershipHandoff(served,
			[]admissionv1.ValidatingWebhookConfiguration{*validWorkloadOwnershipValidator()},
			[]admissionv1.MutatingWebhookConfiguration{*validWorkloadOwnershipMutator()},
		)).To(MatchError(ContainSubstring("interactiveworkloads")))
	})

	It("requires ownership handoff coverage for the exact served API version", func() {
		validator := validWorkloadOwnershipValidator()
		validator.Webhooks[0].Rules[0].APIVersions = []string{"v1alpha1"}
		mutator := validWorkloadOwnershipMutator()
		mutator.Webhooks[0].Rules[0].APIVersions = []string{"v1alpha1"}
		Expect(validateOwnershipHandoff([]resourceDescriptor{legacyWorkloadResources[0]},
			[]admissionv1.ValidatingWebhookConfiguration{*validator},
			[]admissionv1.MutatingWebhookConfiguration{*mutator},
		)).To(MatchError(ContainSubstring("v2alpha1")))
	})
})

func newObject(apiVersion, kind, namespace, name string, uid types.UID) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetAPIVersion(apiVersion)
	object.SetKind(kind)
	object.SetNamespace(namespace)
	object.SetName(name)
	object.SetUID(uid)
	return object
}
