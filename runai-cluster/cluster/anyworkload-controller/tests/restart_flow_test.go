/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/plugins/restartrules"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	dromoscontroller "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/nodeaffinity"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

// foreignWorkloadGVK is a workload kind the restart feature does not own.
var foreignWorkloadGVK = metav1.GroupVersionKind{
	Group:   dromosv1alpha1.GroupVersion.Group,
	Version: dromosv1alpha1.GroupVersion.Version,
	Kind:    "OtherWorkload",
}

const (
	// restartFlowChildDeletionTimeout time waiting for deletion
	restartFlowChildDeletionTimeout = 5 * time.Second

	// childHoldFinalizer keeps a child JobSet terminating so a restart cannot complete.
	childHoldFinalizer = "test.run.ai/restart-flow-hold"
)

// newValidRestartRequest builds a valid RestartRequest from a workload
func newValidRestartRequest(ctx context.Context, c client.Client, name string,
	workload *dromosv1alpha1.WorkloadRunner, overrides runv1alpha1.SpecOverrides) *runv1alpha1.RestartRequest {
	GinkgoHelper()
	current := &dromosv1alpha1.WorkloadRunner{}
	Expect(c.Get(ctx, client.ObjectKeyFromObject(workload), current)).To(Succeed())

	rr := newRestartRequest(name, workload.Namespace, workload.Name, workloadRunnerGVK)
	rr.Labels = map[string]string{commonconstants.RunaiWorkloadNameLabel: workload.Name}
	rr.Spec.WorkloadRef.Generation = current.Generation
	rr.Spec.SpecOverrides = overrides
	return rr
}

// nodeAffinityOverride is the standard advice-shaped spec override (the only admitted path).
func nodeAffinityOverride() runv1alpha1.SpecOverrides {
	value := `{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[` +
		`{"matchExpressions":[{"key":"kubernetes.io/hostname","operator":"NotIn","values":["node-a"]}]}]}}}`
	return runv1alpha1.SpecOverrides{{
		Op:    runv1alpha1.JSONPatchOperationAdd,
		Path:  "/spec/instructions/nodeAffinity",
		Value: &runtime.RawExtension{Raw: []byte(value)},
	}}
}

// holdChild pins the child JobSet with a finalizer so its deletion freeze
func holdChild(ctx context.Context, c client.Client, name, namespace string) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		child := &jobsetv1alpha2.JobSet{}
		g.Expect(c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, child)).To(Succeed())
		if !controllerutil.ContainsFinalizer(child, childHoldFinalizer) {
			controllerutil.AddFinalizer(child, childHoldFinalizer)
			g.Expect(c.Update(ctx, child)).To(Succeed())
		}
	}, timeout, interval).Should(Succeed())
}

// releaseChild release child from freeze
func releaseChild(ctx context.Context, c client.Client, name, namespace string) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		child := &jobsetv1alpha2.JobSet{}
		err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, child)
		if k8sErrors.IsNotFound(err) {
			return
		}
		g.Expect(err).NotTo(HaveOccurred())
		if controllerutil.ContainsFinalizer(child, childHoldFinalizer) {
			controllerutil.RemoveFinalizer(child, childHoldFinalizer)
			g.Expect(c.Update(ctx, child)).To(Succeed())
		}
	}, timeout, interval).Should(Succeed())
}

// rrConditionMessages concatenates all condition messages for substring assertions.
func rrConditionMessages(rr *runv1alpha1.RestartRequest) string {
	var sb strings.Builder
	for _, c := range rr.Status.Conditions {
		sb.WriteString(c.Message)
		sb.WriteString("\n")
	}
	return sb.String()
}

// impersonatedClient builds a direct client acting as the given user (RBAC granted in BeforeAll).
func impersonatedClient(username string) client.Client {
	GinkgoHelper()
	impersonatedCfg := rest.CopyConfig(cfg)
	impersonatedCfg.Impersonate = rest.ImpersonationConfig{UserName: username}
	c, err := client.New(impersonatedCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	return c
}

// setupRestartRequestWebhookConfiguration registers the RestartRequest validating webhook
func setupRestartRequestWebhookConfiguration(ctx context.Context, c client.Client, certDir,
	url string) *admissionregistrationv1.ValidatingWebhookConfiguration {
	GinkgoHelper()
	caCert, err := os.ReadFile(filepath.Join(certDir, "tls.crt"))
	Expect(err).NotTo(HaveOccurred())

	vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "runai-restartrequest-validation"},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: "validate-restartrequest.run.ai",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(url),
					CABundle: caCert,
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create, admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"run.ai"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"restartrequests"},
							Scope:       ptr.To(admissionregistrationv1.NamespacedScope),
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
				FailurePolicy:           ptr.To(admissionregistrationv1.Fail),
			},
		},
	}
	Expect(c.Create(ctx, vwc)).To(Succeed())
	return vwc
}

var _ = Describe("Restart Flow Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager    ctrl.Manager
		k8sClient     client.Client
		managerCtx    context.Context
		managerCancel context.CancelFunc

		restartNS = "runai-restart-flow-test"

		controllerName = "restart-flow-wlr-controller"

		// tamperClient acts as a regular user, saClient as a cluster service account. saUser is
		// allowed by the system:serviceaccount: prefix rule, not the allowlist; tamperUser must
		// not be a substring of the envtest admin username or the tamper-denial specs go blind.
		tamperClient client.Client
		saClient     client.Client
		tamperUser   = "restart-flow-tamper-user"
		saUser       = "system:serviceaccount:runai-restart-flow:restart-sa"

		rrWebhookConfig  *admissionregistrationv1.ValidatingWebhookConfiguration
		impersonationCRB *rbacv1.ClusterRoleBinding
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())
		k8sClient = k8sManager.GetClient()

		// Since allowlisted identities to write spec.managedBy discover that identity and allowlist it
		clientset, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		ssr, err := clientset.AuthenticationV1().SelfSubjectReviews().
			Create(managerCtx, &authenticationv1.SelfSubjectReview{}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		adminUser := ssr.Status.UserInfo.Username

		// create restart components
		Expect(restart.NewSetup(10, restartFlowChildDeletionTimeout)(managerCtx, k8sManager)).To(Succeed())

		fw := framework.NewFramework(framework.NewRegistry(
			restartrules.NewPlugin([]string{adminUser}),
			nodeaffinity.NewPlugin(k8sClient),
		), nil)
		//nolint:staticcheck // SA1019: matches the other integration tests' recorder wiring
		wlrController := dromoscontroller.NewWorkloadRunnerController(k8sClient, fw,
			k8sManager.GetEventRecorderFor(controllerName)).WithName(controllerName)
		Expect(wlrController.SetupWithManager(managerCtx, k8sManager)).To(Succeed())
		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(managerCtx)).To(Succeed())
		}()

		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: restartNS,
		}})).To(Succeed())

		rrWebhookConfig = setupRestartRequestWebhookConfiguration(managerCtx, k8sClient, absCertDir,
			fmt.Sprintf("https://127.0.0.1:%d%s", webhookPort, restart.RestartRequestValidatorPath))

		// Add write permissions so impersonated identities
		impersonationCRB = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "restart-flow-impersonated-users"},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "cluster-admin"},
			Subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: tamperUser},
				{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: saUser},
			},
		}
		Expect(k8sClient.Create(managerCtx, impersonationCRB)).To(Succeed())
		tamperClient = impersonatedClient(tamperUser)
		saClient = impersonatedClient(saUser)

		// dromos renders children only for GVKs with a ready Karta
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		Expect(k8sClient.Delete(managerCtx, rrWebhookConfig)).To(Succeed())
		Expect(k8sClient.Delete(managerCtx, impersonationCRB)).To(Succeed())
		deleteAllKartas(managerCtx, k8sClient)
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	AfterEach(func() {
		// sweep hold finalizers first so cleanup never wedges on a spec that bailed early
		children := &jobsetv1alpha2.JobSetList{}
		Expect(k8sClient.List(managerCtx, children)).To(Succeed())
		for i := range children.Items {
			releaseChild(managerCtx, k8sClient, children.Items[i].Name, children.Items[i].Namespace)
		}

		deleteAllForType(managerCtx, k8sClient, &dromosv1alpha1.WorkloadRunnerList{})
		Eventually(func(g Gomega) {
			runners := &dromosv1alpha1.WorkloadRunnerList{}
			g.Expect(k8sClient.List(managerCtx, runners)).To(Succeed())
			g.Expect(runners.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())

		deleteAllForType(managerCtx, k8sClient, &runv1alpha1.RestartRequestList{})
		Eventually(func(g Gomega) {
			requests := &runv1alpha1.RestartRequestList{}
			g.Expect(k8sClient.List(managerCtx, requests)).To(Succeed())
			g.Expect(requests.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())

		deleteAllForType(managerCtx, k8sClient, &jobsetv1alpha2.JobSetList{})
		Eventually(func(g Gomega) {
			remaining := &jobsetv1alpha2.JobSetList{}
			g.Expect(k8sClient.List(managerCtx, remaining)).To(Succeed())
			g.Expect(remaining.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())
	})

	Describe("restart reconciler flow", func() {
		It("restarts a plain workload end to end through the production admission path", func() {
			runnerName := "flow-happy"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			runnerUID := runner.UID
			child := &jobsetv1alpha2.JobSet{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)
			}, timeout, interval).Should(Succeed())
			childUID := child.UID
			rrObj := newValidRestartRequest(managerCtx, k8sClient, "flow-happy-rr", runner, nodeAffinityOverride())
			rr := createResource(managerCtx, k8sClient, rrObj)

			By("completing the restart with the child recorded as cleared")
			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
				clearCondition := apimeta.FindStatusCondition(currentRR.Status.Conditions, "FrameworkClear")
				g.Expect(clearCondition).NotTo(BeNil())
				g.Expect(clearCondition.Status).To(Equal(metav1.ConditionTrue))
				g.Expect(clearCondition.Reason).To(Equal("Deleted"))
			}, timeout, interval).Should(Succeed())

			By("handing the workload back with its identity intact and instructions materialized")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Labels).NotTo(HaveKey(restart.RequestRefLabel))
				g.Expect(current.Spec.ManagedBy).To(BeEmpty())
				// the WLR object survives the restart, so the CP workloadId is inherently preserved
				g.Expect(current.UID).To(Equal(runnerUID))
				g.Expect(current.Spec.Instructions).NotTo(BeNil())
				g.Expect(current.Spec.Instructions.NodeAffinity).NotTo(BeNil())
				g.Expect(current.Spec.Instructions.NodeAffinity.NodeAffinity).NotTo(BeNil())
				required := current.Spec.Instructions.NodeAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
				g.Expect(required).NotTo(BeNil())
				g.Expect(required.NodeSelectorTerms).To(HaveLen(1))
				g.Expect(required.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
					Key:      corev1.LabelHostname,
					Operator: corev1.NodeSelectorOpNotIn,
					Values:   []string{"node-a"},
				}))
			}, timeout, interval).Should(Succeed())

			By("emitting the restart lifecycle events on the workload")
			Eventually(func(g Gomega) {
				events := &corev1.EventList{}
				g.Expect(k8sClient.List(managerCtx, events, client.InNamespace(restartNS))).To(Succeed())
				reasons := make([]string, 0, len(events.Items))
				for _, e := range events.Items {
					if e.InvolvedObject.Name == runnerName {
						reasons = append(reasons, e.Reason)
					}
				}
				g.Expect(reasons).To(ContainElements("RestartInitiated", "RestartSucceeded"))
			}, timeout, interval).Should(Succeed())

			By("recycling the child with the exclusion affinity on its pod templates")
			Eventually(func(g Gomega) {
				child := &jobsetv1alpha2.JobSet{}
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)).To(Succeed())
				g.Expect(child.UID).NotTo(Equal(childUID))
				// envtest runs no JobSet controller, so the pod template is the honest
				// assertable surface for "exclusions on the child's pods"
				g.Expect(child.Spec.ReplicatedJobs).NotTo(BeEmpty())
				for _, rj := range child.Spec.ReplicatedJobs {
					podSpec := rj.Template.Spec.Template.Spec
					g.Expect(podSpec.Affinity).NotTo(BeNil())
					g.Expect(podSpec.Affinity.NodeAffinity).NotTo(BeNil())
					required := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
					g.Expect(required).NotTo(BeNil())
					g.Expect(required.NodeSelectorTerms).To(HaveLen(1))
					g.Expect(required.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
						Key:      corev1.LabelHostname,
						Operator: corev1.NodeSelectorOpNotIn,
						Values:   []string{"node-a"},
					}))
				}
			}, timeout, interval).Should(Succeed())

			By("letting dromos overwrite the Restarting phase with a real one")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Status.Phase).NotTo(Equal(dromosv1alpha1.WorkloadRunnerPhaseRestarting))
			}, timeout, interval).Should(Succeed())
		})

		It("restarts and recycles the child when the request carries no overrides", func() {
			runnerName := "flow-noover"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			child := &jobsetv1alpha2.JobSet{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)
			}, timeout, interval).Should(Succeed())
			childUID := child.UID

			rr := createResource(managerCtx, k8sClient,
				newValidRestartRequest(managerCtx, k8sClient, "flow-noover-rr", runner, nil))

			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Spec.ManagedBy).To(BeEmpty())
				g.Expect(current.Spec.Instructions).To(BeNil()) // no spurious spec diff

				child := &jobsetv1alpha2.JobSet{}
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)).To(Succeed())
				g.Expect(child.UID).NotTo(Equal(childUID))
			}, timeout, interval).Should(Succeed())
		})

		It("fails the request and hands back when the child never finishes deleting", func() {
			runnerName := "flow-stuck"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, &jobsetv1alpha2.JobSet{})
			}, timeout, interval).Should(Succeed())
			holdChild(managerCtx, k8sClient, runnerName, restartNS)

			rr := createResource(managerCtx, k8sClient,
				newValidRestartRequest(managerCtx, k8sClient, "flow-stuck-rr", runner, nil))

			By("never reporting success while the child is still terminating")
			Consistently(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).NotTo(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
			}, time.Second*2, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseFailed))
				g.Expect(rrConditionMessages(currentRR)).To(ContainSubstring("deletion timed out"))
			}, restartFlowChildDeletionTimeout+timeout, interval).Should(Succeed())

			By("still handing the workload back on failure")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Labels).NotTo(HaveKey(restart.RequestRefLabel))
				g.Expect(current.Spec.ManagedBy).To(BeEmpty())
			}, timeout, interval).Should(Succeed())

			releaseChild(managerCtx, k8sClient, runnerName, restartNS)
		})

		It("freezes a workload mid-restart: phase, request label, and user writes are all protected", func() {
			runnerName := "flow-frozen"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, &jobsetv1alpha2.JobSet{})
			}, timeout, interval).Should(Succeed())

			By("allowing regular-user edits while the workload is NOT under restart")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				if current.Annotations == nil {
					current.Annotations = map[string]string{}
				}
				current.Annotations["pre-restart-touch"] = "ok"
				g.Expect(tamperClient.Update(managerCtx, current)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			holdChild(managerCtx, k8sClient, runnerName, restartNS)
			rr := createResource(managerCtx, k8sClient,
				newValidRestartRequest(managerCtx, k8sClient, "flow-frozen-rr", runner, nil))
			Eventually(func() bool {
				return handoffOccurred(managerCtx, k8sClient, runner, rr)
			}, timeout, interval).Should(BeTrue())

			By("keeping the Restarting phase while the child is going down")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseRestarting))
			}, timeout, interval).Should(Succeed())

			// perform write attempts
			By("denying user tampering, allowing SA writes, and protecting the request-ref label")
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())

				tampered := current.DeepCopy()
				tampered.Annotations["tamper"] = "x"
				g.Expect(tamperClient.Update(managerCtx, tampered)).
					To(MatchError(ContainSubstring("reserved for system controllers")), "deny regular user write")

				saTouched := current.DeepCopy()
				saTouched.Annotations["sa-touch"] = "ok"
				g.Expect(saClient.Update(managerCtx, saTouched)).To(Succeed(), "allow service account write")

				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				stolen := current.DeepCopy()
				stolen.Labels[restart.RequestRefLabel] = "stolen-uid"
				g.Expect(k8sClient.Update(managerCtx, stolen)).
					To(MatchError(ContainSubstring("cannot be modified once set")), "deny request-ref rewrite")

				unlabeled := current.DeepCopy()
				delete(unlabeled.Labels, restart.RequestRefLabel)
				g.Expect(k8sClient.Update(managerCtx, unlabeled)).
					To(MatchError(ContainSubstring("must be set to a non-empty value")), "deny request-ref removal")
			}, timeout, interval).Should(Succeed())

			By("completing the restart once the child is released (our own writes stay admissible)")
			releaseChild(managerCtx, k8sClient, runnerName, restartNS)
			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
			}, timeout, interval).Should(Succeed())
		})

		It("self-heals a workload born with our managedBy and no request label", func() {
			runnerName := "flow-orphan-nolabel"
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: runnerName, Namespace: restartNS},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template:  marshalRawExtension(test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build()),
					ManagedBy: restart.RestartControllerManagedBy,
				},
			}
			createResource(managerCtx, k8sClient, runner)

			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Spec.ManagedBy).To(BeEmpty())
			}, timeout, interval).Should(Succeed())

			By("letting dromos take over and render the child")
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, &jobsetv1alpha2.JobSet{})
			}, timeout, interval).Should(Succeed())
		})

		It("self-heals a workload whose request label points at a vanished request", func() {
			runnerName := "flow-orphan-dangling"
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runnerName,
					Namespace: restartNS,
					Labels:    map[string]string{restart.RequestRefLabel: "no-such-request-uid"},
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template:  marshalRawExtension(test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build()),
					ManagedBy: restart.RestartControllerManagedBy,
				},
			}
			createResource(managerCtx, k8sClient, runner)

			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				g.Expect(current.Spec.ManagedBy).To(BeEmpty())
				g.Expect(current.Labels).NotTo(HaveKey(restart.RequestRefLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("restarts the same workload a second time after the first restart completed", func() {
			runnerName := "flow-again"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)

			for i, rrName := range []string{"flow-again-first", "flow-again-second"} {
				child := &jobsetv1alpha2.JobSet{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)
				}, timeout, interval).Should(Succeed())
				childUID := child.UID

				rr := createResource(managerCtx, k8sClient,
					newValidRestartRequest(managerCtx, k8sClient, rrName, runner, nil))
				Eventually(func(g Gomega) {
					currentRR := &runv1alpha1.RestartRequest{}
					g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
					g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
				}, timeout, interval).Should(Succeed())

				Eventually(func(g Gomega) {
					current := &dromosv1alpha1.WorkloadRunner{}
					g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
					g.Expect(current.Labels).NotTo(HaveKey(restart.RequestRefLabel))
					g.Expect(current.Spec.ManagedBy).To(BeEmpty())

					child := &jobsetv1alpha2.JobSet{}
					g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, child)).To(Succeed())
					g.Expect(child.UID).NotTo(Equal(childUID), "restart %d must recycle the child", i+1)
				}, timeout, interval).Should(Succeed())
			}
		})

	})

	Describe("RestartRequest admission", func() {
		DescribeTable("denies invalid requests at CREATE",
			func(mutate func(rr *runv1alpha1.RestartRequest), expectedErr string) {
				runnerName := "flow-admission"
				runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
					test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)

				rr := newValidRestartRequest(managerCtx, k8sClient, "flow-admission-rr", runner, nil)
				mutate(rr)
				err := k8sClient.Create(managerCtx, rr)
				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
			},
			Entry("missing workload-name label",
				func(rr *runv1alpha1.RestartRequest) { rr.Labels = nil },
				"is required"),
			Entry("workload-name label not matching workloadRef",
				func(rr *runv1alpha1.RestartRequest) {
					rr.Labels[commonconstants.RunaiWorkloadNameLabel] = "someone-else"
				},
				"must match workloadRef.name"),
			Entry("foreign workload GVK",
				func(rr *runv1alpha1.RestartRequest) { rr.Spec.WorkloadRef.GVK = foreignWorkloadGVK },
				"unsupported workload GVK"),
			Entry("nonexistent workload",
				func(rr *runv1alpha1.RestartRequest) {
					rr.Spec.WorkloadRef.Name = "ghost"
					rr.Labels[commonconstants.RunaiWorkloadNameLabel] = "ghost"
				},
				"not found"),
			Entry("stale workload generation",
				func(rr *runv1alpha1.RestartRequest) { rr.Spec.WorkloadRef.Generation += 7 },
				"does not match workload generation"),
			Entry("disallowed spec override path",
				func(rr *runv1alpha1.RestartRequest) {
					rr.Spec.SpecOverrides = runv1alpha1.SpecOverrides{{
						Op:    runv1alpha1.JSONPatchOperationAdd,
						Path:  "/spec/template",
						Value: &runtime.RawExtension{Raw: []byte(`{}`)},
					}}
				},
				"is not allowed"),
		)

		It("denies a second request while one is active", func() {
			runnerName := "flow-second"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: restartNS}, &jobsetv1alpha2.JobSet{})
			}, timeout, interval).Should(Succeed())
			holdChild(managerCtx, k8sClient, runnerName, restartNS) // keeps the first request active

			first := createResource(managerCtx, k8sClient,
				newValidRestartRequest(managerCtx, k8sClient, "flow-second-winner", runner, nil))
			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(first), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseInProgress))
			}, timeout, interval).Should(Succeed())

			second := newValidRestartRequest(managerCtx, k8sClient, "flow-second-loser", runner, nil)
			err := k8sClient.Create(managerCtx, second)
			Expect(err).To(MatchError(ContainSubstring("active RestartRequest already exists")))

			releaseChild(managerCtx, k8sClient, runnerName, restartNS)
		})

		It("denies spec changes to an existing request", func() {
			runnerName := "flow-immutable"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			rr := createResource(managerCtx, k8sClient,
				newValidRestartRequest(managerCtx, k8sClient, "flow-immutable-rr", runner, nil))

			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				currentRR.Spec.WorkloadRef.Generation++
				err := k8sClient.Update(managerCtx, currentRR)
				g.Expect(err).To(MatchError(ContainSubstring("RestartRequest spec is immutable")))
			}, timeout, interval).Should(Succeed())
		})

	})

	Describe("WorkloadRunner restart rules standalone", func() {
		It("rejects a foreign managedBy value on CREATE and UPDATE", func() {
			runnerName := "flow-foreign-mgr"
			foreign := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: runnerName + "-create", Namespace: restartNS},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template:  marshalRawExtension(test.NewJobSetBuilder(runnerName+"-create", restartNS).WithDefault().Build()),
					ManagedBy: "some-other-controller",
				},
			}
			Expect(k8sClient.Create(managerCtx, foreign)).To(MatchError(ContainSubstring("supported values")))

			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				current.Spec.ManagedBy = "some-other-controller"
				err := k8sClient.Update(managerCtx, current)
				g.Expect(err).To(MatchError(ContainSubstring("supported values")))
			}, timeout, interval).Should(Succeed())
		})

		It("blocks the request-ref label on an unmanaged workload", func() {
			runnerName := "flow-reflabel"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, restartNS,
				test.NewJobSetBuilder(runnerName, restartNS).WithDefault().Build(), nil)
			Eventually(func(g Gomega) {
				current := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
				if current.Labels == nil {
					current.Labels = map[string]string{}
				}
				current.Labels[restart.RequestRefLabel] = "fake-uid"
				err := k8sClient.Update(managerCtx, current)
				g.Expect(err).To(MatchError(ContainSubstring("must be removed")))
			}, timeout, interval).Should(Succeed())
		})
	})
})
