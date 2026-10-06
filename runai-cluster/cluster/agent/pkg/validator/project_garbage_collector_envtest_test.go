/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Project orphan deletion admission", func() {
	It("allows collector finalization after agent deletion while rejecting unrelated writes", func() {
		assets := os.Getenv("KUBEBUILDER_ASSETS")
		if assets == "" {
			Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
		}
		_, filename, _, ok := goruntime.Caller(0)
		Expect(ok).To(BeTrue())
		environment := &envtest.Environment{
			BinaryAssetsDirectory: assets,
			CRDDirectoryPaths:     []string{filepath.Join(filepath.Dir(filename), "../../../sdk/config/crd/bases/run.ai_projects.yaml")},
			ErrorIfCRDPathMissing: true,
		}
		config, err := environment.Start()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(environment.Stop()).To(Succeed()) })

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
		Expect(admissionregistrationv1.AddToScheme(scheme)).To(Succeed())
		admin, err := client.New(config, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
		ctx := context.Background()
		Expect(admin.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "runai"}})).To(Succeed())
		identities := map[string]string{}
		for _, name := range []string{operands.RunaiAgentName, operands.RunaiClusterSyncName,
			operands.RunaiOperatorName, operands.KRMProjectControllerName, "generic-garbage-collector"} {
			namespace := "runai"
			if name == "generic-garbage-collector" {
				namespace = metav1.NamespaceSystem
			}
			account := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			Expect(admin.Create(ctx, account)).To(Succeed())
			identities[name] = string(account.UID)
		}
		impersonate := func(username, uid string) client.Client {
			userConfig := rest.CopyConfig(config)
			// This test isolates webhook authorization from RBAC authorization.
			userConfig.Impersonate = rest.ImpersonationConfig{UserName: username, UID: uid, Groups: []string{"system:masters"}}
			result, createErr := client.New(userConfig, client.Options{Scheme: scheme})
			Expect(createErr).NotTo(HaveOccurred())
			return result
		}
		agent := impersonate("system:serviceaccount:runai:"+operands.RunaiAgentName, identities[operands.RunaiAgentName])
		collector := impersonate("system:serviceaccount:kube-system:generic-garbage-collector", identities["generic-garbage-collector"])
		tenant := impersonate("system:serviceaccount:tenant:generic-garbage-collector", "tenant-uid")

		lookupConfig := rest.CopyConfig(config)
		lookupConfig.Impersonate = rest.ImpersonationConfig{
			UserName: "system:serviceaccount:runai:custom-nmc",
			Groups:   []string{"system:serviceaccounts", "system:serviceaccounts:runai", "system:authenticated"},
		}
		lookupClient, err := client.New(lookupConfig, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
		server := httptest.NewTLSServer(&admission.Webhook{Handler: NewProjectValidator(admin, lookupClient, nil)})
		DeferCleanup(server.Close)
		ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
		Expect(admin.Create(ctx, &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "project-deletion-test"},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{{
				Name: "project-deletion.test.run.ai", AdmissionReviewVersions: []string{"v1"},
				SideEffects:   ptr.To(admissionregistrationv1.SideEffectClassNone),
				FailurePolicy: ptr.To(admissionregistrationv1.Fail),
				ClientConfig:  admissionregistrationv1.WebhookClientConfig{URL: ptr.To(server.URL), CABundle: ca},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create,
						admissionregistrationv1.Update, admissionregistrationv1.Delete},
					Rule: admissionregistrationv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"*"}, Resources: []string{"projects"}},
				}},
			}},
		})).To(Succeed())

		newProject := func(name string) *unstructured.Unstructured {
			project := &unstructured.Unstructured{}
			project.SetGroupVersionKind(schema.GroupVersionKind{Group: "run.ai", Version: "v1", Kind: "Project"})
			project.SetName(name)
			return project
		}
		// Wait until the API server has observed the webhook registration.
		Eventually(func() string {
			probe := newProject("admission-probe")
			createErr := collector.Create(ctx, probe)
			if createErr == nil {
				Expect(agent.Delete(ctx, probe)).To(Succeed())
				return ""
			}
			return createErr.Error()
		}, 10*time.Second, 100*time.Millisecond).Should(ContainSubstring("operation not allowed"))

		project := newProject("orphan-deletion")
		Expect(agent.Create(ctx, project)).To(Succeed())
		Expect(agent.Delete(ctx, project, client.PropagationPolicy(metav1.DeletePropagationOrphan))).To(Succeed())
		Expect(admin.Get(ctx, client.ObjectKeyFromObject(project), project)).To(Succeed())
		Expect(project.GetDeletionTimestamp()).NotTo(BeNil())
		Expect(project.GetFinalizers()).To(ContainElement(metav1.FinalizerOrphanDependents))

		removal := client.RawPatch(types.MergePatchType, []byte(`{"metadata":{"finalizers":[]}}`))
		err = tenant.Patch(ctx, project.DeepCopy(), removal)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("operation not allowed"))
		err = collector.Patch(ctx, project.DeepCopy(), client.RawPatch(types.MergePatchType,
			[]byte(`{"metadata":{"finalizers":[],"labels":{"unexpected":"change"}}}`)))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("operation not allowed"))

		// Without the cross-namespace grant, a custom NMC agent cannot verify the collector UID.
		err = collector.Patch(ctx, project.DeepCopy(), removal)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("forbidden"))
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: "runai-agent-garbage-collector-runai", Namespace: metav1.NamespaceSystem},
			Rules: []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"serviceaccounts"},
				ResourceNames: []string{"generic-garbage-collector"}, Verbs: []string{"get"}}},
		}
		Expect(admin.Create(ctx, role)).To(Succeed())
		Expect(admin.Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: role.Name},
			Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: "custom-nmc", Namespace: "runai"}},
		})).To(Succeed())
		Eventually(func() error {
			return lookupClient.Get(ctx, client.ObjectKey{Name: "generic-garbage-collector", Namespace: metav1.NamespaceSystem}, &corev1.ServiceAccount{})
		}, 10*time.Second, 100*time.Millisecond).Should(Succeed())
		Expect(apierrors.IsForbidden(lookupClient.List(ctx, &corev1.ServiceAccountList{}, client.InNamespace(metav1.NamespaceSystem)))).To(BeTrue())
		Expect(apierrors.IsForbidden(lookupClient.Get(ctx,
			client.ObjectKey{Name: "default", Namespace: metav1.NamespaceSystem}, &corev1.ServiceAccount{}))).To(BeTrue())

		// Envtest has no garbage collector; submit the metadata patch it would send.
		Expect(collector.Patch(ctx, project, removal)).To(Succeed())
		Eventually(func() bool {
			return apierrors.IsNotFound(admin.Get(ctx, client.ObjectKeyFromObject(project), newProject(project.GetName())))
		}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
	})
})
