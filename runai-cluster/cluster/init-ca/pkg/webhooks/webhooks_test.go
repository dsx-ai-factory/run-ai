package webhooks

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gotest.tools/assert"
	registrations "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/certificates"
)

func TestFilterOwnershipWebhooksForServedLegacyResources(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.Discovery().(*fakediscovery.FakeDiscovery).Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "run.ai/v2alpha1",
			APIResources: []metav1.APIResource{{Name: "trainingworkloads"}},
		},
	}
	ownershipValidator := &registrations.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipValidationName},
		Webhooks: []registrations.ValidatingWebhook{{Rules: []registrations.RuleWithOperations{{
			Rule: registrations.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v2alpha1"}, Resources: []string{
				"trainingworkloads", "interactiveworkloads", "distributedworkloads",
			}},
		}}}},
	}
	ownershipMutator := &registrations.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: inferenceOwnershipMutationName},
		Webhooks: []registrations.MutatingWebhook{{Rules: []registrations.RuleWithOperations{{
			Rule: registrations.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v1alpha1"}, Resources: []string{
				"distributedinferenceworkloads",
			}},
		}}}},
	}

	validators, mutators, err := filterOwnershipWebhooksByDiscovery(
		client.Discovery(), []*registrations.ValidatingWebhookConfiguration{ownershipValidator},
		[]*registrations.MutatingWebhookConfiguration{ownershipMutator},
	)

	if err != nil {
		t.Fatalf("filter ownership webhooks: %v", err)
	}
	if len(validators) != 1 {
		t.Fatalf("validators length = %d, want 1", len(validators))
	}
	resources := validators[0].Webhooks[0].Rules[0].Resources
	if len(resources) != 1 || resources[0] != "trainingworkloads" {
		t.Fatalf("validator resources = %v, want [trainingworkloads]", resources)
	}
	if len(mutators) != 0 {
		t.Fatalf("mutators length = %d, want 0 for absent distributed inference CRD", len(mutators))
	}
}

func TestUpdateWebhooksAppliesDesiredAnnotations(t *testing.T) {
	validatorName := "existing-validator"
	mutatorName := "existing-mutator"
	client := fake.NewSimpleClientset(
		&registrations.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: validatorName, Labels: map[string]string{"app": "runai", deployVersionLabelName: "old"},
				Annotations: map[string]string{"existing": "preserved"},
			},
			Webhooks: []registrations.ValidatingWebhook{{Name: validatorName}},
		},
		&registrations.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: mutatorName, Labels: map[string]string{"app": "runai", deployVersionLabelName: "old"},
				Annotations: map[string]string{"existing": "preserved"},
			},
			Webhooks: []registrations.MutatingWebhook{{Name: mutatorName}},
		},
	)
	desiredAnnotations := map[string]string{"service.beta.openshift.io/inject-cabundle": "true"}
	desiredValidator := &registrations.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: validatorName, Labels: map[string]string{deployVersionLabelName: deployVersionExample},
			Annotations: desiredAnnotations,
		},
		Webhooks: []registrations.ValidatingWebhook{{
			Name: validatorName, ClientConfig: registrations.WebhookClientConfig{CABundle: cabundleexample},
		}},
	}
	desiredMutator := &registrations.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mutatorName, Labels: map[string]string{deployVersionLabelName: deployVersionExample},
			Annotations: desiredAnnotations,
		},
		Webhooks: []registrations.MutatingWebhook{{
			Name: mutatorName, ClientConfig: registrations.WebhookClientConfig{CABundle: cabundleexample},
		}},
	}

	if err := createOrUpdateWebhooks(
		context.Background(), client,
		[]*registrations.ValidatingWebhookConfiguration{desiredValidator},
		[]*registrations.MutatingWebhookConfiguration{desiredMutator},
	); err != nil {
		t.Fatalf("update webhooks: %v", err)
	}
	validator, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(
		context.Background(), validatorName, metav1.GetOptions{},
	)
	if err != nil {
		t.Fatalf("get validator: %v", err)
	}
	mutator, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(
		context.Background(), mutatorName, metav1.GetOptions{},
	)
	if err != nil {
		t.Fatalf("get mutator: %v", err)
	}
	if validator.Annotations["service.beta.openshift.io/inject-cabundle"] != "true" {
		t.Errorf("validator annotations = %v, want OpenShift service CA injection", validator.Annotations)
	}
	if mutator.Annotations["service.beta.openshift.io/inject-cabundle"] != "true" {
		t.Errorf("mutator annotations = %v, want OpenShift service CA injection", mutator.Annotations)
	}
	if validator.Annotations["existing"] != "preserved" || mutator.Annotations["existing"] != "preserved" {
		t.Errorf("existing annotations were not preserved: validator=%v mutator=%v", validator.Annotations, mutator.Annotations)
	}
}

const (
	deployVersionExample = "abcdefg"
)

var (
	cabundleexample  = []byte("cabundleexample")
	cabundleexample2 = []byte("cabundleexample2")
	cabundleexample3 = []byte("cabundleexample3")
	cabundleexample4 = []byte("cabundleexample4")
)

func TestWebhooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhooks Suite")
}

var _ = Describe("getWebhooksFromFile", func() {
	Context("getting admission webhooks for cert manager", func() {
		It("should be able to read the webhooks data", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil, /* application controller */
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile("./test_input.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(validators)).To(Equal(1))
			Expect(len(mutators)).To(Equal(1))
			Expect(validators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(mutators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(validators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
			Expect(mutators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
		})
	})
	Context("getting admission webhooks for k8s certificates", func() {
		It("should be able to read the webhooks data", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile("./test_input.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(validators)).To(Equal(1))
			Expect(len(mutators)).To(Equal(1))
			Expect(validators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(mutators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(validators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
			Expect(mutators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
			Expect(len(validators[0].Annotations)).To(Equal(0))
			Expect(len(mutators[0].Annotations)).To(Equal(0))
		})
	})
	Context("webhooks for workload controller", func() {
		It("skips stale legacy controller webhooks when their serving certificate is disabled", func() {
			webhookCertificates := &WebhookCertificates{
				AdmissionControllerCertificateDetails: certificates.NewCertificateDetails(cabundleexample),
			}

			validators, mutators, err := getWebhooksFromFile(
				"./disabled_legacy_webhooks_input.yaml", deployVersionExample, webhookCertificates,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(validators).To(BeEmpty())
			Expect(mutators).To(BeEmpty())
		})

		It("should use other certificate details for inference workload controller", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				certificates.NewCertificateDetails(cabundleexample3),
				certificates.NewCertificateDetails(cabundleexample2),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile("./test_input_with_workload.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(validators)).To(Equal(3)) // 2 for workload, 1 for policy
			Expect(len(mutators)).To(Equal(1))   // 1 for workload
			for _, validator := range validators {
				Expect(validator.Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
				if strings.Contains(validator.Name, "inference-workload") || strings.Contains(validator.Name, "inference-policy") {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample3))
				} else if strings.Contains(validator.Name, "workload") || strings.Contains(validator.Name, "policy") {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample2))
				} else {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
				}
			}

			Expect(mutators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(mutators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
		})
		It("should use other certificate details for workload controller", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				certificates.NewCertificateDetails(cabundleexample3),
				certificates.NewCertificateDetails(cabundleexample2),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile("./test_input_with_workload.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(validators)).To(Equal(3)) // 2 for workload, 1 for policy
			Expect(len(mutators)).To(Equal(1))   // 1 for workload
			for _, validator := range validators {
				Expect(validator.Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
				if strings.Contains(validator.Name, "workload") || strings.Contains(validator.Name, "policy") {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample2))
				} else {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
				}
			}

			Expect(mutators[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			Expect(mutators[0].Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
		})
	})
	Context("webhooks for pod group assigner and accessrule controller", func() {
		It("should use other certificate details for pod group assigner and accessrule controller", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				certificates.NewCertificateDetails(cabundleexample2),
				certificates.NewCertificateDetails(cabundleexample2),
				certificates.NewCertificateDetails(cabundleexample3),
				certificates.NewCertificateDetails(cabundleexample4),
				certificates.NewCertificateDetails(cabundleexample4),
				nil,
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile(
				"./test_input_with_pod_group_assigner_accessrule_controller.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(validators)).To(Equal(4))
			Expect(len(mutators)).To(Equal(2))
			for _, validator := range validators {
				Expect(validator.Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
				if strings.Contains(validator.Name, "workload") || strings.Contains(validator.Name, "policy") {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample2))
				} else if validator.Name == "runai-accessrule-validation" {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample4))
				} else {
					Expect(validator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
				}
			}

			for _, mutator := range mutators {
				Expect(mutator.Labels[deployVersionLabelName]).To(Equal(deployVersionExample))

				if strings.Contains(mutator.Name, "workload") || strings.Contains(mutator.Name, "policy") {
					Expect(mutator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample))
				} else if mutator.Name == "runai-pod-group-mutation" {
					Expect(mutator.Webhooks[0].ClientConfig.CABundle).To(Equal(cabundleexample3))
				}
			}
		})
	})
})

func testApplyWebhooks(kubeClient *fake.Clientset, actions int) {
	webhookCertificates := &WebhookCertificates{
		certificates.NewCertificateDetails(cabundleexample),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	}
	err := ApplyWebhooks(context.TODO(), kubeClient, "./test_input.yaml", deployVersionExample,
		webhookCertificates)
	Expect(err).To(BeNil())
	Expect(len(kubeClient.Actions())).To(Equal(actions))
	mwcl, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(mwcl.Items)).To(Equal(1))
	Expect(mwcl.Items[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
	vwcl, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(vwcl.Items)).To(Equal(1))
	Expect(vwcl.Items[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
}

var _ = Describe("ApplyWebhooks", func() {
	var (
		kubeClient *fake.Clientset
	)
	BeforeEach(func() {
		kubeClient = fake.NewSimpleClientset()
	})
	Context("First run / no previous webhooks", func() {
		It("should be able to create new webhooks", func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			}
			err := ApplyWebhooks(context.TODO(), kubeClient, "./test_input.yaml", deployVersionExample,
				webhookCertificates)
			Expect(err).To(BeNil())
			Expect(len(kubeClient.Actions())).To(Equal(6))
			mwcl, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
			Expect(err).To(BeNil())
			Expect(len(mwcl.Items)).To(Equal(1))
			Expect(mwcl.Items[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
			vwcl, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
			Expect(err).To(BeNil())
			Expect(len(vwcl.Items)).To(Equal(1))
			Expect(vwcl.Items[0].Labels[deployVersionLabelName]).To(Equal(deployVersionExample))
		})
	})
	Context("With previous webhooks", func() {
		BeforeEach(func() {
			webhookCertificates := &WebhookCertificates{
				certificates.NewCertificateDetails(cabundleexample),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			}
			validators, mutators, err := getWebhooksFromFile("./test_input.yaml", "otherDeployVersion",
				webhookCertificates)
			Expect(err).To(BeNil())
			for _, validator := range validators {
				_, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.TODO(), validator, metav1.CreateOptions{})
				Expect(err).To(BeNil())
			}
			for _, mutator := range mutators {
				_, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), mutator, metav1.CreateOptions{})
				Expect(err).To(BeNil())
			}
		})
		It("should be able to update existing webhooks", func() {
			testApplyWebhooks(kubeClient, 8)
		})
		It("should remove old webhooks", func() {
			_, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.TODO(), &registrations.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "old-validator",
					Labels: map[string]string{
						"app":                  "runai",
						deployVersionLabelName: "oldDeployVersion",
					},
				},
				Webhooks: []registrations.ValidatingWebhook{
					{Name: "old-validator"},
				},
			}, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			testApplyWebhooks(kubeClient, 10)
		})
	})
})

type ChoiceTestDetails struct {
	webhookName    string
	expectedChoice []byte
}

func TestCertificateChoice(t *testing.T) {
	workloadCert := certificates.CertificateDetails{
		CaBundle: []byte("workloadCaBundle"),
	}
	inferenceWorkloadCert := certificates.CertificateDetails{
		CaBundle: []byte("workloadCaBundle"),
	}
	admissionCert := certificates.CertificateDetails{
		CaBundle: []byte("admissionCaBundle"),
	}
	podGroupAssignerCert := certificates.CertificateDetails{
		CaBundle: []byte("podGroupAssignerCaBundle"),
	}
	accessruleControllerCert := certificates.CertificateDetails{
		CaBundle: []byte("accessruleControllerCaBundle"),
	}
	agentCert := certificates.CertificateDetails{
		CaBundle: []byte("agentCaBundle"),
	}
	anyworkloadControllerCert := certificates.CertificateDetails{
		CaBundle: []byte("anyworkloadControllerCaBundle"),
	}
	applicationControllerCert := certificates.CertificateDetails{
		CaBundle: []byte("applicationControllerCaBundle"),
	}

	tests := []ChoiceTestDetails{
		{
			webhookName:    "runai-workload-ownership-protection-validation",
			expectedChoice: anyworkloadControllerCert.CaBundle,
		},
		{
			webhookName:    "runai-inference-workload-ownership-mutation",
			expectedChoice: anyworkloadControllerCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-interactive-policy-validation",
			expectedChoice: workloadCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-training-policy-validation",
			expectedChoice: workloadCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-interactive-workload-validation",
			expectedChoice: workloadCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-training-workload-validation",
			expectedChoice: workloadCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-inference-policy-validation",
			expectedChoice: inferenceWorkloadCert.CaBundle,
		},
		{
			webhookName:    "run-ai-v1-inference-workload-validation",
			expectedChoice: inferenceWorkloadCert.CaBundle,
		},
		{
			webhookName:    "something-totally-different-validation",
			expectedChoice: admissionCert.CaBundle,
		},
		{
			webhookName:    "runai-pod-group-mutation",
			expectedChoice: podGroupAssignerCert.CaBundle,
		},
		{
			webhookName:    "runai-accessrule-validation",
			expectedChoice: accessruleControllerCert.CaBundle,
		},
		{
			webhookName:    "runai-project-validation",
			expectedChoice: agentCert.CaBundle,
		},
		{
			webhookName:    "runai-secret-pvc-validation",
			expectedChoice: agentCert.CaBundle,
		},
		{
			webhookName:    "application-validation",
			expectedChoice: applicationControllerCert.CaBundle,
		},
		{
			webhookName:    "runai-restartrequest-validation",
			expectedChoice: anyworkloadControllerCert.CaBundle,
		},
	}

	for _, curTest := range tests {
		webhookCertificates := &WebhookCertificates{
			AdmissionControllerCertificateDetails:         &admissionCert,
			InferenceWorkloadControllerCertificateDetails: &inferenceWorkloadCert,
			WorkloadControllerCertificateDetails:          &workloadCert,
			PodGroupAssignerCertificateDetails:            &podGroupAssignerCert,
			AccessruleControllerCertificateDetails:        &accessruleControllerCert,
			AnyworkloadControllerCertificateDetails:       &anyworkloadControllerCert,
			ApplicationControllerCertificateDetails:       &applicationControllerCert,
			AdviceRestartControllerCertificateDetails:     nil,
			AgentCertificateDetails:                       &agentCert,
		}
		actualChoice := chooseCertificateDetails(curTest.webhookName, webhookCertificates)
		assert.Equal(t, string(actualChoice.CaBundle), string(curTest.expectedChoice), curTest)
	}
}
