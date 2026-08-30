package controller

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	kartautils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/karta"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/webhook/mutator"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/webhook/validator"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	timeout          = time.Second * 10
	interval         = time.Millisecond * 250
	cacheSyncTimeout = time.Minute
)

// marshalRawExtension creates a runtime.RawExtension from an object
// by marshaling it to JSON. Panics on marshal error (acceptable in tests).
func marshalRawExtension(obj client.Object) runtime.RawExtension {
	raw, err := json.Marshal(obj)
	if err != nil {
		panic("failed to marshal object in test: " + err.Error())
	}
	return runtime.RawExtension{Raw: raw}
}

// newTestPod creates a basic test Pod with the given name and namespace.
func newTestPod(name, namespace string) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// newPodRawExtension creates a basic test Pod and marshals it to a RawExtension.
func newPodRawExtension(name, namespace string) runtime.RawExtension {
	return marshalRawExtension(newTestPod(name, namespace))
}

// newJobRawExtension creates a basic test Job and marshals it to a RawExtension.
func newJobRawExtension(name, namespace string) runtime.RawExtension {
	job := &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "jobset.x-k8s.io/v1alpha2",
			Kind:       "JobSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return marshalRawExtension(job)
}

func newTestExternalAccessWithIngress(ingressName, host, serviceName string, port int32, path string) *dromosv1alpha1.NetworkingExternalAccessConfig {
	pathType := networkingv1.PathTypePrefix
	return &dromosv1alpha1.NetworkingExternalAccessConfig{
		Ingress: &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: ingressName},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: host,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     path,
										PathType: &pathType,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: serviceName,
												Port: networkingv1.ServiceBackendPort{Number: port},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// createResource creates a resource, ignoring AlreadyExists errors.
func createResource[T client.Object](ctx context.Context, k8sClient client.Client, obj T) T {
	Eventually(func() error {
		err := k8sClient.Create(ctx, obj)

		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}, timeout, interval).Should(Succeed())
	Eventually(func() error {
		return k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	}, timeout, interval).Should(Succeed())
	return obj
}

// createWorkloadRunner creates a WorkloadRunner with the given name and namespace.
func createWorkloadRunner(ctx context.Context, k8sClient client.Client, name, namespace string, clientObject client.Object, instructions *dromosv1alpha1.Instructions) *dromosv1alpha1.WorkloadRunner {
	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template:     marshalRawExtension(clientObject),
			Instructions: instructions,
		},
	}
	return createResource(ctx, k8sClient, runner)
}

// startManager runs the manager, cancelling managerCtx if it fails to start so that
// cached reads fail the suite instead of blocking forever.
func startManager(k8sManager ctrl.Manager, managerCtx context.Context, managerCancel context.CancelFunc) {
	go func() {
		defer GinkgoRecover()
		err := k8sManager.Start(managerCtx)
		if err != nil {
			managerCancel()
		}
		Expect(err).NotTo(HaveOccurred())
	}()
}

func setupWebhooks(fw *framework.Framework, k8sClient client.Client, managerCtx context.Context, k8sManager ctrl.Manager) {
	webhookServer := k8sManager.GetWebhookServer()

	validatorHandler := validator.NewWorkloadRunnerValidatorHandler(fw)
	webhookServer.Register("/validate-workloadrunner", &webhook.Admission{Handler: validatorHandler})

	mutatorHandler := mutator.NewWorkloadRunnerMutatorHandler(fw)
	webhookServer.Register("/mutate-workloadrunner", &webhook.Admission{Handler: mutatorHandler})
}

func deleteAllForType[T client.ObjectList](ctx context.Context, k8sClient client.Client, objs T) {
	Eventually(func() error {
		return k8sClient.List(ctx, objs)
	}, timeout, interval).Should(Succeed())
	items, err := meta.ExtractList(objs)
	Expect(err).NotTo(HaveOccurred())
	for _, obj := range items {
		clientObj, ok := obj.(client.Object)
		Expect(ok).To(BeTrue())
		_ = k8sClient.Delete(ctx, clientObj)
	}
}

func setupWebhookConfigurations(absCertDir string, managerCtx context.Context, k8sClient client.Client, webhookPort int, validatingWebhookName string, mutatingWebhookName string) {
	// Read CA cert
	caCert, err := os.ReadFile(filepath.Join(absCertDir, "tls.crt"))
	Expect(err).NotTo(HaveOccurred())

	// Create ValidatingWebhookConfiguration
	validatingWebhook := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: validatingWebhookName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: "validate-workloadrunner.run.ai",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(webhookUrlValidate),
					CABundle: caCert,
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"run.ai"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"workloadrunners"},
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
				FailurePolicy:           ptr.To(admissionregistrationv1.Fail),
			},
		},
	}
	Expect(k8sClient.Create(managerCtx, validatingWebhook)).To(Succeed())

	mutatingWebhook := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mutatingWebhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "mutate-workloadrunner.run.ai",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(webhookUrlMutate),
					CABundle: caCert,
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"run.ai"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"workloadrunners"},
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
				FailurePolicy:           ptr.To(admissionregistrationv1.Fail),
			},
		},
	}
	Expect(k8sClient.Create(managerCtx, mutatingWebhook)).To(Succeed())
}

func deleteAllJobSets(ctx context.Context, k8sClient client.Client) {
	deleteAllForType(ctx, k8sClient, &jobsetv1alpha2.JobSetList{})
}

func deleteAllKartas(ctx context.Context, k8sClient client.Client) {
	deleteAllForType(ctx, k8sClient, &kartav1alpha1.KartaList{})
}

func deleteAllWorkloadRunners(ctx context.Context, k8sClient client.Client) {
	deleteAllForType(ctx, k8sClient, &dromosv1alpha1.WorkloadRunnerList{})
}

func deleteAllTestResources(ctx context.Context, k8sClient client.Client) {
	deleteAllJobSets(ctx, k8sClient)
	deleteAllWorkloadRunners(ctx, k8sClient)
	deleteAllKartas(ctx, k8sClient)
}

func waitForWebhookToBeClosed(webhookUrl string) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	Eventually(func() bool {
		resp, err := httpClient.Get(webhookUrlValidate)
		if err != nil || resp.StatusCode != http.StatusServiceUnavailable {
			return true
		}
		defer resp.Body.Close()
		return false
	}, timeout, interval).Should(BeTrue())
}

// makeCondition creates a condition map for JobSet status
func makeCondition(conditionType, status string) map[string]interface{} {
	return map[string]interface{}{
		"type":               conditionType,
		"status":             status,
		"lastTransitionTime": metav1.Now().Format(time.RFC3339),
		"reason":             conditionType,
		"message":            "",
	}
}

// updateJobSetStatus updates the status.conditions of a JobSet
func updateJobSetStatus(ctx context.Context, k8sClient client.Client, name, namespace string, conditions []interface{}) {
	jobset := &unstructured.Unstructured{}
	jobset.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "jobset.x-k8s.io",
		Version: "v1alpha2",
		Kind:    "JobSet",
	})

	Eventually(func() error {
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, jobset)
		if err != nil {
			return err
		}

		_ = unstructured.SetNestedSlice(jobset.Object, conditions, "status", "conditions")
		return k8sClient.Status().Update(ctx, jobset)
	}, timeout, interval).Should(Succeed())
}

func waitForSetupToComplete(k8sClient client.Client, managerCtx context.Context, _ int) {
	// Bounded: List blocks until the cache syncs, and Eventually cannot interrupt it.
	cacheCtx, cancelCacheCtx := context.WithTimeout(managerCtx, cacheSyncTimeout)
	defer cancelCacheCtx()
	Eventually(func() error {
		return k8sClient.List(cacheCtx, &dromosv1alpha1.WorkloadRunnerList{})
	}, cacheSyncTimeout, interval).Should(Succeed())

	// Wait for webhook server to be ready by checking if it's listening
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	Eventually(func() error {
		resp, err := httpClient.Get(webhookUrlValidate)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		return nil
	}, timeout, interval).Should(Succeed())
}

func updateResourceStatus[T client.Object](ctx context.Context, k8sClient client.Client, obj T) {
	Eventually(func() error {
		return k8sClient.Status().Update(ctx, obj)
	}, timeout, interval).Should(Succeed())
}

func createKartaWithStatus(ctx context.Context, k8sClient client.Client, karta *kartav1alpha1.Karta) *kartav1alpha1.Karta {
	kartaCopy := karta.DeepCopy()
	karta = createResource(ctx, k8sClient, karta)
	karta.Status = kartaCopy.Status
	updateResourceStatus(ctx, k8sClient, karta)
	Eventually(func() bool {
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      karta.Name,
			Namespace: karta.Namespace,
		}, karta)
		return err == nil && karta.Status.Conditions != nil
	}, timeout, interval).Should(BeTrue())
	return karta
}

func newKarta(karta *kartav1alpha1.Karta, ready bool, crdExists bool) *kartav1alpha1.Karta {
	now := metav1.Now()

	readyStatus := metav1.ConditionFalse
	if ready && crdExists {
		readyStatus = metav1.ConditionTrue
	}

	conditions := []metav1.Condition{
		{
			Type:               string(kartautils.Ready),
			Status:             readyStatus,
			LastTransitionTime: now,
			Reason:             "Ready",
			Message:            "Karta is ready",
		},
	}

	karta = karta.DeepCopy()
	karta.Status.Conditions = conditions
	return karta
}

func newKartaWithValidStatus(karta *kartav1alpha1.Karta) *kartav1alpha1.Karta {
	return newKarta(karta, true, true)
}
