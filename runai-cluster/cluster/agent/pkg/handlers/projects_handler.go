package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"k8s.io/utils/ptr"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NoMaxAllowed           = -1
	DefaultOverQuotaWeight = 1
)

type ProjectsHandler struct {
	*CommonHandler
	lastUpdated time.Time
}

func NewProjectsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	var prototype client.Object = &kaiorg.Project{}
	var inspector ObjectInspector = KaiProjectObjectInspector(true)

	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		prototype, inspector, HandlingOptions{deletePropagationOrphan: true})
	if err != nil {
		return nil, err
	}

	return &ProjectsHandler{
		CommonHandler: commonHandler,
		lastUpdated:   time.Time{},
	}, nil
}

func (h *ProjectsHandler) Handle() error {
	log.Debug().Msgf("Starting to sync projects, lastUpdated is %v", h.lastUpdated)

	projects, err := h.getOrgUnitProjects()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return nil
	}

	projectsToAddUpdate, projectsToDelete, projectsWithDeletionRequest, lastUpdated, err := h.getClusterObjects(projects)
	// projects with deletion request are updated to be deleted, therefore they should be count as "created or update" metric
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(projectsToAddUpdate)+len(projectsWithDeletionRequest))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(projectsToDelete))

	// do not change order of delete and create since there can be deletion of project and then creation of new one with the same name
	innerErr := h.handleDeletionRequest(projectsWithDeletionRequest)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	_, innerErr = h.DeleteObjectsById(projectsToDelete, v2.ProjectIdLabel)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	_, _, innerErr = h.CreateOrUpdateObjects(projectsToAddUpdate)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		log.Error().Msgf("Failed to handle Projects, error: %s", err.Error())
		return err
	}

	h.lastUpdated = lastUpdated

	log.Debug().Msgf("Successfully handled Projects")
	return err
}

func (h *ProjectsHandler) MetricLabel() string {
	return "projects"
}

func (h *ProjectsHandler) getClusterObjects(projects []orgUnitApi.ProjectSync) (
	[]client.Object, map[string]types.NamespacedName, []client.Object, time.Time, error) {
	var err error
	projectsToAddUpdate := []client.Object{}
	projectsWithDeletionRequest := []client.Object{}
	projectsToDelete := map[string]types.NamespacedName{}
	lastUpdated := h.lastUpdated

	// Fetch only the cluster KAI departments referenced by the synced projects, so we can
	// validate that every project's parent department is present before creating/updating it.
	departmentIds, err := h.getRelatedClusterDepartments(projects)
	if err != nil {
		return nil, nil, nil, lastUpdated, err
	}

	for _, project := range projects {
		if project.DeletedAt.IsSet() && project.DeletedAt.Get() != nil {
			projectsToDelete[project.Id] = types.NamespacedName{Name: project.Name, Namespace: ""}
		} else {
			clusterProject, convErr := convertOrgUnitToKaiProject(project)
			if convErr != nil {
				err = multierror.Append(err, convErr)
				continue
			}

			if clusterProject.Spec.DeletionType != nil {
				projectsWithDeletionRequest = append(projectsWithDeletionRequest, clusterProject)
			} else {
				if !h.isKaiDepartmentPresent(clusterProject, departmentIds) {
					log.Error().Msgf("Skipping KAI project <%s>: parent department <%s> not found in the cluster",
						clusterProject.Name, clusterProject.Spec.Parent)
					err = multierror.Append(err, fmt.Errorf(
						"parent department <%s> of KAI project <%s> not found in the cluster",
						clusterProject.Spec.Parent, clusterProject.Name))
					continue
				}
				projectsToAddUpdate = append(projectsToAddUpdate, clusterProject)
			}
		}

		if project.UpdatedAt.After(lastUpdated) {
			lastUpdated = project.UpdatedAt
		}
	}

	return projectsToAddUpdate, projectsToDelete, projectsWithDeletionRequest, lastUpdated, err
}

func (h *ProjectsHandler) isKaiDepartmentPresent(clusterProject *kaiorg.Project, departmentIds map[string]bool) bool {
	departmentId := clusterProject.Labels[schedv2alpha2.DepartmentIdLabel]
	if departmentIds[departmentId] {
		return true
	}
	// Not matched by department-id; fall back to matching the parent department by name.
	department := &kaiorg.Department{}
	err := h.cachedClient.Get(context.Background(),
		types.NamespacedName{Name: clusterProject.Spec.Parent, Namespace: ""}, department)
	return err == nil
}

func convertOrgUnitToKaiProject(backendProject orgUnitApi.ProjectSync) (*kaiorg.Project, error) {
	project := &kaiorg.Project{}
	project.Name = backendProject.Name

	project.Labels = map[string]string{
		v2.ProjectIdLabel: backendProject.Id,
	}
	if backendProject.Parent.IsSet() && backendProject.Parent.Get() != nil {
		project.Spec.Parent = backendProject.Parent.Get().Name
		project.Labels[schedv2alpha2.DepartmentIdLabel] = backendProject.Parent.Get().Id
	}

	if backendProject.RequestedNamespace.IsSet() && backendProject.RequestedNamespace.Get() != nil {
		project.Spec.Namespace = *backendProject.RequestedNamespace.Get()
	}
	if backendProject.EnforceRunaiScheduler != nil {
		project.Spec.EnforceKaiScheduler = *backendProject.EnforceRunaiScheduler
	}

	project.Spec.DefaultNodePools = backendProject.DefaultNodePools

	project.Spec.Queues = convertToKaiQueueConfigs(project.Name, backendProject.Resources)

	// Product-specific fields that are not part of the generic KAI spec are carried as
	// annotations (scheduling rules, deprecated node-affinity, force-delete).
	project.Annotations = buildKaiProjectAnnotations(backendProject)

	if backendProject.Deletion.IsSet() && backendProject.Deletion.Get() != nil {
		deletionType := kaiorg.DeletionType(backendProject.Deletion.Get().Type)
		project.Spec.DeletionType = &deletionType
	}

	return project, nil
}

// kaiProjectManagedAnnotations are the annotation keys the agent owns on a KAI Project.
// They are the single source of truth for both writing (buildKaiProjectAnnotations) and
// drift detection (KaiProjectObjectInspector).
var kaiProjectManagedAnnotations = []string{
	constants.InteractiveJobTimeLimitSecondsAnnotation,
	constants.InteractiveJobMaxIdleDurationSecondsAnnotation,
	constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation,
	constants.TrainingJobTimeLimitSecondsAnnotation,
	constants.TrainingJobMaxIdleDurationSecondsAnnotation,
	constants.NodeAffinityInteractiveAnnotation,
	constants.NodeAffinityTrainAnnotation,
	constants.ForceDeleteAnnotation,
	constants.UpdatedAtAnnotation,
}

func buildKaiProjectAnnotations(backendProject orgUnitApi.ProjectSync) map[string]string {
	annotations := map[string]string{}

	if backendProject.SchedulingRules.IsSet() && backendProject.SchedulingRules.Get() != nil {
		rules := backendProject.SchedulingRules.Get()
		setSecondsAnnotation(annotations, constants.TrainingJobTimeLimitSecondsAnnotation, rules.TrainingJobTimeLimitSeconds)
		setSecondsAnnotation(annotations, constants.TrainingJobMaxIdleDurationSecondsAnnotation, rules.TrainingJobMaxIdleDurationSeconds)
		setSecondsAnnotation(annotations, constants.InteractiveJobTimeLimitSecondsAnnotation, rules.InteractiveJobTimeLimitSeconds)
		setSecondsAnnotation(annotations, constants.InteractiveJobMaxIdleDurationSecondsAnnotation, rules.InteractiveJobMaxIdleDurationSeconds)
		setSecondsAnnotation(annotations, constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation, rules.InteractiveJobPreemptIdleDurationSeconds)
	}

	interactiveNodeAffinity, trainNodeAffinity := convertBackendJobsNodeAffinity(backendProject.NodeTypes)
	setStringSliceAnnotation(annotations, constants.NodeAffinityInteractiveAnnotation, interactiveNodeAffinity)
	setStringSliceAnnotation(annotations, constants.NodeAffinityTrainAnnotation, trainNodeAffinity)

	if backendProject.Deletion.IsSet() && backendProject.Deletion.Get() != nil {
		annotations[constants.ForceDeleteAnnotation] = strconv.FormatBool(backendProject.Deletion.Get().Force)
	}

	annotations[constants.UpdatedAtAnnotation] = backendProject.UpdatedAt.UTC().Format(time.RFC3339)

	if len(annotations) == 0 {
		return nil
	}
	return annotations
}

func setSecondsAnnotation(annotations map[string]string, key string, value orgUnitApi.NullableInt32) {
	if !value.IsSet() || value.Get() == nil {
		return
	}
	annotations[key] = strconv.FormatInt(int64(*value.Get()), 10)
}

func setStringSliceAnnotation(annotations map[string]string, key string, values []string) {
	if len(values) == 0 {
		return
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		log.Error().Msgf("Failed to marshal annotation <%s>, error: %s", key, err.Error())
		return
	}
	annotations[key] = string(encoded)
}

type KaiProjectObjectInspector bool

func (p KaiProjectObjectInspector) DeepCopySpec(from, to client.Object) {
	fromProject, ok := from.(*kaiorg.Project)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toProject, ok := to.(*kaiorg.Project)
	if !ok {
		return
	}

	fromProject.Spec.DeepCopyInto(&toProject.Spec)
	copyOrgUnitIdLabel(fromProject, toProject, v2.ProjectIdLabel)
	copyOrgUnitIdLabel(fromProject, toProject, schedv2alpha2.DepartmentIdLabel)
	copyManagedAnnotations(fromProject, toProject, kaiProjectManagedAnnotations)
}

func (p KaiProjectObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftProject, ok := left.(*kaiorg.Project)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightProject, ok := right.(*kaiorg.Project)
	if !ok {
		return false
	}

	if !specJsonEqual(leftProject.Spec, rightProject.Spec) {
		return false
	}

	if !managedAnnotationsEqual(leftProject.Annotations, rightProject.Annotations, kaiProjectManagedAnnotations) {
		return false
	}

	if !idLabelEqual(leftProject.Labels, rightProject.Labels, v2.ProjectIdLabel) {
		return false
	}
	return idLabelEqual(leftProject.Labels, rightProject.Labels, schedv2alpha2.DepartmentIdLabel)
}

func copyManagedAnnotations(from, to client.Object, keys []string) {
	fromAnnotations := from.GetAnnotations()
	toAnnotations := to.GetAnnotations()
	if toAnnotations == nil {
		toAnnotations = map[string]string{}
	}
	for _, key := range keys {
		if value, found := fromAnnotations[key]; found {
			toAnnotations[key] = value
		} else {
			delete(toAnnotations, key)
		}
	}
	to.SetAnnotations(toAnnotations)
}

func managedAnnotationsEqual(left, right map[string]string, keys []string) bool {
	for _, key := range keys {
		leftValue, foundLeft := left[key]
		rightValue, foundRight := right[key]
		if foundLeft != foundRight || leftValue != rightValue {
			return false
		}
	}
	return true
}

// getRelatedClusterDepartments fetches only the cluster departments referenced by the given projects
// (filtered server-side by the projects' parent department-ids) and returns the sets of department-ids
// that exist in the cluster. These are used to validate that a project's parent department is present before the project is reconciled.
func (h *ProjectsHandler) getRelatedClusterDepartments(projects []orgUnitApi.ProjectSync) (map[string]bool, error) {
	departmentIds := map[string]bool{}

	// Collect the parent department-ids referenced by the projects.
	referencedIds := map[string]bool{}
	for _, project := range projects {
		if !project.DeletedAt.IsSet() && project.Parent.IsSet() && project.Parent.Get() != nil && project.Parent.Get().Id != "" {
			referencedIds[project.Parent.Get().Id] = true
		}
	}
	if len(referencedIds) == 0 {
		return departmentIds, nil
	}

	idValues := make([]string, 0, len(referencedIds))
	for id := range referencedIds {
		idValues = append(idValues, id)
	}

	requirement, err := labels.NewRequirement(schedv2alpha2.DepartmentIdLabel, selection.In, idValues)
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*requirement)

	departmentList := &kaiorg.DepartmentList{}
	if err := h.cachedClient.List(context.Background(), departmentList,
		&client.ListOptions{LabelSelector: selector}); err != nil {
		log.Error().Msgf("Failed to list KAI departments from cluster: err: %v", err)
		return nil, err
	}

	for _, department := range departmentList.Items {
		if departmentId, found := department.Labels[schedv2alpha2.DepartmentIdLabel]; found {
			departmentIds[departmentId] = true
		}
	}

	return departmentIds, nil
}

func (h *ProjectsHandler) getOrgUnitProjects() ([]orgUnitApi.ProjectSync, error) {
	projects := []orgUnitApi.ProjectSync{}

	offset := int32(0)
	limit := int32(config.Config().ProjectsLimit)

	for {
		log.Debug().Msgf("Getting Projects from org-unit; offset: <%d>, limit: <%d>, lastUpdated: <%v>",
			offset, limit, h.lastUpdated)

		projectsBulk, err := h.backendClient.OrgUnitClient.GetProjectsSync(
			h.backendClient.Context(), config.Config().ClusterUuid, limit, offset, h.lastUpdated, true)

		if err != nil {
			log.Error().Msgf("Failed to get projects from org-unit: err: %v", err)
			return projects, err
		}

		projects = append(projects, projectsBulk.Projects...)

		if len(projectsBulk.Projects) < int(limit) {
			log.Debug().Msgf("Finished getting all projects")
			break
		}

		offset = offset + limit
	}

	return projects, nil
}

func (h *ProjectsHandler) handleDeletionRequest(projectsWithDeletion []client.Object) error {
	var err error
	projectsToDelete := []types.NamespacedName{}
	for _, project := range projectsWithDeletion {
		projectsToDelete = append(projectsToDelete, types.NamespacedName{Name: project.GetName(), Namespace: ""})
	}
	_, _, innerErr := h.CreateOrUpdateObjects(projectsWithDeletion)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}
	_, innerErr = h.DeleteObjects(projectsToDelete)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}
	return err
}

func getPriorityOrDefault(priority *int32) *int32 {
	defaultPriority := ptr.To(int32(constants.DefaultQueuePriority))
	if priority == nil {
		return defaultPriority
	}
	return priority
}

// convertBackendJobsNodeAffinity returns NodeAffinityInteractive, NodeAffinityTrain
func convertBackendJobsNodeAffinity(jobsNodeAffinity *orgUnitApi.NodeTypesPerWorkload) ([]string, []string) {
	if jobsNodeAffinity == nil {
		return []string{}, []string{}
	}

	trainingNames := []string{}
	workspaceNames := []string{}
	for _, trainingId := range jobsNodeAffinity.Training {
		name, found := jobsNodeAffinity.Names[trainingId]
		if found {
			trainingNames = append(trainingNames, name)
		}
	}
	for _, workspaceId := range jobsNodeAffinity.Workspace {
		name, found := jobsNodeAffinity.Names[workspaceId]
		if found {
			workspaceNames = append(workspaceNames, name)
		}
	}

	return workspaceNames, trainingNames
}
