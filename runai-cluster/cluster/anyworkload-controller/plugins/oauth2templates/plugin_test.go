/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func proxy(name string) corev1.Container {
	return corev1.Container{
		Name:  name,
		Image: "quay.io/oauth2-proxy/oauth2-proxy:v7.15.4",
		Args:  []string{"--custom-templates-dir=" + oauth2proxy.TemplatesPath},
	}
}

func podResource(t *testing.T, pod corev1.PodSpec, template bool) (*unstructured.Unstructured, *kartav1alpha1.Karta, []string) {
	t.Helper()
	podMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	gomega.NewWithT(t).Expect(err).To(gomega.Succeed())
	definition := &kartav1alpha1.SpecDefinition{PodSpecPath: ptr.To(".spec")}
	path := []string{"spec"}
	var spec any = podMap
	if template {
		spec = map[string]any{"template": map[string]any{"spec": podMap}}
		definition = &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.template")}
		path = []string{"spec", "template", "spec"}
	}
	return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "batch/v1", "kind": "Job", "metadata": map[string]any{"name": "test"}, "spec": spec,
		}}, &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{Name: "job", SpecDefinition: definition},
			},
		}}, path
}

func TestMountGeneratedRegularAndInitProxies(t *testing.T) {
	for _, template := range []bool{false, true} {
		name := "pod-spec"
		if template {
			name = "pod-template"
		}
		t.Run(name, func(t *testing.T) {
			userContainer := proxy("user-proxy")
			pod := corev1.PodSpec{
				Containers:     []corev1.Container{proxy("runai-oauth2-proxy"), userContainer},
				InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0"), proxy("runai-oauth2-proxy-12")},
			}
			resource, karta, path := podResource(t, pod, template)
			plugin := &Plugin{configMapName: "custom-template-map"}
			gomega.NewWithT(t).Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
			podMap, found, err := unstructured.NestedMap(resource.Object, path...)
			gomega.NewWithT(t).Expect(err).To(gomega.Succeed())
			gomega.NewWithT(t).Expect(found).To(gomega.BeTrue())
			var actual corev1.PodSpec
			gomega.NewWithT(t).Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(podMap, &actual)).To(gomega.Succeed())
			gomega.NewWithT(t).Expect(actual.Volumes).To(gomega.HaveLen(1))
			gomega.NewWithT(t).Expect(actual.Volumes[0].Name).To(gomega.Equal(oauth2proxy.VolumeName))
			gomega.NewWithT(t).Expect(actual.Volumes[0].ConfigMap.Name).To(gomega.Equal("custom-template-map"))
			gomega.NewWithT(t).Expect(actual.Volumes[0].ConfigMap.Items).To(gomega.Equal([]corev1.KeyToPath{{Key: "error.html", Path: "error.html"}}))
			gomega.NewWithT(t).Expect(actual.Volumes[0].ConfigMap.Optional).To(gomega.BeNil())
			for _, container := range []corev1.Container{actual.Containers[0], actual.InitContainers[0], actual.InitContainers[1]} {
				gomega.NewWithT(t).Expect(container.VolumeMounts).To(gomega.Equal([]corev1.VolumeMount{{Name: oauth2proxy.VolumeName, MountPath: oauth2proxy.TemplatesPath, ReadOnly: true}}))
			}
			gomega.NewWithT(t).Expect(actual.Containers[1]).To(gomega.Equal(userContainer))
			once := resource.DeepCopy()
			gomega.NewWithT(t).Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
			gomega.NewWithT(t).Expect(resource.Object).To(gomega.Equal(once.Object), "repeated mutation must converge")
		})
	}
}

func TestUnrelatedContainersAndDisabledConfigurationRemainUnchanged(t *testing.T) {
	wrongFlag := proxy("runai-oauth2-proxy")
	wrongFlag.Args = []string{"--custom-templates-dir=/custom"}
	pod := corev1.PodSpec{Containers: []corev1.Container{
		proxy("oauth2-proxy"), proxy("runai-oauth2-proxy-custom"), proxy("runai-oauth2-proxy-"), wrongFlag,
	}}
	resource, karta, _ := podResource(t, pod, true)
	original := resource.DeepCopy()
	gomega.NewWithT(t).Expect((&Plugin{configMapName: oauth2proxy.ConfigMapName}).MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
	gomega.NewWithT(t).Expect(resource.Object).To(gomega.Equal(original.Object))

	t.Setenv(oauth2proxy.TemplatesConfigMapEnvVar, "")
	_, err := New(context.Background(), framework.PluginDeps{})
	gomega.NewWithT(t).Expect(err).To(gomega.MatchError(framework.ErrPluginDisabled))
	t.Setenv(oauth2proxy.TemplatesConfigMapEnvVar, oauth2proxy.ConfigMapName)
	built, err := New(context.Background(), framework.PluginDeps{})
	gomega.NewWithT(t).Expect(err).To(gomega.Succeed())
	gomega.NewWithT(t).Expect(built.(*Plugin).configMapName).To(gomega.Equal(oauth2proxy.ConfigMapName))
}

func TestMountAllKartaChildTemplates(t *testing.T) {
	g := gomega.NewWithT(t)
	pod := corev1.PodSpec{InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0")}}
	resource, karta, _ := podResource(t, pod, true)
	template, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
	g.Expect(err).To(gomega.Succeed())
	resource.Object["spec"] = map[string]any{
		"leaderWorkerTemplate": map[string]any{"leaderTemplate": template, "workerTemplate": template},
	}
	karta.Spec.StructureDefinition.RootComponent.SpecDefinition = nil
	for _, name := range []string{"leaderTemplate", "workerTemplate"} {
		karta.Spec.StructureDefinition.ChildComponents = append(karta.Spec.StructureDefinition.ChildComponents, kartav1alpha1.ComponentDefinition{
			Name:           name,
			SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.leaderWorkerTemplate." + name)},
		})
	}
	plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName}
	g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
	for _, name := range []string{"leaderTemplate", "workerTemplate"} {
		volumes, found, err := unstructured.NestedSlice(resource.Object, "spec", "leaderWorkerTemplate", name, "spec", "volumes")
		g.Expect(err).To(gomega.Succeed())
		g.Expect(found).To(gomega.BeTrue())
		g.Expect(volumes).To(gomega.HaveLen(1))
	}
	once := resource.DeepCopy()
	g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
	g.Expect(resource.Object).To(gomega.Equal(once.Object))
}

func TestTemplateConflictsDoNotChangeResource(t *testing.T) {
	for _, tc := range []struct {
		name    string
		volumes []corev1.Volume
		mounts  []corev1.VolumeMount
	}{
		{name: "volume", volumes: []corev1.Volume{{Name: oauth2proxy.VolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}}},
		{name: "mount name", mounts: []corev1.VolumeMount{{Name: oauth2proxy.VolumeName, MountPath: "/other"}}},
		{name: "mount path", mounts: []corev1.VolumeMount{{Name: "user-volume", MountPath: oauth2proxy.TemplatesPath}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			container := proxy("runai-oauth2-proxy")
			container.VolumeMounts = tc.mounts
			resource, karta, _ := podResource(t, corev1.PodSpec{Containers: []corev1.Container{container}, Volumes: tc.volumes}, true)
			original := resource.DeepCopy()
			err := (&Plugin{configMapName: oauth2proxy.ConfigMapName}).MutateResource(context.Background(), nil, resource, karta, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.MatchError(gomega.ContainSubstring("conflict")))
			gomega.NewWithT(t).Expect(resource.Object).To(gomega.Equal(original.Object))
		})
	}
}

func TestFragmentedProxyReportsUnsupportedKarta(t *testing.T) {
	resource, karta, _ := podResource(t, corev1.PodSpec{Containers: []corev1.Container{proxy("runai-oauth2-proxy")}}, false)
	karta.Spec.StructureDefinition.RootComponent.SpecDefinition = &kartav1alpha1.SpecDefinition{
		FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{ContainersPath: ptr.To(".spec.containers")},
	}
	original := resource.DeepCopy()
	err := (&Plugin{configMapName: oauth2proxy.ConfigMapName}).MutateResource(context.Background(), nil, resource, karta, nil)
	gomega.NewWithT(t).Expect(err).To(gomega.MatchError(gomega.ContainSubstring("complete pod spec")))
	gomega.NewWithT(t).Expect(resource.Object).To(gomega.Equal(original.Object))
}

func TestExistingLegacyResourcesKeepTheirPodSpecs(t *testing.T) {
	for _, kind := range []string{"Job", "Service"} {
		t.Run(kind, func(t *testing.T) {
			g := gomega.NewWithT(t)
			container := proxy("runai-oauth2-proxy")
			container.Image = "internal.example/oauth2-proxy:runai-version"
			resource, karta, _ := podResource(t, corev1.PodSpec{Containers: []corev1.Container{container}}, true)
			if kind == "Service" {
				resource.SetAPIVersion("serving.knative.dev/v1")
				resource.SetKind(kind)
			}
			original := resource.DeepCopy()
			plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName, reader: fake.NewClientBuilder().WithObjects(original.DeepCopy()).Build()}
			g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
			g.Expect(resource.Object).To(gomega.Equal(original.Object), "controller upgrade must not alter existing unmounted proxies")
		})
	}
}

func TestNewResourceAndChangedProxyImageGetTemplates(t *testing.T) {
	for _, existing := range []bool{false, true} {
		name := "new resource"
		if existing {
			name = "changed proxy image"
		}
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			resource, karta, path := podResource(t, corev1.PodSpec{InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0")}}, true)
			builder := fake.NewClientBuilder()
			if existing {
				legacy := resource.DeepCopy()
				containers, _, err := unstructured.NestedSlice(legacy.Object, append(path, "initContainers")...)
				g.Expect(err).To(gomega.Succeed())
				containers[0].(map[string]any)["image"] = "internal.example/oauth2-proxy:old"
				g.Expect(unstructured.SetNestedSlice(legacy.Object, containers, append(path, "initContainers")...)).To(gomega.Succeed())
				builder = builder.WithObjects(legacy)
			}
			plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName, reader: builder.Build()}
			g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
			volumes, found, err := unstructured.NestedSlice(resource.Object, append(path, "volumes")...)
			g.Expect(err).To(gomega.Succeed())
			g.Expect(found).To(gomega.BeTrue())
			g.Expect(volumes).To(gomega.HaveLen(1))
		})
	}
}

func TestPreviouslyMountedResourceRetainsTemplates(t *testing.T) {
	g := gomega.NewWithT(t)
	resource, karta, _ := podResource(t, corev1.PodSpec{InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0")}}, true)
	live := resource.DeepCopy()
	plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName}
	g.Expect(plugin.MutateResource(context.Background(), nil, live, karta, nil)).To(gomega.Succeed())
	plugin.reader = fake.NewClientBuilder().WithObjects(live.DeepCopy()).Build()
	g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
	g.Expect(resource.Object).To(gomega.Equal(live.Object), "reconciliation must keep the controller-owned mounts")
}

func TestAdmissionInjectedContainersDoNotChangeLegacyProxyIdentity(t *testing.T) {
	g := gomega.NewWithT(t)
	pod := corev1.PodSpec{
		Containers:     []corev1.Container{proxy("runai-oauth2-proxy")},
		InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0")},
	}
	resource, karta, _ := podResource(t, pod, true)
	original := resource.DeepCopy()
	pod.Containers = append([]corev1.Container{{Name: "injected-sidecar", Image: "sidecar:latest"}}, pod.Containers...)
	pod.InitContainers = append([]corev1.Container{{Name: "injected-init", Image: "init:latest"}}, pod.InitContainers...)
	live, _, _ := podResource(t, pod, true)
	plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName, reader: fake.NewClientBuilder().WithObjects(live).Build()}
	g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.Succeed())
	g.Expect(resource.Object).To(gomega.Equal(original.Object), "unrelated injected containers must not trigger template mounts on legacy resources")
}

type failingReader struct {
	client.Reader
	err error
}

func (r failingReader) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return r.err
}

func TestReadFailureIsNotTreatedAsANewResource(t *testing.T) {
	g := gomega.NewWithT(t)
	resource, karta, _ := podResource(t, corev1.PodSpec{InitContainers: []corev1.Container{proxy("runai-oauth2-proxy-0")}}, true)
	original := resource.DeepCopy()
	readErr := errors.New("read denied")
	plugin := &Plugin{configMapName: oauth2proxy.ConfigMapName, reader: failingReader{err: readErr}}
	g.Expect(plugin.MutateResource(context.Background(), nil, resource, karta, nil)).To(gomega.MatchError(readErr))
	g.Expect(resource.Object).To(gomega.Equal(original.Object))

	unrelated, unrelatedKarta, _ := podResource(t, corev1.PodSpec{Containers: []corev1.Container{proxy("user-proxy")}}, true)
	g.Expect(plugin.MutateResource(context.Background(), nil, unrelated, unrelatedKarta, nil)).To(gomega.Succeed(), "unrelated resources do not require a live-resource read")
}
