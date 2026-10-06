/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"

	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = DescribeTable("Application cleanup through admission", func(username string, deleteNamespace bool) {
	assets := os.Getenv("KUBEBUILDER_ASSETS")
	if assets == "" {
		Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}
	_, filename, _, ok := goruntime.Caller(0)
	Expect(ok).To(BeTrue())
	environment := &envtest.Environment{
		BinaryAssetsDirectory: assets,
		CRDDirectoryPaths:     []string{filepath.Join(filepath.Dir(filename), "../../../sdk/config/crd/bases/run.ai_applications.yaml")},
		ErrorIfCRDPathMissing: true,
	}
	config, err := environment.Start()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(environment.Stop()).To(Succeed()) })

	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(admissionregistrationv1.AddToScheme(scheme)).To(Succeed())
	admin, err := client.New(config, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	ctx := context.Background()
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "runai-repro-app"}}
	Expect(admin.Create(ctx, namespace)).To(Succeed())

	impersonate := func(identity string) client.Client {
		userConfig := rest.CopyConfig(config)
		// Isolate admission checks from RBAC so a denial must come from the webhook.
		userConfig.Impersonate = rest.ImpersonationConfig{UserName: identity, Groups: []string{"system:masters"}}
		result, createErr := client.New(userConfig, client.Options{Scheme: scheme})
		Expect(createErr).NotTo(HaveOccurred())
		return result
	}
	applicationController := "system:serviceaccount:runai:application-controller"
	controller := impersonate(applicationController)
	cleanup := impersonate(username)
	tenant := impersonate("system:serviceaccount:runai-repro-app:generic-garbage-collector")
	server := httptest.NewTLSServer(&admission.Webhook{Handler: NewServiceAccountValidator(admin, []string{applicationController})})
	DeferCleanup(server.Close)
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	Expect(admin.Create(ctx, &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "application-cleanup-test"},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "application-cleanup.test.run.ai", AdmissionReviewVersions: []string{"v1"},
			SideEffects:   ptr.To(admissionregistrationv1.SideEffectClassNone),
			FailurePolicy: ptr.To(admissionregistrationv1.Fail),
			ClientConfig:  admissionregistrationv1.WebhookClientConfig{URL: ptr.To(server.URL), CABundle: ca},
			Rules: []admissionregistrationv1.RuleWithOperations{{
				Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create,
					admissionregistrationv1.Update, admissionregistrationv1.Delete},
				Rule: admissionregistrationv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"*"}, Resources: []string{"applications"}},
			}},
		}},
	})).To(Succeed())

	// Wait for webhook registration before testing deletion and denied writes.
	Eventually(func() string {
		probe := &runv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Name: "admission-probe", Namespace: namespace.Name}}
		createErr := cleanup.Create(ctx, probe)
		if createErr == nil {
			Expect(controller.Delete(ctx, probe)).To(Succeed())
			return ""
		}
		return createErr.Error()
	}, 10*time.Second, 100*time.Millisecond).Should(ContainSubstring("is not permitted"))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sh.helm.release.v1.repro-am.v1", Namespace: namespace.Name},
		Type:       "helm.sh/release.v1",
	}
	Expect(admin.Create(ctx, secret)).To(Succeed())
	application := &runv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "repro-am", Namespace: namespace.Name,
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "v1", Kind: "Secret", Name: secret.Name, UID: secret.UID}},
		},
		Spec: runv1alpha1.ApplicationSpec{Source: "helm", Type: "alertmanager"},
	}
	Expect(controller.Create(ctx, application)).To(Succeed())
	application.Status.K8sPhase = "uninstalling"
	Expect(controller.Status().Update(ctx, application)).To(Succeed())

	update := application.DeepCopy()
	update.Spec.Type = "unexpected-change"
	err = cleanup.Update(ctx, update)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("is not permitted"))
	err = tenant.Delete(ctx, application)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("is not permitted"))

	// Envtest has no cleanup controllers; send the requests they would make.
	if deleteNamespace {
		Expect(admin.Delete(ctx, namespace)).To(Succeed())
		Expect(admin.Get(ctx, client.ObjectKeyFromObject(namespace), namespace)).To(Succeed())
		Expect(namespace.DeletionTimestamp).NotTo(BeNil())
		Expect(cleanup.DeleteAllOf(ctx, &runv1alpha1.Application{}, client.InNamespace(namespace.Name))).To(Succeed())
	} else {
		Expect(admin.Delete(ctx, secret)).To(Succeed())
		Expect(apierrors.IsNotFound(admin.Get(ctx, client.ObjectKeyFromObject(secret), &corev1.Secret{}))).To(BeTrue())
		Expect(cleanup.Delete(ctx, application)).To(Succeed())
	}
	Expect(apierrors.IsNotFound(admin.Get(ctx, client.ObjectKeyFromObject(application), &runv1alpha1.Application{}))).To(BeTrue())
},
	Entry("garbage collection after Helm release removal", "system:serviceaccount:kube-system:generic-garbage-collector", false),
	Entry("namespace deletion", "system:serviceaccount:kube-system:namespace-controller", true),
)
