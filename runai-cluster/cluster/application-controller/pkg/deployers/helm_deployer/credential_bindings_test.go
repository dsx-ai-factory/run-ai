/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Runtime credential bindings", func() {
	var loaded *chart.Chart
	BeforeEach(func() {
		loaded = &chart.Chart{
			Metadata: &chart.Metadata{Name: "binding-test", Version: "1.0.0", APIVersion: "v2"},
			Values: map[string]interface{}{
				"pulls": []interface{}{
					map[string]interface{}{"name": "default-first", "keep": "sibling"},
					map[string]interface{}{"name": "default-second"},
				},
				"global": map[string]interface{}{
					"imagePullSecret": map[string]interface{}{"name": "default-pull", "create": true},
					"ngcApiSecret":    map[string]interface{}{"name": "default-api", "create": true},
				},
				"custom":  map[string]interface{}{"name": "default-custom", "create": true},
				"deleted": "chart-default",
			},
		}
	})

	It("preserves default array entries and siblings and does not mutate inputs", func() {
		overrides := map[string]interface{}{"unrelated": "keep"}
		values, err := applyCredentialBindings(loaded, overrides, []v1alpha1.HelmCredentialBinding{{Variable: "pulls[0].name", SecretName: "selected"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(values["pulls"]).To(Equal([]interface{}{
			map[string]interface{}{"name": "selected", "keep": "sibling"},
			map[string]interface{}{"name": "default-second"},
		}))
		Expect(overrides).To(Equal(map[string]interface{}{"unrelated": "keep"}))
		Expect(loaded.Values["pulls"].([]interface{})[0].(map[string]interface{})["name"]).To(Equal("default-first"))
	})

	It("lets selected credentials win while preserving explicit nulls and custom create settings", func() {
		overrides := map[string]interface{}{
			"global":  map[string]interface{}{"imagePullSecret": map[string]interface{}{"name": "user-name", "create": true}},
			"deleted": nil,
		}
		values, err := applyCredentialBindings(loaded, overrides, []v1alpha1.HelmCredentialBinding{
			{Variable: "global.imagePullSecret.name", SecretName: "pull-reference"},
			{Variable: "global.ngcApiSecret.name", SecretName: "api-reference"},
			{Variable: "custom.name", SecretName: "custom-reference"},
		})
		Expect(err).NotTo(HaveOccurred())
		effective, err := chartutil.CoalesceValues(loaded, values)
		Expect(err).NotTo(HaveOccurred())
		Expect(effective["global"]).To(Equal(map[string]interface{}{
			"imagePullSecret": map[string]interface{}{"name": "pull-reference", "create": false},
			"ngcApiSecret":    map[string]interface{}{"name": "api-reference", "create": false},
		}))
		Expect(effective["custom"]).To(Equal(map[string]interface{}{"name": "custom-reference", "create": true}))
		Expect(effective).NotTo(HaveKey("deleted"))
		Expect(values["custom"]).NotTo(HaveKey("create"))
	})

	It("respects user array replacement and applies multiple bindings within it", func() {
		overrides := map[string]interface{}{"pulls": []interface{}{
			map[string]interface{}{"name": "override", "extra": "preserved"},
		}}
		values, err := applyCredentialBindings(loaded, overrides, []v1alpha1.HelmCredentialBinding{
			{Variable: "pulls[0].name", SecretName: "selected"},
			{Variable: "pulls[1].name", SecretName: "second-selected"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(values["pulls"]).To(Equal([]interface{}{
			map[string]interface{}{"name": "selected", "extra": "preserved"},
			map[string]interface{}{"name": "second-selected"},
		}))
	})

	It("does not trigger another upgrade for the values it deployed", func() {
		bindings := []v1alpha1.HelmCredentialBinding{{Variable: "pulls[0].name", SecretName: "selected"}}
		values, err := applyCredentialBindings(loaded, nil, bindings)
		Expect(err).NotTo(HaveOccurred())
		spec := &v1alpha1.HelmSpec{ChartUrl: "https://example.test/chart.tgz", CredentialBindings: bindings}
		status := &v1alpha1.HelmStatus{ChartUrl: spec.ChartUrl}
		rel := &release.Release{Chart: loaded, Config: values}
		Expect(ReleaseMatchesDesired(rel, spec, status)).To(BeTrue())
		spec.CredentialBindings = []v1alpha1.HelmCredentialBinding{{Variable: "pulls[0].name", SecretName: "changed"}}
		Expect(ReleaseMatchesDesired(rel, spec, status)).To(BeFalse())
	})

	It("preserves imported dependency defaults and remains stable after Helm processes dependencies", func() {
		loaded.Values = map[string]interface{}{}
		loaded.Metadata.Dependencies = []*chart.Dependency{{Name: "child", Version: "1.0.0", ImportValues: []interface{}{"config"}}}
		loaded.AddDependency(&chart.Chart{
			Metadata: &chart.Metadata{Name: "child", Version: "1.0.0", APIVersion: "v2"},
			Values: map[string]interface{}{"exports": map[string]interface{}{"config": map[string]interface{}{
				"pulls": []interface{}{map[string]interface{}{"name": "default", "keep": "sibling"}, map[string]interface{}{"name": "second"}},
			}}},
		})
		bindings := []v1alpha1.HelmCredentialBinding{{Variable: "pulls[0].name", SecretName: "selected"}}
		values, err := applyCredentialBindings(loaded, nil, bindings)
		Expect(err).NotTo(HaveOccurred())
		Expect(values["pulls"]).To(Equal([]interface{}{map[string]interface{}{"name": "selected", "keep": "sibling"}, map[string]interface{}{"name": "second"}}))
		Expect(chartutil.ProcessDependenciesWithMerge(loaded, values)).To(Succeed())
		spec := &v1alpha1.HelmSpec{ChartUrl: "https://example.test/chart.tgz", CredentialBindings: bindings}
		Expect(ReleaseMatchesDesired(&release.Release{Chart: loaded, Config: values}, spec, &v1alpha1.HelmStatus{ChartUrl: spec.ChartUrl})).To(BeTrue())
	})

	It("accepts contiguous array destinations regardless of submission order", func() {
		loaded.Values = map[string]interface{}{}
		values, err := applyCredentialBindings(loaded, nil, []v1alpha1.HelmCredentialBinding{
			{Variable: "pulls[1].name", SecretName: "second"},
			{Variable: "pulls[0].name", SecretName: "first"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(values["pulls"]).To(Equal([]interface{}{map[string]interface{}{"name": "first"}, map[string]interface{}{"name": "second"}}))
	})

	It("rejects overlapping destinations and scalar parents before Helm", func() {
		_, err := applyCredentialBindings(loaded, nil, []v1alpha1.HelmCredentialBinding{
			{Variable: "global.imagePullSecret.name", SecretName: "selected"},
			{Variable: "global.imagePullSecret.create", SecretName: "invalid"},
		})
		Expect(err).To(MatchError(ContainSubstring("overlap")))
		_, err = applyCredentialBindings(loaded, map[string]interface{}{"pulls": "scalar"}, []v1alpha1.HelmCredentialBinding{{Variable: "pulls[0].name", SecretName: "selected"}})
		Expect(err).To(HaveOccurred())
	})

	It("checks target Secret availability before queuing Helm", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		app := &v1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: "project"}, Spec: v1alpha1.ApplicationSpec{Helm: &v1alpha1.HelmSpec{
			CredentialBindings: []v1alpha1.HelmCredentialBinding{{Variable: "custom.name", SecretName: "selected"}},
		}}}
		_, err := NewInstallParams(context.Background(), client, app)
		Expect(err).To(MatchError(ContainSubstring("check credential synchronization and retry")))
		Expect(client.Create(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "selected", Namespace: "project"}})).To(Succeed())
		_, err = NewInstallParams(context.Background(), client, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("renders the actual AIQ 2.2.1 consumers when a retained chart is supplied", func() {
		archive := os.Getenv("RUNAI_TEST_AIQ_CHART")
		if archive == "" {
			Skip("set RUNAI_TEST_AIQ_CHART to the aiq2-web 2.2.1 archive for the real-chart render check")
		}
		aiq, err := loader.Load(archive)
		Expect(err).NotTo(HaveOccurred())
		Expect(aiq.Metadata.Version).To(Equal("2.2.1"))
		overrides := map[string]interface{}{"aiq": map[string]interface{}{"apps": map[string]interface{}{"backend": map[string]interface{}{
			"extraEnv": []interface{}{map[string]interface{}{
				"name": "NVIDIA_API_KEY", "valueFrom": map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": "old", "key": "NGC_API_KEY"}},
			}},
		}}}}
		bindings := []v1alpha1.HelmCredentialBinding{
			{Variable: "aiq.apps.backend.imagePullSecrets[0].name", SecretName: "registry-reference"},
			{Variable: "aiq.apps.frontend.imagePullSecrets[0].name", SecretName: "registry-reference"},
			{Variable: "aiq.apps.backend.extraEnv[0].valueFrom.secretKeyRef.name", SecretName: "api-reference"},
		}
		values, err := applyCredentialBindings(aiq, overrides, bindings)
		Expect(err).NotTo(HaveOccurred())
		Expect(chartutil.ProcessDependenciesWithMerge(aiq, values)).To(Succeed())
		renderValues, err := chartutil.ToRenderValues(aiq, values, chartutil.ReleaseOptions{Name: "bindings-test", Namespace: "project", IsInstall: true}, chartutil.DefaultCapabilities)
		Expect(err).NotTo(HaveOccurred())
		rendered, err := engine.Render(aiq, renderValues)
		Expect(err).NotTo(HaveOccurred())
		consumerCount := 0
		apiConsumer := false
		for _, output := range rendered {
			for _, document := range strings.Split(output, "---") {
				var deployment struct {
					Kind string `json:"kind"`
					Spec struct {
						Template corev1.PodTemplateSpec `json:"template"`
					} `json:"spec"`
				}
				if yaml.Unmarshal([]byte(document), &deployment) != nil || deployment.Kind != "Deployment" {
					continue
				}
				pod := deployment.Spec.Template.Spec
				for _, reference := range pod.ImagePullSecrets {
					if reference.Name == "registry-reference" {
						consumerCount++
					}
				}
				for _, container := range pod.Containers {
					for _, env := range container.Env {
						if env.Name == "NVIDIA_API_KEY" && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
							Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal("api-reference"))
							Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("NGC_API_KEY"))
							apiConsumer = true
						}
					}
				}
			}
		}
		Expect(consumerCount).To(Equal(2))
		Expect(apiConsumer).To(BeTrue())
		spec := &v1alpha1.HelmSpec{ChartUrl: "https://example.test/aiq.tgz", CredentialBindings: bindings, Values: map[string]apiextensionsv1.JSON{}}
		for key, value := range overrides {
			raw, err := json.Marshal(value)
			Expect(err).NotTo(HaveOccurred())
			spec.Values[key] = apiextensionsv1.JSON{Raw: raw}
		}
		Expect(ReleaseMatchesDesired(&release.Release{Chart: aiq, Config: values}, spec, &v1alpha1.HelmStatus{ChartUrl: spec.ChartUrl})).To(BeTrue())
	})
})
