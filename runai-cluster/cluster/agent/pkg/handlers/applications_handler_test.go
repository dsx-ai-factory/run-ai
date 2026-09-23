/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type applicationSyncTransport func(*http.Request) (*http.Response, error)

func (f applicationSyncTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type applicationDeleteFailureClient struct{ client.Client }

func (applicationDeleteFailureClient) Delete(context.Context, client.Object, ...client.DeleteOption) error {
	return fmt.Errorf("test deletion failure")
}

var _ = Describe("ApplicationsHandler", func() {
	Describe("ApplicationSpecCopier", func() {
		var copier *ApplicationSpecCopier

		BeforeEach(func() {
			copier = &ApplicationSpecCopier{}
		})

		It("should deep-copy spec from one Application to another", func() {
			from := &runv1alpha1.Application{
				Spec: runv1alpha1.ApplicationSpec{
					Source: "helm",
					Type:   "my-chart",
					Helm: &runv1alpha1.HelmSpec{
						ChartRef: &runv1alpha1.ChartRef{
							Repository: "https://charts.example.com",
							Name:       "my-chart",
							Version:    "1.0.0",
						},
						Values: map[string]apiextensionsv1.JSON{
							"key": {Raw: []byte(`"value"`)},
						},
					},
				},
			}
			to := &runv1alpha1.Application{}

			err := copier.DeepCopySpec(from, to)
			Expect(err).NotTo(HaveOccurred())
			Expect(to.Spec.Source).To(Equal("helm"))
			Expect(to.Spec.Type).To(Equal("my-chart"))
			Expect(to.Spec.Helm).NotTo(BeNil())
			Expect(to.Spec.Helm.ChartRef.Name).To(Equal("my-chart"))
			Expect(to.Spec.Helm.Values).To(HaveKey("key"))
		})

		It("should overwrite existing spec on target", func() {
			from := &runv1alpha1.Application{
				Spec: runv1alpha1.ApplicationSpec{
					Source: "dynamo",
					Type:   "new-type",
				},
			}
			to := &runv1alpha1.Application{
				Spec: runv1alpha1.ApplicationSpec{
					Source: "helm",
					Type:   "old-type",
					Helm: &runv1alpha1.HelmSpec{
						ChartUrl: "https://old.example.com/chart.tgz",
					},
				},
			}

			err := copier.DeepCopySpec(from, to)
			Expect(err).NotTo(HaveOccurred())
			Expect(to.Spec.Source).To(Equal("dynamo"))
			Expect(to.Spec.Type).To(Equal("new-type"))
			Expect(to.Spec.Helm).To(BeNil())
		})

		It("should return error when 'from' is not an Application", func() {
			from := &corev1.Pod{}
			to := &runv1alpha1.Application{}

			err := copier.DeepCopySpec(from, to)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("from object is not an Application"))
		})

		It("should return error when 'to' is not an Application", func() {
			from := &runv1alpha1.Application{
				Spec: runv1alpha1.ApplicationSpec{Source: "helm"},
			}
			to := &corev1.Pod{}

			err := copier.DeepCopySpec(from, to)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("to object is not an Application"))
		})
	})

	Describe("convertToApplication", func() {
		var (
			handler      *ApplicationsHandler
			fakeClient   client.Client
			testProjInfo applicationProjectInfo
		)

		BeforeEach(func() {
			scheme := runtime.NewScheme()
			Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(runaiv2.AddToScheme(scheme)).To(Succeed())
			Expect(corev1.AddToScheme(scheme)).To(Succeed())

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

			mockBackendClient := &runaiBackendClient.BackendClient{}
			object := &runv1alpha1.Application{}
			objectInspector := ApplicationObjectInspector{
				specCopier: &ApplicationSpecCopier{},
			}
			handlingOptions := HandlingOptions{deletePropagationOrphan: false}
			commonHandler, err := NewCommonHandler(mockBackendClient, fakeClient, fakeClient, object, objectInspector, handlingOptions)
			Expect(err).NotTo(HaveOccurred())

			handler = &ApplicationsHandler{
				CommonHandler: commonHandler,
			}

			testProjInfo = applicationProjectInfo{
				namespace:      "test-namespace",
				projectName:    "test-project",
				projectId:      "proj-1",
				departmentId:   "dept-1",
				departmentName: "dept-1-name",
			}

			_ = rcpconfig.Set(&config.AgentConfig{ClusterUuid: "test-cluster-uuid"})
		})

		It("should correctly map fields to Application CRD", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-1"),
				Name:      convutils.Ref("my-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("my-chart"),
				ProjectId: convutils.Ref("proj-1"),
			}
			updateCount := int64(5)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Name).To(Equal("my-app"))
			Expect(app.Namespace).To(Equal("test-namespace"))
			Expect(app.Spec.Source).To(Equal("helm"))
			Expect(app.Spec.Type).To(Equal("my-chart"))
			Expect(app.Annotations[annotationApplicationUpdateCounter]).To(Equal("5"))
			Expect(app.Annotations[clusterCommonConstants.ControlPlaneApplicationIdAnnotation]).To(Equal("app-id-1"))
			Expect(app.Annotations[annotationClusterId]).To(Equal("test-cluster-uuid"))
			Expect(app.Annotations[clusterCommonConstants.AnnotationRunaiApp]).To(Equal("true"))
			Expect(app.Labels[clusterCommonConstants.DepartmentIdLabel]).To(Equal("dept-1"))
			Expect(app.Labels[clusterCommonConstants.DepartmentNameLabel]).To(Equal("dept-1-name"))
			Expect(app.Labels[clusterCommonConstants.ProjectNameLabel]).To(Equal("test-project"))
			Expect(app.Labels[clusterCommonConstants.ProjectIdLabel]).To(Equal("proj-1"))
			Expect(app.Spec.Helm).To(BeNil())
		})

		It("should set helm spec with chartRef when Spec is present", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-2"),
				Name:      convutils.Ref("helm-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("nginx"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("https://charts.bitnami.com/bitnami"),
						ChartName:    convutils.Ref("nginx"),
						ChartVersion: convutils.Ref("15.0.0"),
					},
				},
			}
			updateCount := int64(1)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm).NotTo(BeNil())
			Expect(app.Spec.Helm.ChartRef).NotTo(BeNil())
			Expect(app.Spec.Helm.ChartRef.Repository).To(Equal("https://charts.bitnami.com/bitnami"))
			Expect(app.Spec.Helm.ChartRef.Name).To(Equal("nginx"))
			Expect(app.Spec.Helm.ChartRef.Version).To(Equal("15.0.0"))
			Expect(app.Spec.Helm.ChartUrl).To(BeEmpty())
		})

		It("should set helm spec with chartUrl when ChartURL is set", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-3"),
				Name:      convutils.Ref("url-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("custom"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						ChartURL: convutils.Ref("https://example.com/charts/my-chart-1.0.0.tgz"),
					},
				},
			}
			updateCount := int64(1)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm).NotTo(BeNil())
			Expect(app.Spec.Helm.ChartUrl).To(Equal("https://example.com/charts/my-chart-1.0.0.tgz"))
			Expect(app.Spec.Helm.ChartRef).To(BeNil())
			Expect(app.Spec.Helm.RequiresControlPlaneAuth).To(BeFalse())
		})

		It("should set requiresControlPlaneAuth when the control-plane flag is set", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:                       convutils.Ref("app-id-cp"),
				Name:                     convutils.Ref("cp-chart-app"),
				Source:                   convutils.Ref("helm"),
				Type:                     convutils.Ref("custom"),
				ProjectId:                convutils.Ref("proj-1"),
				RequiresControlPlaneAuth: convutils.Ref(true),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						ChartURL: convutils.Ref("https://cp.example.com/api/v1/ai-applications/charts/abc/content"),
					},
				},
			}
			updateCount := int64(1)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm).NotTo(BeNil())
			Expect(app.Spec.Helm.ChartUrl).To(Equal("https://cp.example.com/api/v1/ai-applications/charts/abc/content"))
			Expect(app.Spec.Helm.RequiresControlPlaneAuth).To(BeTrue())
		})

		It("should convert values map to apiextensionsv1.JSON", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-4"),
				Name:      convutils.Ref("values-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("my-chart"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("https://charts.example.com"),
						ChartName:    convutils.Ref("my-chart"),
						ChartVersion: convutils.Ref("1.0.0"),
					},
					Values: map[string]interface{}{
						"replicaCount": float64(3),
						"image":        "nginx:latest",
					},
				},
			}
			updateCount := int64(1)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm.Values).To(HaveLen(2))
			Expect(app.Spec.Helm.Values).To(HaveKey("replicaCount"))
			Expect(app.Spec.Helm.Values).To(HaveKey("image"))
			Expect(app.Spec.Helm.Values["replicaCount"].Raw).To(Equal([]byte(`3`)))
			Expect(app.Spec.Helm.Values["image"].Raw).To(Equal([]byte(`"nginx:latest"`)))
		})

		It("should resolve auth secret from asset ID label", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-registry-secret",
					Namespace: "test-namespace",
					Labels: map[string]string{
						assetIdLabel: "asset-123",
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), secret)).To(Succeed())

			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-5"),
				Name:      convutils.Ref("auth-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("private-chart"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("https://private.charts.example.com"),
						ChartName:    convutils.Ref("private-chart"),
						ChartVersion: convutils.Ref("2.0.0"),
					},
					Auth: &workloads_openapi.AIApplicationAuth{
						AssetId: convutils.Ref("asset-123"),
					},
				},
			}
			updateCount := int64(1)

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm.Auth).NotTo(BeNil())
			Expect(app.Spec.Helm.Auth.SecretRef).NotTo(BeNil())
			Expect(app.Spec.Helm.Auth.SecretRef.Name).To(Equal("my-registry-secret"))
		})

		It("should prefer generic secret for HTTP Helm repository auth", func() {
			assetId := "asset-ngc-http"
			Expect(fakeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ngc-docker-secret",
					Namespace: "test-namespace",
					Labels: map[string]string{
						assetIdLabel:                           assetId,
						clusterCommonConstants.RunaiAssetLabel: dockerRegistryKindLabel,
					},
				},
			})).To(Succeed())
			Expect(fakeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ngc-generic-secret",
					Namespace: "test-namespace",
					Labels: map[string]string{
						assetIdLabel:                           assetId,
						clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
					},
				},
			})).To(Succeed())

			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-ngc-http"),
				Name:      convutils.Ref("auth-http-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("private-chart"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("https://helm.ngc.nvidia.com/nvidia/nim"),
						ChartName:    convutils.Ref("nim-chart"),
						ChartVersion: convutils.Ref("1.0.0"),
					},
					Auth: &workloads_openapi.AIApplicationAuth{
						AssetId: convutils.Ref(assetId),
					},
				},
			}

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm.Auth.SecretRef.Name).To(Equal("ngc-generic-secret"))
		})

		It("should prefer docker-registry secret for OCI Helm repository auth", func() {
			assetId := "asset-ngc-oci"
			Expect(fakeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ngc-generic-secret-oci",
					Namespace: "test-namespace",
					Labels: map[string]string{
						assetIdLabel:                           assetId,
						clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
					},
				},
			})).To(Succeed())
			Expect(fakeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ngc-docker-secret-oci",
					Namespace: "test-namespace",
					Labels: map[string]string{
						assetIdLabel:                           assetId,
						clusterCommonConstants.RunaiAssetLabel: dockerRegistryKindLabel,
					},
				},
			})).To(Succeed())

			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-ngc-oci"),
				Name:      convutils.Ref("auth-oci-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("private-chart"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("oci://nvcr.io/nvidia"),
						ChartName:    convutils.Ref("private-chart"),
						ChartVersion: convutils.Ref("1.0.0"),
					},
					Auth: &workloads_openapi.AIApplicationAuth{
						AssetId: convutils.Ref(assetId),
					},
				},
			}

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Helm.Auth.SecretRef.Name).To(Equal("ngc-docker-secret-oci"))
		})

		It("should return error when auth secret is not found", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-6"),
				Name:      convutils.Ref("missing-secret-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("private-chart"),
				ProjectId: convutils.Ref("proj-1"),
				Spec: &workloads_openapi.AIApplicationSpec{
					ChartRef: workloads_openapi.AIApplicationSpecChartRef{
						Repo:         convutils.Ref("https://private.charts.example.com"),
						ChartName:    convutils.Ref("private-chart"),
						ChartVersion: convutils.Ref("2.0.0"),
					},
					Auth: &workloads_openapi.AIApplicationAuth{
						AssetId: convutils.Ref("nonexistent-asset"),
					},
				},
			}
			updateCount := int64(1)

			_, err := handler.convertToApplication(context.Background(), testProjInfo, item, &updateCount)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no secret found"))
		})

		It("should handle nil updateCount", func() {
			item := &workloads_openapi.AIApplicationSyncItem{
				Id:        convutils.Ref("app-id-7"),
				Name:      convutils.Ref("nil-count-app"),
				Source:    convutils.Ref("helm"),
				Type:      convutils.Ref("my-chart"),
				ProjectId: convutils.Ref("proj-1"),
			}

			app, err := handler.convertToApplication(context.Background(), testProjInfo, item, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Annotations[annotationApplicationUpdateCounter]).To(Equal("0"))
		})
	})

	Describe("resolveSecretNameFromAssetId", func() {
		var (
			handler    *ApplicationsHandler
			fakeClient client.Client
		)

		BeforeEach(func() {
			scheme := runtime.NewScheme()
			Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(corev1.AddToScheme(scheme)).To(Succeed())

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

			mockBackendClient := &runaiBackendClient.BackendClient{}
			object := &runv1alpha1.Application{}
			objectInspector := ApplicationObjectInspector{
				specCopier: &ApplicationSpecCopier{},
			}
			handlingOptions := HandlingOptions{deletePropagationOrphan: false}
			commonHandler, err := NewCommonHandler(mockBackendClient, fakeClient, fakeClient, object, objectInspector, handlingOptions)
			Expect(err).NotTo(HaveOccurred())

			handler = &ApplicationsHandler{
				CommonHandler: commonHandler,
			}
		})

		It("should find a secret by asset-id label in the correct namespace", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "found-secret",
					Namespace: "my-namespace",
					Labels: map[string]string{
						assetIdLabel: "target-asset",
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), secret)).To(Succeed())

			secretInOtherNs := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-secret",
					Namespace: "other-namespace",
					Labels: map[string]string{
						assetIdLabel: "target-asset",
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), secretInOtherNs)).To(Succeed())

			name, err := handler.resolveSecretNameFromAssetId(context.Background(), "my-namespace", "target-asset")
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("found-secret"))
		})

		It("should return error when no secret matches", func() {
			_, err := handler.resolveSecretNameFromAssetId(context.Background(), "empty-namespace", "no-such-asset")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no secret found"))
		})

		It("should not match secrets with a different asset-id", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-asset-secret",
					Namespace: "my-namespace",
					Labels: map[string]string{
						assetIdLabel: "different-asset",
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), secret)).To(Succeed())

			_, err := handler.resolveSecretNameFromAssetId(context.Background(), "my-namespace", "target-asset")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no secret found"))
		})
	})

	Describe("processApplications", func() {
		var (
			handler    *ApplicationsHandler
			fakeClient client.Client
			scheme     *runtime.Scheme
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(runaiv2.AddToScheme(scheme)).To(Succeed())
			Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
			Expect(corev1.AddToScheme(scheme)).To(Succeed())

			project := &runaiv2.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-project",
					Labels: map[string]string{
						runaiv2.ProjectIdLabel:          "proj-1",
						schedv2alpha2.DepartmentIdLabel: "dept-1",
					},
				},
				Spec: runaiv2.ProjectSpec{
					Department: "dept-1-name",
				},
				Status: runaiv2.ProjectStatus{
					Namespace: "proj-namespace",
				},
			}

			// The kai.resources project mirrors the legacy one so projectInfoMap resolves
			// the same namespace/department.
			kaiProject := &kaiorg.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-project",
					Labels: map[string]string{
						runaiv2.ProjectIdLabel:          "proj-1",
						schedv2alpha2.DepartmentIdLabel: "dept-1",
					},
				},
				Spec: kaiorg.ProjectSpec{
					Parent: "dept-1-name",
				},
				Status: kaiorg.ProjectStatus{
					Namespace: "proj-namespace",
				},
			}

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(project, kaiProject).
				Build()

			mockBackendClient := &runaiBackendClient.BackendClient{}
			object := &runv1alpha1.Application{}
			objectInspector := ApplicationObjectInspector{
				specCopier: &ApplicationSpecCopier{},
			}
			handlingOptions := HandlingOptions{deletePropagationOrphan: false}
			commonHandler, err := NewCommonHandler(mockBackendClient, fakeClient, fakeClient, object, objectInspector, handlingOptions)
			Expect(err).NotTo(HaveOccurred())

			handler = &ApplicationsHandler{
				CommonHandler: commonHandler,
				syncReport:    &ApplicationSyncReportInfo{},
			}

			_ = rcpconfig.Set(&config.AgentConfig{ClusterUuid: "test-cluster-uuid"})
		})

		DescribeTable("retries pending credentials across restarts without losing later updates", func(deleted, deletionFails bool) {
			ctx := context.Background()
			Expect(fakeClient.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: configmaps.AgentConfigName, Namespace: clusterCommonConstants.NamespaceRunAi,
			}, Data: map[string]string{configmaps.ApplicationsSyncCounter: "0"}})).To(Succeed())
			waiting := workloads_openapi.AIApplicationSyncInfo{
				UpdateCount: convutils.Ref(int64(1)),
				Application: &workloads_openapi.AIApplicationSyncItem{
					Id: convutils.Ref("waiting-id"), Name: convutils.Ref("waiting"), ProjectId: convutils.Ref("proj-1"),
					Spec: &workloads_openapi.AIApplicationSpec{CredentialBindings: []workloads_openapi.AIApplicationCredentialBinding{{
						Variable: "custom.name", AssetId: "asset-id", CredentialType: "generic",
					}}},
				},
			}
			other := workloads_openapi.AIApplicationSyncInfo{
				UpdateCount: convutils.Ref(int64(2)),
				Application: &workloads_openapi.AIApplicationSyncItem{
					Id: convutils.Ref("other-id"), Name: convutils.Ref("other"), ProjectId: convutils.Ref("proj-1"),
				},
			}
			updates := []workloads_openapi.AIApplicationSyncInfo{waiting, other}
			var requested []int64
			var messages []string
			cfg := workloads_openapi.NewConfiguration()
			cfg.HTTPClient = &http.Client{Transport: applicationSyncTransport(func(request *http.Request) (*http.Response, error) {
				after, err := strconv.ParseInt(request.URL.Query().Get("updatesAfter"), 10, 64)
				Expect(err).NotTo(HaveOccurred())
				requested = append(requested, after)
				if request.Body != nil {
					var report workloads_openapi.ClusterSyncReport
					Expect(json.NewDecoder(request.Body).Decode(&report)).To(Succeed())
					for _, update := range report.Updates {
						if message := update.Message.Get(); message != nil {
							messages = append(messages, *message)
						}
					}
				}
				response := workloads_openapi.AIApplicationSyncResponse{
					SystemInfo: &workloads_openapi.SystemInfo{HighestUpdate: updates[len(updates)-1].UpdateCount},
				}
				for _, update := range updates {
					if update.GetUpdateCount() > after {
						response.Applications = []workloads_openapi.AIApplicationSyncInfo{update}
						response.ResponseInfo = &workloads_openapi.ResponseInfo{HighestUpdate: update.UpdateCount}
						break
					}
				}
				raw, err := json.Marshal(response)
				Expect(err).NotTo(HaveOccurred())
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
			})}
			handler.backendClient.WorkloadsClient = workloads_openapi.NewAPIClient(cfg)
			Expect(handler.Handle()).To(Succeed())
			Expect(requested).To(Equal([]int64{0, 1}))
			Expect(messages).To(ContainElement(ContainSubstring("check credential synchronization")))
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "other", Namespace: "proj-namespace"}, &runv1alpha1.Application{})).To(Succeed())
			Expect(errors.IsNotFound(fakeClient.Get(ctx, client.ObjectKey{Name: "waiting", Namespace: "proj-namespace"}, &runv1alpha1.Application{}))).To(BeTrue())
			persisted, err := configmaps.GetAgentConfig(fakeClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted.ApplicationsSyncCounter).To(BeZero())
			handler = &ApplicationsHandler{CommonHandler: handler.CommonHandler, syncCounter: persisted.ApplicationsSyncCounter}
			if deleted {
				waiting.UpdateCount = convutils.Ref(int64(3))
				waiting.Application.DesiredPhase = convutils.Ref(applicationDesiredPhaseDeleted)
				updates = []workloads_openapi.AIApplicationSyncInfo{other, waiting}
				if deletionFails {
					other.UpdateCount = convutils.Ref(int64(4))
					updates = []workloads_openapi.AIApplicationSyncInfo{waiting, other}
					Expect(fakeClient.Create(ctx, &runv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: "proj-namespace"}})).To(Succeed())
					handler.runaiClient = applicationDeleteFailureClient{Client: fakeClient}
					Expect(handler.Handle()).To(MatchError(ContainSubstring("test deletion failure")))
					// A higher page cursor would tell the backend to prune this failed delete.
					Expect(requested).To(Equal([]int64{0, 1, 0}))
					Expect(handler.syncCounter).To(BeZero())
					handler.runaiClient = fakeClient
				}
			} else {
				Expect(fakeClient.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
					Name: "selected", Namespace: "proj-namespace", Labels: map[string]string{assetIdLabel: "asset-id", clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel},
				}})).To(Succeed())
			}
			Expect(handler.Handle()).To(Succeed())
			app := &runv1alpha1.Application{}
			err = fakeClient.Get(ctx, client.ObjectKey{Name: "waiting", Namespace: "proj-namespace"}, app)
			if deleted {
				Expect(errors.IsNotFound(err)).To(BeTrue())
				Expect(handler.syncCounter).To(Equal(updates[len(updates)-1].GetUpdateCount()))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(app.Spec.Helm.CredentialBindings[0].SecretName).To(Equal("selected"))
				Expect(handler.syncCounter).To(Equal(int64(2)))
			}
		}, Entry("when the Secret arrives", false, false), Entry("when the pending application is deleted", true, false), Entry("when a deletion fails before later updates", true, true))

		It("should create applications in the cluster", func() {
			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(1)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:        convutils.Ref("app-1"),
							Name:      convutils.Ref("new-app"),
							Source:    convutils.Ref("helm"),
							Type:      convutils.Ref("nginx"),
							ProjectId: convutils.Ref("proj-1"),
							Spec: &workloads_openapi.AIApplicationSpec{
								ChartRef: workloads_openapi.AIApplicationSpecChartRef{
									Repo:         convutils.Ref("https://charts.example.com"),
									ChartName:    convutils.Ref("nginx"),
									ChartVersion: convutils.Ref("1.0.0"),
								},
							},
						},
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 1)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			app := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "new-app",
				Namespace: "proj-namespace",
			}, app)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Spec.Source).To(Equal("helm"))
			Expect(app.Spec.Helm).NotTo(BeNil())
			Expect(app.Spec.Helm.ChartRef.Name).To(Equal("nginx"))
			Expect(app.Labels[clusterCommonConstants.DepartmentIdLabel]).To(Equal("dept-1"))
			Expect(app.Labels[clusterCommonConstants.DepartmentNameLabel]).To(Equal("dept-1-name"))
			Expect(app.Labels[clusterCommonConstants.ProjectNameLabel]).To(Equal("test-project"))
			Expect(app.Labels[clusterCommonConstants.ProjectIdLabel]).To(Equal("proj-1"))
		})

		It("should delete applications with desiredPhase deleted", func() {
			existing := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "to-delete",
					Namespace: "proj-namespace",
				},
				Spec: runv1alpha1.ApplicationSpec{
					Source: "helm",
					Type:   "old-chart",
				},
			}
			Expect(fakeClient.Create(context.Background(), existing)).To(Succeed())

			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(2)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:           convutils.Ref("app-del"),
							Name:         convutils.Ref("to-delete"),
							ProjectId:    convutils.Ref("proj-1"),
							DesiredPhase: convutils.Ref("deleted"),
						},
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 1)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			app := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "to-delete",
				Namespace: "proj-namespace",
			}, app)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should handle mixed create and delete", func() {
			existing := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delete-me",
					Namespace: "proj-namespace",
				},
				Spec: runv1alpha1.ApplicationSpec{Source: "helm"},
			}
			Expect(fakeClient.Create(context.Background(), existing)).To(Succeed())

			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(1)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:        convutils.Ref("app-new"),
							Name:      convutils.Ref("create-me"),
							Source:    convutils.Ref("helm"),
							Type:      convutils.Ref("nginx"),
							ProjectId: convutils.Ref("proj-1"),
							Spec: &workloads_openapi.AIApplicationSpec{
								ChartRef: workloads_openapi.AIApplicationSpecChartRef{
									Repo:         convutils.Ref("https://charts.example.com"),
									ChartName:    convutils.Ref("nginx"),
									ChartVersion: convutils.Ref("1.0.0"),
								},
							},
						},
					},
					{
						UpdateCount: convutils.Ref(int64(2)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:           convutils.Ref("app-del"),
							Name:         convutils.Ref("delete-me"),
							ProjectId:    convutils.Ref("proj-1"),
							DesiredPhase: convutils.Ref("deleted"),
						},
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 2)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			created := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "create-me",
				Namespace: "proj-namespace",
			}, created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Spec.Helm.ChartRef.Name).To(Equal("nginx"))

			deleted := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "delete-me",
				Namespace: "proj-namespace",
			}, deleted)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should skip applications with unknown project and record error in sync report", func() {
			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(1)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:        convutils.Ref("app-unknown"),
							Name:      convutils.Ref("unknown-proj-app"),
							Source:    convutils.Ref("helm"),
							Type:      convutils.Ref("nginx"),
							ProjectId: convutils.Ref("nonexistent-project"),
							Spec: &workloads_openapi.AIApplicationSpec{
								ChartRef: workloads_openapi.AIApplicationSpecChartRef{
									Repo:         convutils.Ref("https://charts.example.com"),
									ChartName:    convutils.Ref("nginx"),
									ChartVersion: convutils.Ref("1.0.0"),
								},
							},
						},
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 1)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			app := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "unknown-proj-app",
				Namespace: "proj-namespace",
			}, app)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			Expect(handler.syncReport.theReport.Updates[0].Message.Get()).NotTo(BeNil())
			Expect(*handler.syncReport.theReport.Updates[0].Message.Get()).To(ContainSubstring("unknown project"))
		})

		It("should record error in sync report for failed secret resolution", func() {
			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(1)),
						Application: &workloads_openapi.AIApplicationSyncItem{
							Id:        convutils.Ref("app-bad-auth"),
							Name:      convutils.Ref("bad-auth-app"),
							Source:    convutils.Ref("helm"),
							Type:      convutils.Ref("private"),
							ProjectId: convutils.Ref("proj-1"),
							Spec: &workloads_openapi.AIApplicationSpec{
								ChartRef: workloads_openapi.AIApplicationSpecChartRef{
									Repo:         convutils.Ref("https://private.example.com"),
									ChartName:    convutils.Ref("private"),
									ChartVersion: convutils.Ref("1.0.0"),
								},
								Auth: &workloads_openapi.AIApplicationAuth{
									AssetId: convutils.Ref("missing-asset-id"),
								},
							},
						},
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 1)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			app := &runv1alpha1.Application{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{
				Name:      "bad-auth-app",
				Namespace: "proj-namespace",
			}, app)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			Expect(handler.syncReport.theReport.Updates[0].Message.Get()).NotTo(BeNil())
			Expect(*handler.syncReport.theReport.Updates[0].Message.Get()).To(ContainSubstring("no secret found"))
		})

		It("should handle nil application in sync info", func() {
			response := &workloads_openapi.AIApplicationSyncResponse{
				Applications: []workloads_openapi.AIApplicationSyncInfo{
					{
						UpdateCount: convutils.Ref(int64(1)),
						Application: nil,
					},
				},
			}

			handler.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, 1)
			err := handler.processApplications(context.Background(), response)
			Expect(err).NotTo(HaveOccurred())

			Expect(handler.syncReport.theReport.Updates[0].Message.Get()).NotTo(BeNil())
			Expect(*handler.syncReport.theReport.Updates[0].Message.Get()).To(ContainSubstring("nil application"))
		})
	})
})
