package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	var err error

	innerErr := h.handleClusterRole(ctx, allGvks, h.ClusterRoleReadName, false)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = h.handleClusterRole(ctx, allGvks, h.ClusterRoleCrudName, true)
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
	var err error

	innerErr := h.handleClusterRoleOnDeletion(ctx, allGvks, h.ClusterRoleReadName, false)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = h.handleClusterRoleOnDeletion(ctx, allGvks, h.ClusterRoleCrudName, true)
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

func (h *KartaRBACHandler) handleClusterRole(ctx context.Context, gvks []schema.GroupVersionKind, clusterRoleName string, crudRole bool) error {
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
	if err == nil && cr != nil && hasGvksRbacInClusterRole(gvks, cr, crudRole) {
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

	err = h.addRuleToClusterRole(ctx, cr, gvks, crudRole, permissionsMsg)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully added %s permissions RBAC for <%s> (and %d additional gvks) to cluster role %s",
		permissionsMsg, gvks[0].String(), len(gvks)-1, cr.Name)
	return nil
}

func (h *KartaRBACHandler) addRuleToClusterRole(ctx context.Context, cr *rbacv1.ClusterRole, gvks []schema.GroupVersionKind, crudRole bool, permissionsMsg string) error {
	verbs := []string{"get", "list", "watch"}
	if crudRole {
		verbs = append(verbs, "create", "update", "delete", "patch")
	}
	groups := datastructs.Set[string]{}
	resources := []string{}
	for _, gvk := range gvks {
		groups.Add(gvk.Group)
		resources = append(resources, resourceNameFromGVK(gvk))
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

func (h *KartaRBACHandler) handleClusterRoleOnDeletion(ctx context.Context, gvks []schema.GroupVersionKind, clusterRoleName string, crudRole bool) error {
	if len(gvks) == 0 {
		return nil
	}

	permissionsMsg := "read"
	if crudRole {
		permissionsMsg = "CRUD"
	}

	if clusterRoleName == "" {
		log.Warn().Msgf("No cluster role defined for RBAC %s deletion handling for %s", permissionsMsg, gvks[0].String())
		return nil
	}

	cr, err := h.getClusterRole(ctx, clusterRoleName)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Msgf("Cluster role '%s' not found on deletion handling for %s, nothing to do", clusterRoleName, gvks[0].String())
			return nil
		}
		return fmt.Errorf("failed to get cluster role '%s': %w", clusterRoleName, err)
	}

	if !hasGvksRbacInClusterRole(gvks, cr, crudRole) {
		log.Debug().Msgf("%s permissions RBAC for %s (and %d additional gvks) are already removed from cluster role '%s'",
			permissionsMsg, gvks[0].String(), len(gvks)-1, clusterRoleName)
		return nil
	}

	if !h.RBACUpdaterEnabled {
		// no error in this case on removal
		return nil
	}

	err = h.removeRuleFromClusterRole(ctx, cr, gvks, crudRole, permissionsMsg)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully removed %s permissions RBAC for <%s> (and %d additional gvks) from cluster role %s", permissionsMsg, gvks[0].String(), len(gvks)-1, cr.Name)
	return nil
}

func (h *KartaRBACHandler) removeRuleFromClusterRole(ctx context.Context, cr *rbacv1.ClusterRole, gvks []schema.GroupVersionKind, crudRole bool, permissionsMsg string) error {
	// Build target set of (apiGroup|resource) pairs to remove
	targetsToRemove := datastructs.Set[string]{}
	for _, g := range gvks {
		key := g.Group + "|" + resourceNameFromGVK(g)
		targetsToRemove.Add(key)
	}

	newRules := []rbacv1.PolicyRule{}
	// search for a rule that has ALL the verbs we need
	for _, rule := range cr.Rules {
		if !HasAllVerbs(rule, crudRole) {
			newRules = append(newRules, rule)
			continue
		}

		targetsInCurrentRule := datastructs.Set[string]{}
		for _, apiGroup := range rule.APIGroups {
			for _, resource := range rule.Resources {
				key := apiGroup + "|" + resource
				if targetsToRemove.Has(key) {
					targetsInCurrentRule.Add(key)
				}
			}
		}

		// will remove the rule if it has ALL the targets we want to remove

		if len(targetsInCurrentRule) != len(targetsToRemove) {
			newRules = append(newRules, rule)
		}
	}

	cr.Rules = newRules
	if err := h.Client.Update(ctx, cr); err != nil {
		return fmt.Errorf("failed to update cluster role '%s' with removal of %s RBAC for <%s> (and %d additional gvks), error: %w",
			cr.Name, permissionsMsg, gvks[0].String(), len(gvks)-1, err)
	}

	return nil
}

func hasGvksRbacInClusterRole(gvks []schema.GroupVersionKind, cr *rbacv1.ClusterRole, hasCrud bool) bool {
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
			resource := resourceNameFromGVK(gvk)
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

func resourceNameFromGVK(gvk schema.GroupVersionKind) string {
	return strings.ToLower(gvk.Kind) + "s"
}
