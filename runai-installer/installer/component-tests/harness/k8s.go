/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"
	"fmt"
	"strings"
	"sync"

	k8slabels "k8s.io/apimachinery/pkg/labels"
)

// FakeK8sEnv seeds the initial state of a FakeK8sWrapper.
type FakeK8sEnv struct {
	Namespaces                      []string
	CRDs                            []string
	MutatingWebhookConfigurations   map[string]map[string]string // name -> labels
	ValidatingWebhookConfigurations map[string]map[string]string // name -> labels
	// WebhookServiceNamespaces maps a webhook configuration name to the
	// namespaces its clientConfig.service entries point at.
	WebhookServiceNamespaces map[string][]string
	ResourceInterfaces       map[string]map[string]string // name -> labels
	ClusterRoles             map[string]map[string]string // name -> labels
	PriorityClasses          map[string]map[string]string // name -> labels
	// RunAICRFinalizersByCRD seeds CRs that carry finalizers, keyed by CRD name.
	// e.g. {"nodepools.run.ai": {"default", "gpu-pool"}}
	RunAICRFinalizersByCRD map[string][]string
}

// FakeK8sWrapper is an in-memory implementation of uninstall.K8sWrapper.
type FakeK8sWrapper struct {
	mu                              sync.Mutex
	namespaces                      map[string]struct{}
	crds                            map[string]struct{}
	mutatingWebhookConfigurations   map[string]map[string]string
	validatingWebhookConfigurations map[string]map[string]string
	webhookServiceNamespaces        map[string][]string
	resourceInterfaces              map[string]map[string]string
	clusterRoles                    map[string]map[string]string
	priorityClasses                 map[string]map[string]string
	// crsWithFinalizers tracks CRs that still carry finalizers, keyed by CRD name.
	crsWithFinalizers map[string]map[string]struct{}
	// callLog records K8sWrapper method calls in arrival order so tests can
	// assert on relative ordering (e.g. strip-finalizers happens before delete-CRD).
	callLog []string
}

func NewFakeK8s(env FakeK8sEnv) *FakeK8sWrapper {
	namespaces := make(map[string]struct{}, len(env.Namespaces))
	for _, n := range env.Namespaces {
		namespaces[n] = struct{}{}
	}
	crds := make(map[string]struct{}, len(env.CRDs))
	for _, c := range env.CRDs {
		crds[c] = struct{}{}
	}
	crsWithFinalizers := make(map[string]map[string]struct{}, len(env.RunAICRFinalizersByCRD))
	for crdName, crNames := range env.RunAICRFinalizersByCRD {
		set := make(map[string]struct{}, len(crNames))
		for _, name := range crNames {
			set[name] = struct{}{}
		}
		crsWithFinalizers[crdName] = set
	}

	return &FakeK8sWrapper{
		namespaces:                      namespaces,
		crds:                            crds,
		mutatingWebhookConfigurations:   cloneLabels(env.MutatingWebhookConfigurations),
		validatingWebhookConfigurations: cloneLabels(env.ValidatingWebhookConfigurations),
		webhookServiceNamespaces:        env.WebhookServiceNamespaces,
		resourceInterfaces:              cloneLabels(env.ResourceInterfaces),
		clusterRoles:                    cloneLabels(env.ClusterRoles),
		priorityClasses:                 cloneLabels(env.PriorityClasses),
		crsWithFinalizers:               crsWithFinalizers,
	}
}

func cloneLabels(src map[string]map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// ListMutatingWebhookConfigurations uses the real k8s label-selector parser so
// that tests exercise the same selector syntax that the production code passes
// to the Kubernetes API.  This catches bugs like a malformed selector string
// before they reach a real cluster.
func (f *FakeK8sWrapper) ListMutatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	selector, err := k8slabels.Parse(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector %q: %w", labelSelector, err)
	}
	var names []string
	for name, labels := range f.mutatingWebhookConfigurations {
		if selector.Matches(k8slabels.Set(labels)) || f.webhookRoutesTo(name, serviceNamespaces) {
			names = append(names, name)
		}
	}
	return names, nil
}

// webhookRoutesTo mirrors the production ownership fallback.
func (f *FakeK8sWrapper) webhookRoutesTo(name string, serviceNamespaces []string) bool {
	for _, hookNS := range f.webhookServiceNamespaces[name] {
		for _, ns := range serviceNamespaces {
			if hookNS == ns {
				return true
			}
		}
	}
	return false
}

func (f *FakeK8sWrapper) DeleteMutatingWebhookConfiguration(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.mutatingWebhookConfigurations, name)
	return nil
}

func (f *FakeK8sWrapper) MutatingWebhookConfigurationExists(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, exists := f.mutatingWebhookConfigurations[name]
	return exists
}

// ListValidatingWebhookConfigurations — same rationale as ListMutatingWebhookConfigurations above.
func (f *FakeK8sWrapper) ListValidatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	selector, err := k8slabels.Parse(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector %q: %w", labelSelector, err)
	}
	var names []string
	for name, labels := range f.validatingWebhookConfigurations {
		if selector.Matches(k8slabels.Set(labels)) || f.webhookRoutesTo(name, serviceNamespaces) {
			names = append(names, name)
		}
	}
	return names, nil
}

func (f *FakeK8sWrapper) DeleteValidatingWebhookConfiguration(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.validatingWebhookConfigurations, name)
	return nil
}

func (f *FakeK8sWrapper) ValidatingWebhookConfigurationExists(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, exists := f.validatingWebhookConfigurations[name]
	return exists
}

// ListResourceInterfaces filters by label key prefix client-side, mirroring
// the production implementation.  Using prefix matching (rather than a
// specific label key) ensures tests will catch any Run:AI-owned ResourceInterface
// that the uninstall step failed to delete, even if the label key names change.
// Returns nil, nil when the ResourceInterface CRD is not registered, mirroring
// the real implementation's behaviour when the optimization.nvidia.com CRD is absent.
func (f *FakeK8sWrapper) ListResourceInterfaces(labelKeyPrefix string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.crds["resourceinterfaces.optimization.nvidia.com"]; !ok {
		return nil, nil
	}
	var names []string
	for name, labels := range f.resourceInterfaces {
		for key := range labels {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, name)
				break
			}
		}
	}
	return names, nil
}

func (f *FakeK8sWrapper) DeleteResourceInterface(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.resourceInterfaces, name)
	return nil
}

// ListClusterRoles — same label key prefix rationale as ListResourceInterfaces.
func (f *FakeK8sWrapper) ListClusterRoles(labelKeyPrefix string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var names []string
	for name, labels := range f.clusterRoles {
		for key := range labels {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, name)
				break
			}
		}
	}
	return names, nil
}

func (f *FakeK8sWrapper) DeleteClusterRole(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.clusterRoles, name)
	return nil
}

// ListPriorityClasses — same label key prefix rationale as ListClusterRoles.
func (f *FakeK8sWrapper) ListPriorityClasses(labelKeyPrefix string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var names []string
	for name, labels := range f.priorityClasses {
		for key := range labels {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, name)
				break
			}
		}
	}
	return names, nil
}

func (f *FakeK8sWrapper) DeletePriorityClass(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.priorityClasses, name)
	return nil
}

func (f *FakeK8sWrapper) ForceDeleteNamespace(namespace string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.namespaces, namespace)
	return nil
}

func (f *FakeK8sWrapper) NamespaceExists(namespace string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, exists := f.namespaces[namespace]
	return exists, nil
}

func (f *FakeK8sWrapper) ListCRDs() ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	names := make([]string, 0, len(f.crds))
	for name := range f.crds {
		names = append(names, name)
	}
	return names, nil
}

func (f *FakeK8sWrapper) DeleteCRD(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, "DeleteCRD:"+name)
	delete(f.crds, name)
	return nil
}

func (f *FakeK8sWrapper) StripRunAICRFinalizers(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, "StripRunAICRFinalizers")
	f.crsWithFinalizers = make(map[string]map[string]struct{})
	return nil
}

// CRsWithFinalizers returns the names of CRs that still carry finalizers for
// the given CRD. Used in test assertions.
func (f *FakeK8sWrapper) CRsWithFinalizers(crdName string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var names []string
	for name := range f.crsWithFinalizers[crdName] {
		names = append(names, name)
	}
	return names
}

// FirstCallIndex returns the index of the first entry in the call log whose
// string starts with prefix, or -1 if none match.
func (f *FakeK8sWrapper) FirstCallIndex(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, call := range f.callLog {
		if strings.HasPrefix(call, prefix) {
			return i
		}
	}
	return -1
}
