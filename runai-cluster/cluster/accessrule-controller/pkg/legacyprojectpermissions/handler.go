package legacyprojectpermissions

import (
	"context"
	"encoding/json"
	"slices"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/mappers"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectLegacyPermissionsHandler struct {
	cachedClient    client.Client
	intervalSeconds int

	accessRulesProjectsMapper *mappers.AccessRulesProjectsMapper
}

func NewProjectLegacyPermissionsHandler(cachedClient client.Client, intervalSeconds int) *ProjectLegacyPermissionsHandler {
	accessRulesProjectsMapper := mappers.NewAccessRulesProjectsMapper(cachedClient)

	return &ProjectLegacyPermissionsHandler{
		cachedClient:              cachedClient,
		intervalSeconds:           intervalSeconds,
		accessRulesProjectsMapper: accessRulesProjectsMapper,
	}
}

func (h *ProjectLegacyPermissionsHandler) Start(ctx context.Context, stopCh <-chan struct{}) {
	h.runHandleAllPeriodically(ctx, stopCh)
}

func (h *ProjectLegacyPermissionsHandler) runHandleAllPeriodically(ctx context.Context, stopCh <-chan struct{}) {
	log.Info().Msgf("Starting periodic handling of project permissions every %d seconds", h.intervalSeconds)

	interval := time.Duration(h.intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)

	go func() {
		h.runHandleAll(ctx)
		for {
			select {
			case <-ticker.C:
				h.runHandleAll(ctx)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (h *ProjectLegacyPermissionsHandler) runHandleAll(ctx context.Context) {
	projects := &v2.ProjectList{}
	err := h.cachedClient.List(ctx, projects)
	if err != nil {
		log.Error().Msgf("Failed to list projects, err: %v", err)
		return
	}

	failedProjects := []string{}
	for _, project := range projects.Items {
		err = h.handleProject(ctx, &project)
		if err != nil {
			failedProjects = append(failedProjects, project.Name)
		}
	}

	if len(failedProjects) > 0 {
		log.Error().Msgf("Failed to handle legacy permissions for projects: %v", failedProjects)
	}
}

func (h *ProjectLegacyPermissionsHandler) handleProject(ctx context.Context, project *v2.Project) error {
	accessRules, err := h.accessRulesProjectsMapper.GetAllAccessRulesForOrgUnit(ctx, project.Name, project.Spec.Department)
	if err != nil {
		log.Error().Msgf("Failed getting all access rules to reconcile for project <%s>, error: %s",
			project.Name, err.Error())
		return err
	}

	requestedPermissions := constructPermissionsObjectFromAccessRules(accessRules)

	// if existing is equal to requested - return
	if slices.Equal(project.Spec.Permissions.Users, requestedPermissions.Users) &&
		slices.Equal(project.Spec.Permissions.Groups, requestedPermissions.Groups) {
		return nil
	}

	patchBytes, err := updateProjectPermissionsPatchBytes(requestedPermissions)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update permissions of project <%s>, err: %v", project.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Patch(ctx, project, patch)
	if err != nil {
		log.Error().Msgf("Failed patching project <%s> with permissions <%v>, err: %v",
			project.Name, requestedPermissions, err.Error())
		return err
	}

	log.Info().Msgf("Patched project <%s> with permissions <%v>",
		project.Name, requestedPermissions)
	return nil
}

func updateProjectPermissionsPatchBytes(permissions v2.ProjectPermissions) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"spec": map[string]interface{}{
		"permissions": permissions}})
}

func constructPermissionsObjectFromAccessRules(accessRules []v1alpha1.AccessRule) v2.ProjectPermissions {
	users := map[string]bool{}
	groups := map[string]bool{}

	for _, accessRule := range accessRules {
		switch accessRule.Spec.SubjectType {
		case v1alpha1.UserSubject, v1alpha1.AppSubject, v1alpha1.ServiceAccountSubject:
			users[accessRule.Spec.SubjectId] = true
		case v1alpha1.GroupSubject:
			groups[accessRule.Spec.SubjectId] = true
		case v1alpha1.AllAuthenticatedUsersSubject:
			groups[rolebindingsutils.AuthenticatedUsersGroup] = true
		}
	}

	return v2.ProjectPermissions{
		Users:  extractMapToSortedList(users),
		Groups: extractMapToSortedList(groups),
	}
}

func extractMapToSortedList(m map[string]bool) []string {
	list := make([]string, 0, len(m))
	for key := range m {
		list = append(list, key)
	}
	sort.Strings(list)
	return list
}
