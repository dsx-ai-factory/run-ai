/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AdditionalWorkloadGvksReadClusterRoleLabel = "run.ai/additional-workload-gvks-read"
	AdditionalWorkloadGvksCrudClusterRoleLabel = "run.ai/additional-workload-gvks-crud"
)

type KartaRBACHandler struct {
	Client client.Client

	CachedAdditionalGvksApi CachedAdditionalGvksApi
	RBACUpdaterEnabled      bool
	ClusterRoleReadName     string
	ClusterRoleCrudName     string
}

func NewKartaRBACHandler(client client.Client,
	cachedAdditionalGvksApi CachedAdditionalGvksApi,
	rbacUpdaterEnabled bool, clusterRoleReadName, clusterRoleCrudName string) *KartaRBACHandler {
	return &KartaRBACHandler{
		Client:                  client,
		CachedAdditionalGvksApi: cachedAdditionalGvksApi,
		RBACUpdaterEnabled:      rbacUpdaterEnabled,
		ClusterRoleReadName:     clusterRoleReadName,
		ClusterRoleCrudName:     clusterRoleCrudName,
	}
}

func (h *KartaRBACHandler) Handle(ctx context.Context, gvk schema.GroupVersionKind, additionalGvks []schema.GroupVersionKind) error {
	allGvks := append([]schema.GroupVersionKind{gvk}, additionalGvks...)

	names, resolveErr := h.resolveResourceNames(allGvks)
	if resolveErr != nil {
		return resolveErr
	}

	var err error

	innerErr := h.handleClusterRole(ctx, allGvks, names, h.ClusterRoleReadName, false)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = h.handleClusterRole(ctx, allGvks, names, h.ClusterRoleCrudName, true)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		return err
	}

	for _, additionalGvk := range additionalGvks {
		h.CachedAdditionalGvksApi.AddAdditionalGroupKind(additionalGvk.GroupKind())
	}

	log.Debug().Msgf("Successfully handled RBAC for %s", gvk.String())
	return nil
}

func (h *KartaRBACHandler) HandleDeletion(ctx context.Context, gvk schema.GroupVersionKind, additionalGvks []schema.GroupVersionKind) error {
	allGvks := append([]schema.GroupVersionKind{gvk}, additionalGvks...)
	targets := h.removalTargets(allGvks)
	var err error

	innerErr := h.handleClusterRoleOnDeletion(ctx, targets, h.ClusterRoleReadName, false)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = h.handleClusterRoleOnDeletion(ctx, targets, h.ClusterRoleCrudName, true)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		return err
	}

	for _, additionalGvk := range additionalGvks {
		h.CachedAdditionalGvksApi.RemoveAdditionalGroupKind(additionalGvk.GroupKind())
	}

	log.Debug().Msgf("Successfully handled RBAC deletion for %s", gvk.String())
	return nil
}

func (h *KartaRBACHandler) handleClusterRole(ctx context.Context, gvks []schema.GroupVersionKind, names resourceNames, clusterRoleName string, crudRole bool) error {
	if len(gvks) == 0 {
		return nil
	}

	permissionsMsg := "read"
	if crudRole {
		permissionsMsg = "CRUD"
	}

	if clusterRoleName == "" {
		log.Warn().Msgf("No cluster role defined for RBAC %s handling for %s", permissionsMsg, gvks[0].String())
		return nil
	}

	cr, err := h.getClusterRole(ctx, clusterRoleName)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get cluster role '%s': %w", clusterRoleName, err)
	}
	if err == nil && cr != nil && hasGvksRbacInClusterRole(gvks, cr, crudRole, names.of) {
		log.Debug().Msgf("%s permissions RBAC for %s (and %d additional gvks) already exist in cluster role '%s'",
			permissionsMsg, gvks[0].String(), len(gvks)-1, clusterRoleName)
		return nil
	}

	if !h.RBACUpdaterEnabled {
		return fmt.Errorf("RBAC handler is disabled, but missing %s permissions RBAC for <%s> (or more GVKs) in cluster role '%s'",
			permissionsMsg, gvks[0].String(), clusterRoleName)
	}

	if cr == nil {
		// means the ClusterRole was not found - need to create it
		cr, err = h.createClusterRole(ctx, clusterRoleName, crudRole)
		if err != nil {
			return fmt.Errorf("failed to create cluster role '%s': %w", clusterRoleName, err)
		}
	}

	err = h.addRuleToClusterRole(ctx, cr, gvks, names, crudRole, permissionsMsg)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully added %s permissions RBAC for <%s> (and %d additional gvks) to cluster role %s",
		permissionsMsg, gvks[0].String(), len(gvks)-1, cr.Name)
	return nil
}

func (h *KartaRBACHandler) addRuleToClusterRole(ctx context.Context, cr *rbacv1.ClusterRole, gvks []schema.GroupVersionKind, names resourceNames, crudRole bool, permissionsMsg string) error {
	verbs := []string{"get", "list", "watch"}
	if crudRole {
		verbs = append(verbs, "create", "update", "delete", "patch")
	}
	groups := datastructs.Set[string]{}
	resources := []string{}
	for _, gvk := range gvks {
		groups.Add(gvk.Group)
		resources = append(resources, names.of(gvk))
	}

	apiGroups := []string{}
	for g := range groups {
		apiGroups = append(apiGroups, g)
	}

	// add only 1 rule for all GVKs
	rule := rbacv1.PolicyRule{
		APIGroups: apiGroups,
		Resources: resources,
		Verbs:     verbs,
	}
	cr.Rules = append(cr.Rules, rule)

	err := h.Client.Update(ctx, cr)
	if err != nil {
		return fmt.Errorf("failed to update cluster role '%s' with %s RBAC for <%s> (and %d additional gvks), error: %s",
			cr.Name, permissionsMsg, gvks[0].String(), len(gvks)-1, err.Error())
	}

	return nil
}

func (h *KartaRBACHandler) getClusterRole(ctx context.Context, name string) (*rbacv1.ClusterRole, error) {
	cr := &rbacv1.ClusterRole{}
	err := h.Client.Get(ctx, client.ObjectKey{Name: name}, cr)
	if err != nil {
		return nil, err
	}
	return cr, nil
}

func (h *KartaRBACHandler) createClusterRole(ctx context.Context, name string, crudRole bool) (*rbacv1.ClusterRole, error) {
	label := AdditionalWorkloadGvksReadClusterRoleLabel
	if crudRole {
		label = AdditionalWorkloadGvksCrudClusterRoleLabel
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				label: "true",
			},
		},
	}
	err := h.Client.Create(ctx, cr)
	return cr, err
}

func (h *KartaRBACHandler) handleClusterRoleOnDeletion(ctx context.Context, targets []removalTarget, clusterRoleName string, crudRole bool) error {
	if len(targets) == 0 {
		return nil
	}
	main := targets[0].gvk.String()

	permissionsMsg := "read"
	if crudRole {
		permissionsMsg = "CRUD"
	}

	if clusterRoleName == "" {
		log.Warn().Msgf("No cluster role defined for RBAC %s deletion handling for %s", permissionsMsg, main)
		return nil
	}

	cr, err := h.getClusterRole(ctx, clusterRoleName)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Msgf("Cluster role '%s' not found on deletion handling for %s, nothing to do", clusterRoleName, main)
			return nil
		}
		return fmt.Errorf("failed to get cluster role '%s': %w", clusterRoleName, err)
	}

	if !hasRuleToRemove(cr, targets, crudRole) {
		if unresolved := unresolvedGvks(targets); len(unresolved) > 0 {
			// the API server no longer serves these kinds (for example the CRD was deleted before the type), so
			// the real resource name is unknown and an irregular plural cannot be matched by its Kind
			log.Warn().Msgf("Could not find %s permissions RBAC for %s (and %d additional gvks) in cluster role '%s', "+
				"and the resource name of %v could not be resolved from the API server. "+
				"A rule with an irregular resource name may remain",
				permissionsMsg, main, len(targets)-1, clusterRoleName, unresolved)
			return nil
		}
		log.Debug().Msgf("%s permissions RBAC for %s (and %d additional gvks) are already removed from cluster role '%s'",
			permissionsMsg, main, len(targets)-1, clusterRoleName)
		return nil
	}

	if !h.RBACUpdaterEnabled {
		// no error in this case on removal
		return nil
	}

	err = h.removeRuleFromClusterRole(ctx, cr, targets, crudRole, permissionsMsg)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully removed %s permissions RBAC for <%s> (and %d additional gvks) from cluster role %s",
		permissionsMsg, main, len(targets)-1, cr.Name)
	return nil
}

func (h *KartaRBACHandler) removeRuleFromClusterRole(ctx context.Context, cr *rbacv1.ClusterRole, targets []removalTarget, crudRole bool, permissionsMsg string) error {
	newRules := []rbacv1.PolicyRule{}
	for _, rule := range cr.Rules {
		if !HasAllVerbs(rule, crudRole) || !ruleCoversTargets(rule, targets) {
			newRules = append(newRules, rule)
			continue
		}
		newRules = append(newRules, ruleWithoutTargets(rule, targets)...)
	}

	cr.Rules = newRules
	if err := h.Client.Update(ctx, cr); err != nil {
		return fmt.Errorf("failed to update cluster role '%s' with removal of %s RBAC for <%s> (and %d additional gvks), error: %w",
			cr.Name, permissionsMsg, targets[0].gvk.String(), len(targets)-1, err)
	}

	return nil
}

// removalTarget is a GVK together with every resource name that RBAC rules for it may use.
type removalTarget struct {
	gvk schema.GroupVersionKind
	// names are the name the API server serves (when it can be resolved), the guessed name, and the Kind + "s"
	// name that earlier versions generated. Removal matches any of them, because the API server may no longer
	// serve a type that is being deleted, and because rules written by earlier versions can carry the wrong name.
	names []string
	// served is whether the API server resolves the resource name of the GVK.
	served bool
}

// removalTargets resolves the candidate resource names of each GVK once, for all the matching that follows.
func (h *KartaRBACHandler) removalTargets(gvks []schema.GroupVersionKind) []removalTarget {
	targets := make([]removalTarget, 0, len(gvks))
	for _, gvk := range gvks {
		target := removalTarget{gvk: gvk}
		served, ok, err := h.servedResourceName(gvk)
		if err != nil {
			// deletion does not wait for the API server: the guessed names below still match the rules
			log.Warn().Msgf("%v, matching the rules of <%s> by its guessed names", err, gvk.String())
		}
		if ok {
			target.names = append(target.names, served)
			target.served = true
		}
		for _, name := range []string{resourceNameFromGVK(gvk), strings.ToLower(gvk.Kind) + "s"} {
			if !slices.Contains(target.names, name) {
				target.names = append(target.names, name)
			}
		}
		targets = append(targets, target)
	}
	return targets
}

// hasRuleToRemove reports whether the cluster role has a rule that removeRuleFromClusterRole would change.
func hasRuleToRemove(cr *rbacv1.ClusterRole, targets []removalTarget, crudRole bool) bool {
	for _, rule := range cr.Rules {
		if HasAllVerbs(rule, crudRole) && ruleCoversTargets(rule, targets) {
			return true
		}
	}
	return false
}

// ruleCoversTargets reports whether the rule names every target, under any of the resource names it may have been written with.
func ruleCoversTargets(rule rbacv1.PolicyRule, targets []removalTarget) bool {
	for _, target := range targets {
		if !slices.Contains(rule.APIGroups, target.gvk.Group) || !slices.ContainsFunc(target.names, func(name string) bool {
			return slices.Contains(rule.Resources, name)
		}) {
			return false
		}
	}
	return true
}

// ruleWithoutTargets returns the rule without the resources of the targets. A rule grants the cross product of its
// API groups and resources, so the resources are filtered for each API group separately: a resource name that
// belongs to a removed type in one group is kept for the other groups. Groups that keep the same resources stay
// together in one rule, and the rule disappears when nothing is left.
func ruleWithoutTargets(rule rbacv1.PolicyRule, targets []removalTarget) []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{}
	for _, group := range rule.APIGroups {
		remaining := slices.DeleteFunc(slices.Clone(rule.Resources), func(resource string) bool {
			return isTargetResource(targets, group, resource)
		})
		if len(remaining) == 0 {
			continue
		}

		merged := false
		for i := range rules {
			if slices.Equal(rules[i].Resources, remaining) {
				rules[i].APIGroups = append(rules[i].APIGroups, group)
				merged = true
				break
			}
		}
		if !merged {
			kept := *rule.DeepCopy()
			kept.APIGroups = []string{group}
			kept.Resources = remaining
			rules = append(rules, kept)
		}
	}
	return rules
}

func isTargetResource(targets []removalTarget, group, resource string) bool {
	return slices.ContainsFunc(targets, func(target removalTarget) bool {
		return target.gvk.Group == group && slices.Contains(target.names, resource)
	})
}

// unresolvedGvks returns the kinds whose resource name cannot be resolved from the API server.
func unresolvedGvks(targets []removalTarget) []string {
	unresolved := []string{}
	for _, target := range targets {
		if !target.served {
			unresolved = append(unresolved, target.gvk.String())
		}
	}
	return unresolved
}

// resourceNamer returns the plural resource name that RBAC rules use for a GVK.
type resourceNamer func(gvk schema.GroupVersionKind) string

func hasGvksRbacInClusterRole(gvks []schema.GroupVersionKind, cr *rbacv1.ClusterRole, hasCrud bool, nameOf resourceNamer) bool {
	if len(gvks) == 0 {
		return true
	}

	// Build a fast lookup of pairs (apiGroup, resource) that have all required verbs
	for _, rule := range cr.Rules {
		allowed := datastructs.Set[string]{}
		if !HasAllVerbs(rule, hasCrud) {
			continue
		}
		for _, apiGroup := range rule.APIGroups {
			for _, resource := range rule.Resources {
				key := apiGroup + "|" + resource
				allowed.Add(key)
			}
		}

		// check if all gvks are allowed in this rule
		allFound := true
		for _, gvk := range gvks {
			resource := nameOf(gvk)
			key := gvk.Group + "|" + resource
			if !allowed.Has(key) {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}

	return false
}

func HasAllVerbs(rule rbacv1.PolicyRule, hasCrud bool) bool {
	foundGet := false
	foundList := false
	foundWatch := false

	foundCreate := false
	foundUpdate := false
	foundDelete := false
	foundPatch := false

	for _, verb := range rule.Verbs {
		switch verb {
		case "get":
			foundGet = true
		case "list":
			foundList = true
		case "watch":
			foundWatch = true
		case "create":
			foundCreate = true
		case "update":
			foundUpdate = true
		case "delete":
			foundDelete = true
		case "patch":
			foundPatch = true
		}
	}

	if hasCrud {
		return foundCreate && foundUpdate && foundDelete && foundPatch && foundGet && foundList && foundWatch
	}
	return foundGet && foundList && foundWatch
}

// resourceNames holds the resource name that RBAC rules use for each GVK of one Handle call.
type resourceNames map[schema.GroupVersionKind]string

func (n resourceNames) of(gvk schema.GroupVersionKind) string { return n[gvk] }

// resolveResourceNames returns the resource name of each GVK, which is what RBAC rules must reference. The plural of
// a Kind is defined by its CRD and is not always Kind + "s" (for example Sandbox is served as "sandboxes"), so it is
// read from the REST mapper, once per GVK. A GVK that the API server does not know yet, for example because its CRD
// is installed after the type is registered, gets a name guessed from its Kind. Any other failure of the REST mapper,
// such as a discovery timeout, is returned so that the reconcile is retried: a guess written during such a failure
// can be wrong, and nothing would correct it.
func (h *KartaRBACHandler) resolveResourceNames(gvks []schema.GroupVersionKind) (resourceNames, error) {
	names := make(resourceNames, len(gvks))
	var err error
	for _, gvk := range gvks {
		if _, resolved := names[gvk]; resolved {
			continue
		}
		served, ok, resolveErr := h.servedResourceName(gvk)
		if resolveErr != nil {
			err = multierror.Append(err, resolveErr)
			continue
		}
		if !ok {
			log.Warn().Msgf("Could not resolve the resource name of <%s> from the API server, guessing it from the kind", gvk.String())
			served = resourceNameFromGVK(gvk)
		}
		names[gvk] = served
	}
	if err != nil {
		return nil, err
	}
	return names, nil
}

// servedResourceName returns the resource name the API server serves for the GVK. The bool is false, without an
// error, when the API server does not know the GVK. Any other failure of the REST mapper is returned.
func (h *KartaRBACHandler) servedResourceName(gvk schema.GroupVersionKind) (string, bool, error) {
	mapper := h.Client.RESTMapper()
	if mapper == nil {
		return "", false, nil
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if meta.IsNoMatchError(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve the resource name of <%s> from the API server: %w", gvk.String(), err)
	}
	return mapping.Resource.Resource, true, nil
}

// resourceNameFromGVK guesses the plural resource name of a Kind using the usual English rules. It is only a
// fallback for when the REST mapper cannot resolve the real name: irregular plurals are not covered.
func resourceNameFromGVK(gvk schema.GroupVersionKind) string {
	kind := strings.ToLower(gvk.Kind)
	switch {
	case strings.HasSuffix(kind, "s"), strings.HasSuffix(kind, "x"), strings.HasSuffix(kind, "ch"), strings.HasSuffix(kind, "sh"):
		return kind + "es"
	case len(kind) > 1 && strings.HasSuffix(kind, "y") && !strings.ContainsRune("aeiou", rune(kind[len(kind)-2])):
		return strings.TrimSuffix(kind, "y") + "ies"
	default:
		return kind + "s"
	}
}
