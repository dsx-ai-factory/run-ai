package handlers

import (
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
	"k8s.io/apimachinery/pkg/types"

	auth_openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

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
	clusterAccessRules  []client.Object
	accessRulesToDelete []types.NamespacedName
	lastUpdated         *time.Time
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
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(accessRulesExtractionResult.accessRulesToDelete))

	// don't sync deleted - because we are getting the deleted objects from the backend and calling DeleteNamedObjects
	_, _, innerErr := h.CreateOrUpdateObjects(accessRulesExtractionResult.clusterAccessRules)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	_, innerErr = h.DeleteObjects(accessRulesExtractionResult.accessRulesToDelete)
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
	accessRulesToDelete := []types.NamespacedName{}
	lastUpdated := h.lastUpdated

	for _, backendAccessRule := range backendAccessRules {
		if (backendAccessRule.DeletedAt.IsSet() && backendAccessRule.DeletedAt.Get() != nil) ||
			(backendAccessRule.IsDeleting != nil && *backendAccessRule.IsDeleting) {
			accessRulesToDelete = append(accessRulesToDelete, types.NamespacedName{Name: GenerateAccessRuleName(&backendAccessRule), Namespace: ""})
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
		clusterAccessRules:  clusterAccessRules,
		accessRulesToDelete: accessRulesToDelete,
		lastUpdated:         lastUpdated,
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

	accessRuleId := strconv.FormatInt(int64(backendAccessRule.Id), 10)

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
