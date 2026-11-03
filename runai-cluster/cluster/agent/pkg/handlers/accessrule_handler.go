/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"k8s.io/utils/ptr"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	auth_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxNameLen         = 62
	nameSuffixIfNeeded = "a"
)

type AccessRulesHandler struct {
	*CommonHandler

	lastUpdated *time.Time
}

type AccessRuleObjectInspector bool

type AccessRulesExtractionResult struct {
	clusterAccessRules    []client.Object
	accessRuleIdsToDelete []string
	lastUpdated           *time.Time
}

func NewAccessRulesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&v1alpha1.AccessRule{}, AccessRuleObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &AccessRulesHandler{
		CommonHandler: commonHandler,
		lastUpdated:   nil,
	}, nil
}

func (h *AccessRulesHandler) Handle() error {
	log.Debug().Msgf("Starting to sync accessrules")

	backendAccessRules, err := h.getBackendAccessRulesSinceLastUpdated()
	if err != nil {
		return err
	}

	if len(backendAccessRules) == 0 {
		return nil
	}

	accessRulesExtractionResult, err := h.extractAccessRules(backendAccessRules)

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(accessRulesExtractionResult.clusterAccessRules))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(accessRulesExtractionResult.accessRuleIdsToDelete))

	// don't sync deleted - because we are getting the deleted objects from the backend and deleting them by id
	_, _, innerErr := h.CreateOrUpdateObjects(accessRulesExtractionResult.clusterAccessRules)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	// Delete by the runai/accessrule-id label rather than by the generated name: the name embeds the
	// role name, so after a role rename the object on the cluster no longer carries the name we would
	// generate now, and a name-keyed delete can never reach it.
	innerErr = h.deleteAccessRulesByIds(accessRulesExtractionResult.accessRuleIdsToDelete)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		log.Error().Msgf("Failed to handle AccessRules, error: %s", err.Error())
		return err
	}

	h.lastUpdated = accessRulesExtractionResult.lastUpdated

	log.Debug().Msgf("Successfully handled AccessRules")

	return nil
}

// listAccessRulesByIds returns the AccessRules carrying one of the given runai/accessrule-id values,
// in a single List using a set-based "in" selector.
//
// One List per id would be quadratic: a label selector is only ever a post-filter over the full
// enumeration of the kind, because the informer cache indexes namespace and explicitly registered
// field selectors - never labels. One "in" query walks that enumeration once, matches each object
// with a set lookup, and materializes only the AccessRules that actually match.
//
// The list is typed, not unstructured, so it is served by the same informer that getObject already
// uses. Listing unstructured would start a second informer and a second watch over every AccessRule
// in the cluster, holding a second copy of them in memory.
func (h *AccessRulesHandler) listAccessRulesByIds(accessRuleIds []string) ([]v1alpha1.AccessRule, error) {
	if len(accessRuleIds) == 0 {
		return nil, nil
	}

	requirement, err := labels.NewRequirement(v1alpha1.AccessRuleIdLabel, selection.In, accessRuleIds)
	if err != nil {
		log.Error().Msgf("Failed building <%s in (...)> selector for AccessRules, error: %s",
			v1alpha1.AccessRuleIdLabel, err.Error())
		return nil, err
	}

	accessRuleList := &v1alpha1.AccessRuleList{}
	err = h.cachedClient.List(context.Background(), accessRuleList, &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*requirement),
	})
	if err != nil {
		log.Error().Msgf("Failed listing AccessRules by <%s>, error: %s",
			v1alpha1.AccessRuleIdLabel, err.Error())
		return nil, err
	}

	return accessRuleList.Items, nil
}

// deleteAccessRulesByIds deletes every AccessRule carrying one of the given accessrule ids,
// whatever name it was created under.
//
// Deleting by id rather than by the generated name is what makes a rename survivable: the name
// embeds the role name, so after a rename the object in the cluster no longer carries the name we
// would generate now, and a name-keyed delete could never reach it.
func (h *AccessRulesHandler) deleteAccessRulesByIds(accessRuleIds []string) error {
	accessRules, err := h.listAccessRulesByIds(accessRuleIds)
	if err != nil {
		return err
	}

	for index := range accessRules {
		accessRule := &accessRules[index]

		if innerErr := h.deleteObject(accessRule); innerErr != nil {
			err = multierror.Append(err, innerErr)
			h.createDeletionEvent(accessRule, innerErr)
		}
	}

	return err
}

func (h *AccessRulesHandler) MetricLabel() string {
	return "access_rules"
}

func (h *AccessRulesHandler) getBackendAccessRulesSinceLastUpdated() (
	[]auth_openapi.AccessRuleSync, error,
) {
	var offset int32 = 0

	accessRulesSinceLastUpdated := []auth_openapi.AccessRuleSync{}

	initialLastUpdated := ""
	if h.lastUpdated != nil {
		initialLastUpdated = convertTimeToString(h.lastUpdated)
	}

	limit := config.Config().AccessRulesLimit
	for {
		log.Debug().Msgf("Getting AccessRules from backend; offset: <%d>, limit: <%d>, lastUpdated: <%s>",
			offset, limit, convertTimeToString(h.lastUpdated))

		backendAccessRulesResponse, res, err := h.backendClient.AuthClient.GetSyncAccessRules(
			h.backendClient.Context(), int32(limit), offset, initialLastUpdated, config.Config().ClusterUuid)
		err = h.backendClient.HandleHttpResponse(res, err, "Get AccessRules")
		if err != nil || backendAccessRulesResponse == nil {
			log.Error().Msgf("Had an error while getting accessrules: %+v", err)
			return []auth_openapi.AccessRuleSync{}, err
		}

		currentAccessRules := backendAccessRulesResponse.AccessRules
		accessRulesSinceLastUpdated = append(accessRulesSinceLastUpdated, currentAccessRules...)

		// if we got fewer items than requested (than the limit) -
		// it means the backend finished sending all the relevant data
		if len(currentAccessRules) < limit {
			log.Debug().Msgf("Finished getting all AccessRules from backend")
			break
		}

		offset = offset + int32(limit)
	}

	return accessRulesSinceLastUpdated, nil
}

// extractAccessRules - this function extracts the received backend access rules -
// converts them to cluster objects if needed, or adds them to accessRulesToDelete;
// also calculates the maximal lastUpdated out of all objects (out of UpdatedAt or DeletedAt).
func (h *AccessRulesHandler) extractAccessRules(backendAccessRules []auth_openapi.AccessRuleSync) (AccessRulesExtractionResult, error) {
	var err error
	clusterAccessRules := []client.Object{}
	accessRuleIdsToDelete := []string{}
	lastUpdated := h.lastUpdated

	for _, backendAccessRule := range backendAccessRules {
		if (backendAccessRule.DeletedAt.IsSet() && backendAccessRule.DeletedAt.Get() != nil) ||
			(backendAccessRule.IsDeleting != nil && *backendAccessRule.IsDeleting) {
			accessRuleIdsToDelete = append(accessRuleIdsToDelete, accessRuleIdLabelValue(&backendAccessRule))
		} else {
			clusterAccessRule, innerErr := convertBackendToClusterAccessRule(&backendAccessRule)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
				continue
			}
			clusterAccessRules = append(clusterAccessRules, clusterAccessRule)
		}

		if lastUpdated == nil {
			lastUpdated = &backendAccessRule.UpdatedAt
		} else if backendAccessRule.UpdatedAt.After(*lastUpdated) {
			lastUpdated = &backendAccessRule.UpdatedAt
		}
	}

	return AccessRulesExtractionResult{
		clusterAccessRules:    clusterAccessRules,
		accessRuleIdsToDelete: accessRuleIdsToDelete,
		lastUpdated:           lastUpdated,
	}, err
}

func getPredefinedRoleName(backendAccessRule *auth_openapi.AccessRuleSync) *string {
	if backendAccessRule.PredefinedRole.Get() != nil {
		predefinedRoleName := ptr.Deref(backendAccessRule.PredefinedRole.Get().Name, "")
		if predefinedRoleName != "" {
			return ptr.To(convertRoleNameToK8sAllowed(predefinedRoleName))
		}
	}
	return nil
}

func convertBackendToClusterAccessRule(backendAccessRule *auth_openapi.AccessRuleSync) (*v1alpha1.AccessRule, error) {
	backendAccessRule.RoleName = convertRoleNameToK8sAllowed(backendAccessRule.RoleName)

	accessrule, err := convertJsonAccessRuleFields(backendAccessRule)
	if err != nil {
		return nil, err
	}

	accessrule.Name = GenerateAccessRuleName(backendAccessRule)

	accessRuleId := accessRuleIdLabelValue(backendAccessRule)

	predefinedRoleName := getPredefinedRoleName(backendAccessRule)
	if predefinedRoleName != nil {
		accessrule.Spec.RoleName = *predefinedRoleName
	}

	accessrule.Labels = map[string]string{
		v1alpha1.AccessRuleIdLabel: accessRuleId,
	}

	if backendAccessRule.ScopeType == auth_openapi.SCOPETYPE_TENANT {
		accessrule.Spec.ScopeType = v1alpha1.ClusterScope
		accessrule.Spec.ScopeName = ""
	}

	return accessrule, nil
}

/*
convertJsonAccessRuleFields converts backend accessrule to cluster accessrule, using json marshal/unmarshal.
Not all fields are matching in both objects - so other fields needs to be handled manually.
*/
func convertJsonAccessRuleFields(backendAccessRule *auth_openapi.AccessRuleSync) (*v1alpha1.AccessRule, error) {
	jsonBytes, err := json.Marshal(backendAccessRule)
	if err != nil {
		log.Error().Msgf("Failed to json marshal backend accessrule id <%d>, error: %s",
			backendAccessRule.Id, err.Error())
		return nil, err
	}

	crdAccessRule := &v1alpha1.AccessRule{}
	err = json.Unmarshal(jsonBytes, &crdAccessRule.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json un-marshal crd accessrule id <%d>, error: %s",
			backendAccessRule.Id, err.Error())
		return nil, err
	}

	return crdAccessRule, nil
}

func convertTimeToString(t *time.Time) string {
	if t != nil {
		return t.UTC().Format(time.RFC3339)
	}
	return "nil"
}

// accessRuleIdLabelValue returns the runai/accessrule-id label value for a backend access rule.
// This label is the only stable identity an AccessRule object has, and it is what both the agent
// and cluster-sync key off; the object's name is not stable - see GenerateAccessRuleName.
func accessRuleIdLabelValue(backendAccessRule *auth_openapi.AccessRuleSync) string {
	return strconv.FormatInt(int64(backendAccessRule.Id), 10)
}

// GenerateAccessRuleName - the access rule backend object doesn't have a name;
// we need to generate one because the crd instance needs a name.
func GenerateAccessRuleName(backendAccessRule *auth_openapi.AccessRuleSync) string {
	scopeStr := ""
	if backendAccessRule.ScopeType == auth_openapi.SCOPETYPE_TENANT ||
		backendAccessRule.ScopeType == auth_openapi.SCOPETYPE_CLUSTER {
		scopeStr = string(auth_openapi.SCOPETYPE_CLUSTER)
	} else {
		scopeStr = fmt.Sprintf("%s-%s", backendAccessRule.ScopeType, backendAccessRule.ScopeId)
	}

	subjectStr := backendAccessRule.SubjectId
	if backendAccessRule.SubjectType == auth_openapi.SUBJECTTYPEEXTENDED_GROUP {
		subjectStr = fmt.Sprintf("%s-%s", backendAccessRule.SubjectType, backendAccessRule.SubjectId)
	}

	backendAccessRuleRoleName := backendAccessRule.RoleName

	predefinedRoleName := getPredefinedRoleName(backendAccessRule)
	if predefinedRoleName != nil {
		backendAccessRuleRoleName = *predefinedRoleName
	}

	accessRuleName := fmt.Sprintf("ar-%d-%s-%s-%s",
		backendAccessRule.Id,
		scopeStr,
		backendAccessRuleRoleName,
		subjectStr)

	return convertToK8sResourceName(accessRuleName)
}

func convertToK8sResourceName(name string) string {
	result := name
	if len(name) > maxNameLen {
		result = name[:maxNameLen]
	}

	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return '-'
	}, result)

	if result[len(result)-1] == '-' {
		result = fmt.Sprintf("%s%s", result, nameSuffixIfNeeded)
	}

	result = strings.ToLower(result)
	return result
}

func convertRoleNameToK8sAllowed(roleName string) string {
	result := strings.ReplaceAll(roleName, " ", "-")
	return strings.ToLower(result)
}

func (a AccessRuleObjectInspector) DeepCopySpec(from, to client.Object) {
	fromAccessRule, ok := from.(*v1alpha1.AccessRule)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toAccessRule, ok := to.(*v1alpha1.AccessRule)
	if !ok {
		return
	}

	fromAccessRule.Spec.DeepCopyInto(&toAccessRule.Spec)

	fromAccessRuleId, found := fromAccessRule.Labels[v1alpha1.AccessRuleIdLabel]
	if !found {
		return
	}

	if toAccessRule.Labels == nil {
		toAccessRule.Labels = map[string]string{}
	}
	toAccessRule.Labels[v1alpha1.AccessRuleIdLabel] = fromAccessRuleId
}

func (a AccessRuleObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftAccessRule, ok := left.(*v1alpha1.AccessRule)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightAccessRule, ok := right.(*v1alpha1.AccessRule)
	if !ok {
		return false
	}

	leftSpecJson, err := json.Marshal(leftAccessRule.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal accessrule spec, error: %s", err.Error())
		return false
	}
	rightSpecJson, err := json.Marshal(rightAccessRule.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal accessrule spec, error: %s", err.Error())
		return false
	}

	if !reflect.DeepEqual(leftSpecJson, rightSpecJson) {
		return false
	}

	leftAccessRuleId, foundLeft := leftAccessRule.Labels[v1alpha1.AccessRuleIdLabel]
	rightAccessRuleId, foundRight := rightAccessRule.Labels[v1alpha1.AccessRuleIdLabel]
	return (foundLeft == foundRight) && (leftAccessRuleId == rightAccessRuleId)
}
