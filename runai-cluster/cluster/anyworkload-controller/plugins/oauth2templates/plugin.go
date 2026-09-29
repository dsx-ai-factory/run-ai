/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package oauth2templates supplies the Run:ai error page to OAuth2 proxies
// composed by the control plane, including control planes predating the
// switch to the upstream proxy image.
package oauth2templates

import (
	"context"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The framework sorts plugins by name. Provision templates before networking
// creates standalone proxy Deployments.
const PluginName = "auth-proxy-templates"

type Plugin struct {
	configMapName string
	reader        client.Reader
}

func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	name := os.Getenv(oauth2proxy.TemplatesConfigMapEnvVar)
	if name == "" {
		return nil, framework.ErrPluginDisabled
	}
	return &Plugin{configMapName: name, reader: deps.Client}, nil
}

func (p *Plugin) Name() string { return PluginName }

func (p *Plugin) MutateResource(
	ctx context.Context,
	_ *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	_ *reconcile.State,
) error {
	if p.configMapName == "" {
		return nil
	}
	desired, err := proxyInventory(ctx, karta, resource)
	if err != nil {
		return err
	}
	if len(desired) == 0 {
		return nil
	}
	if p.reader != nil {
		live := &unstructured.Unstructured{}
		live.SetGroupVersionKind(resource.GroupVersionKind())
		err := p.reader.Get(ctx, client.ObjectKeyFromObject(resource), live)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("read existing resource for OAuth2 templates: %w", err)
		}
		if err == nil {
			existing, err := proxyInventory(ctx, karta, live)
			if err != nil {
				return err
			}
			// Existing proxies with the same images already provide their error
			// page. Avoid changing immutable Job/Pod specs or rolling Knative
			// revisions solely because the controller has been upgraded.
			if unchangedLegacyProxies(desired, existing) {
				return nil
			}
		}
	}
	// Leave unrelated resources untouched, and do not commit partial changes
	// when one component has a conflicting volume or mount.
	updated := resource.DeepCopy()
	err = kartautils.MutatePodInstances(ctx, karta, updated, func(instance kartaresource.ExtractedInstance) error {
		pod, err := completePodSpec(instance)
		if err != nil || pod == nil {
			return err
		}
		if !slices.ContainsFunc(pod.Containers, isRunaiProxy) && !slices.ContainsFunc(pod.InitContainers, isRunaiProxy) {
			return nil
		}
		return mountTemplates(pod, p.configMapName)
	})
	if err != nil {
		return fmt.Errorf("mount OAuth2 proxy templates: %w", err)
	}
	resource.Object = updated.Object
	return nil
}

type proxyIdentity struct {
	component string
	instance  string
	init      bool
	name      string
}

type proxySpec struct {
	image     string
	templates bool
}

func proxyInventory(ctx context.Context, karta *kartav1alpha1.Karta, resource *unstructured.Unstructured) (map[proxyIdentity]proxySpec, error) {
	components, err := kartautils.GetAllComponents(kartaresource.NewComponentFactoryFromObject(karta, resource))
	if err != nil {
		return nil, err
	}
	result := map[proxyIdentity]proxySpec{}
	for _, component := range components {
		instances, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return nil, fmt.Errorf("extract OAuth2 proxy instances: %w", err)
		}
		for id, instance := range instances {
			pod, err := completePodSpec(instance)
			if err != nil {
				return nil, err
			}
			if pod == nil {
				continue
			}
			for kind, containers := range [][]corev1.Container{pod.Containers, pod.InitContainers} {
				for _, container := range containers {
					if isRunaiProxy(container) {
						mounted := slices.ContainsFunc(container.VolumeMounts, func(mount corev1.VolumeMount) bool {
							return mount.Name == oauth2proxy.VolumeName && mount.MountPath == oauth2proxy.TemplatesPath
						})
						result[proxyIdentity{component: component.Name(), instance: id, init: kind == 1, name: container.Name}] = proxySpec{
							image: container.Image, templates: mounted,
						}
					}
				}
			}
		}
	}
	return result, nil
}

func unchangedLegacyProxies(desired, existing map[proxyIdentity]proxySpec) bool {
	return maps.EqualFunc(desired, existing, func(proxy, live proxySpec) bool {
		return proxy.image == live.image && !live.templates
	})
}

func completePodSpec(instance kartaresource.ExtractedInstance) (*corev1.PodSpec, error) {
	switch {
	case instance.PodTemplateSpec != nil:
		return &instance.PodTemplateSpec.Spec, nil
	case instance.PodSpec != nil:
		return instance.PodSpec, nil
	case instance.FragmentedPodSpec != nil:
		fragmented := instance.FragmentedPodSpec
		if slices.ContainsFunc(fragmented.Containers, isRunaiProxy) || (fragmented.Container != nil && isRunaiProxy(*fragmented.Container)) {
			return nil, fmt.Errorf("OAuth2 templates require a Karta with a complete pod spec or pod template")
		}
	}
	return nil, nil
}

func isRunaiProxy(container corev1.Container) bool {
	const prefix = "runai-oauth2-proxy"
	if container.Name != prefix {
		suffix, ok := strings.CutPrefix(container.Name, prefix+"-")
		if !ok || suffix == "" {
			return false
		}
		for _, digit := range suffix {
			if digit < '0' || digit > '9' {
				return false
			}
		}
	}
	return slices.Contains(container.Args, "--custom-templates-dir="+oauth2proxy.TemplatesPath)
}

func mountTemplates(pod *corev1.PodSpec, configMapName string) error {
	volume := corev1.Volume{
		Name: oauth2proxy.VolumeName,
		VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
			Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
		}},
	}
	if slices.ContainsFunc(pod.Volumes, func(existing corev1.Volume) bool {
		return existing.Name == volume.Name && !reflect.DeepEqual(existing, volume)
	}) {
		return fmt.Errorf("volume %q conflicts with OAuth2 templates", volume.Name)
	}
	if !slices.ContainsFunc(pod.Volumes, func(existing corev1.Volume) bool { return existing.Name == volume.Name }) {
		pod.Volumes = append(pod.Volumes, volume)
	}

	mount := corev1.VolumeMount{Name: oauth2proxy.VolumeName, MountPath: oauth2proxy.TemplatesPath, ReadOnly: true}
	for _, containers := range [][]corev1.Container{pod.Containers, pod.InitContainers} {
		for i := range containers {
			container := &containers[i]
			if !isRunaiProxy(*container) {
				continue
			}
			if slices.ContainsFunc(container.VolumeMounts, func(existing corev1.VolumeMount) bool {
				return (existing.Name == mount.Name || existing.MountPath == mount.MountPath) && existing != mount
			}) {
				return fmt.Errorf("container %q has a conflicting OAuth2 templates mount", container.Name)
			}
			if !slices.Contains(container.VolumeMounts, mount) {
				container.VolumeMounts = append(container.VolumeMounts, mount)
			}
		}
	}
	return nil
}

var _ framework.ResourceMutatePlugin = &Plugin{}
