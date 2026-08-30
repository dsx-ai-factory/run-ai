package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"encoding/json"

	multierror "github.com/hashicorp/go-multierror"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	replicatedProjectNameLabelKey = "run.ai/replicated-project-name"
	replicatedProjectIdLabelKey   = "run.ai/replicated-project-uid"
	ResourceManualOverrideLabel   = "run.ai/resource-override"
)

type ClusterWideReplicatedResourceHandler struct {
	client.Client
	clusterWideGVKCommon *ClusterWideGVKCommon
}

func NewClusterWideReplicatedResourceHandler(
	client client.Client, clusterWideGVKCommon *ClusterWideGVKCommon,
) *ClusterWideReplicatedResourceHandler {
	return &ClusterWideReplicatedResourceHandler{
		Client:               client,
		clusterWideGVKCommon: clusterWideGVKCommon,
	}
}

func (handler *ClusterWideReplicatedResourceHandler) HandleResource(
	ctx context.Context, project kaiorg.Project) error {
	if len(handler.clusterWideGVKCommon.GetGVKObjectConfigs()) == 0 {
		return nil
	}

	namespaceName := project.Status.Namespace
	if namespaceName == "" {
		log.Ctx(ctx).Info().Msgf("Skipping ServiceMonitor reconciliation for project %s: namespace not yet set", project.Name)
		return nil
	}

	var err error
	if err = utils.ValidateNamespaceExists(ctx, handler.Client, namespaceName, project.Name); err != nil {
		return err
	}

	for gvk, gvkObjectConfigs := range handler.clusterWideGVKCommon.GetGVKObjectConfigs() {
		for _, gvkObjectConfig := range gvkObjectConfigs {
			innerErr := handler.handleSpecificResource(ctx, project, namespaceName, gvk, gvkObjectConfig)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}
		}
	}

	return err
}

func (handler *ClusterWideReplicatedResourceHandler) handleSpecificResource(
	ctx context.Context, project kaiorg.Project, namespace string, gvk schema.GroupVersionKind, gvkObjectConfig ObjectConfig) error {
	log.Ctx(ctx).Info().Msgf("Handling %v-wide resource replication to runai-managed namespaces", gvkObjectConfig.Scope)

	var err error

	objectsReplicatedInNamespace, innerErr := handler.createScopeWideObjectsUnderProjectNamespace(ctx, project,
		namespace, gvk, gvkObjectConfig)
	if innerErr != nil {
		log.Ctx(ctx).Error().Msgf("Failed to replicate cluster-wide objects from the 'runai' namespace to project's namespace %s, gvk %s. error: %s",
			namespace, gvk.String(), innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	objectsReplicatedInNamespaceNames, unreferencedObjectNames, innerErr := handler.deleteUnreferencedObjects(ctx, project,
		namespace, objectsReplicatedInNamespace, gvk, gvkObjectConfig)
	if innerErr != nil {
		log.Ctx(ctx).Error().Msgf("Failed to delete old unreferenced objects in namespace: %s, gvk %s. error: %s",
			namespace, gvk.String(), innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	if gvkObjectConfig.AdditionalCallback != nil {
		innerErr = gvkObjectConfig.AdditionalCallback(ctx, namespace, objectsReplicatedInNamespaceNames, unreferencedObjectNames)
		if innerErr != nil {
			log.Ctx(ctx).Error().Msgf("Failed in additional callback in namespace: %s, gvk %s. error: %s",
				namespace, gvk.String(), innerErr.Error())
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

// Searches for Objects under our 'runai' Namespace which are labeled as "cluster-wide"
// and creates namespaced copies under this Project's Namespace
func (handler *ClusterWideReplicatedResourceHandler) createScopeWideObjectsUnderProjectNamespace(
	ctx context.Context, project kaiorg.Project, namespace string, gvk schema.GroupVersionKind,
	objectConfig ObjectConfig) ([]client.Object, error) {
	objectItems, err := handler.listScopeWideObjectsToReplicate(ctx, project, gvk, objectConfig)
	if err != nil {
		return []client.Object{}, err
	}

	if len(objectItems) == 0 {
		log.Ctx(ctx).Info().Msgf("No %v-wide objects found in the 'runai' Namespace %s", objectConfig.Scope, namespace)
		return []client.Object{}, nil
	}

	return handler.createObjectsUnderNamespace(
		ctx, project, namespace, objectItems, gvk)
}

func (handler *ClusterWideReplicatedResourceHandler) createObjectsUnderNamespace(
	ctx context.Context, project kaiorg.Project, namespace string,
	runaiNsObjects []client.Object, gvk schema.GroupVersionKind,
) ([]client.Object, error) {
	namespacedObjects := []client.Object{}
	var allErrors error
	issueKey := fmt.Sprintf("%s/%s", constants.CloneAnnotation, namespace)
	for _, clusterWideObject := range runaiNsObjects {
		// The original object tracks the clone issue status using annotation for each ns
		// In case there isn't clone issue, no annotation is store for the ns, patching with nil
		// will clear issue from prev attempts to clone the object
		annotations := map[string]interface{}{
			issueKey: nil,
		}
		namespacedObject, err := handler.createNamespacedObjectIfNeeded(
			ctx, project, namespace, clusterWideObject, gvk)
		// add it to the list even if error intentionally,
		// so we don't try to delete it later in "deleteUnreferencedObjects".
		namespacedObjects = append(namespacedObjects, namespacedObject)
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
			annotations[issueKey] = constants.CloneError
		}
		err = handler.patchAnnotations(ctx, clusterWideObject, annotations)
		if err != nil {
			log.Ctx(ctx).Error().Msgf("Failed to update clone issue annotation on %s to namespace %s. error: %s", clusterWideObject.GetName(), namespace, err.Error())
			allErrors = multierror.Append(allErrors, err)
		}
	}
	// If any errors were encountered, signal for a requeue of the current event
	return namespacedObjects, allErrors
}

func (handler *ClusterWideReplicatedResourceHandler) createNamespacedObjectIfNeeded(
	ctx context.Context, project kaiorg.Project, namespace string,
	clusterWideObject client.Object, gvk schema.GroupVersionKind,
) (client.Object, error) {
	existingNamespacedClusterWideObject, err := scheme.Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	existingNamespacedObject, ok := existingNamespacedClusterWideObject.(client.Object)
	if !ok {
		return nil, fmt.Errorf("failed to parse existingNamespacedClusterWideObject to client.Object")
	}

	err = handler.Client.Get(ctx,
		client.ObjectKey{Namespace: namespace, Name: clusterWideObject.GetName()},
		existingNamespacedObject,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return handler.handleNonExisting(ctx, project, namespace, clusterWideObject, gvk)
		}
		log.Ctx(ctx).Error().Msgf("Error encountered while trying to read namespaced Object for Project. project %s, namespace %s, gvk %s, object %s. error: %s",
			project.Name, namespace, gvk.String(), clusterWideObject.GetName(), err.Error())
		return nil, err
	}

	err = handler.handleExisting(ctx, project, namespace, clusterWideObject, existingNamespacedObject, gvk)
	return existingNamespacedObject, err
}

func (handler *ClusterWideReplicatedResourceHandler) handleExisting(
	ctx context.Context, project kaiorg.Project, namespace string,
	clusterWideObject client.Object, existingNamespacedObject client.Object,
	gvk schema.GroupVersionKind,
) error {
	if !handler.clusterWideGVKCommon.ShouldHandleExisting(ctx, existingNamespacedObject) {
		log.Ctx(ctx).Debug().Msgf("Object already exists in namespace, it will not be updated. gvk %s, object %s, namespace %s, project %s",
			gvk.String(), existingNamespacedObject.GetName(), namespace, project.Name)
		return nil
	}
	if IsManuallyOverridden(existingNamespacedObject) {
		log.Ctx(ctx).Debug().Msgf(
			"Object already exists and is marked as manually overridden, it will not be updated. gvk %s, object %s, namespace %s, project %s",
			gvk.String(), existingNamespacedObject.GetName(), namespace, project.Name)
		return nil
	}

	newNamespacedObject, err := handler.createNamespacedObjectFromExisting(
		ctx, project, namespace, clusterWideObject, gvk)
	if err != nil {
		return err
	}
	// Preserve VolumeName from existing PVC to maintain PV binding
	handler.clusterWideGVKCommon.PreserveImmutableFields(existingNamespacedObject, newNamespacedObject)

	if handler.clusterWideGVKCommon.AreEqual(ctx, newNamespacedObject, existingNamespacedObject) {
		log.Ctx(ctx).Debug().Msgf("Object already exists in Namespace and no changes were detected. gvk %s, object %s, namespace %s",
			gvk.String(), existingNamespacedObject.GetName(), namespace)
		return nil
	}

	log.Ctx(ctx).Info().Msgf("Object exists in Namespace and changes were detected, updating with new data. gvk %s, object %s, namespace %s",
		gvk.String(), existingNamespacedObject.GetName(), namespace)

	immutableFieldsBackup := existingNamespacedObject.DeepCopyObject().(client.Object)
	handler.updateExistingWithNewObject(ctx, existingNamespacedObject, newNamespacedObject)
	handler.clusterWideGVKCommon.PreserveImmutableFields(immutableFieldsBackup, existingNamespacedObject)

	err = handler.Client.Update(ctx, existingNamespacedObject)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating object. gvk %s, object %s, namespace %s. error: %s",
			gvk.String(), existingNamespacedObject.GetName(), namespace, err.Error())
		return err
	}
	return nil
}

func (handler *ClusterWideReplicatedResourceHandler) handleNonExisting(
	ctx context.Context, project kaiorg.Project, namespace string,
	clusterWideObject client.Object, gvk schema.GroupVersionKind,
) (client.Object, error) {
	log.Ctx(ctx).Info().Msgf("Creating Secret of project. gvk %s, object %s, project %s, namespace %s",
		gvk.String(), clusterWideObject.GetName(), project.Name, namespace)

	newNamespacedObject, err := handler.createNamespacedObjectFromExisting(
		ctx, project, namespace, clusterWideObject, gvk)
	if err != nil {
		return nil, err
	}

	err = handler.Client.Create(ctx, newNamespacedObject)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Ctx(ctx).Error().Msgf("Failed to create namespaced Object for Project. project %s, namespace %s, gvk %s, object %s. error: %s",
			project.Name, namespace, gvk.String(), clusterWideObject.GetName(), err.Error())
		return newNamespacedObject, err
	}
	return newNamespacedObject, err
}

// deleteUnreferencedObjects - By "unreferenced" we mean secrets that are owned by the project
// (meaning this controller created them) and that their counterparts don't exist under the runai namespace anymore.
func (handler *ClusterWideReplicatedResourceHandler) deleteUnreferencedObjects(ctx context.Context, project kaiorg.Project,
	namespace string, objectsReplicatedInNamespace []client.Object, gvk schema.GroupVersionKind,
	objectConfig ObjectConfig) (objectsReplicatedInNamespaceNames []string,
	unreferencedObjectNames []string, err error) {
	allLabeledObjectsInNamespace, err := handler.listAllScopeWideObjectsInProjectsNamespace(ctx, project,
		gvk, namespace, objectConfig)
	if err != nil {
		return objectsReplicatedInNamespaceNames, unreferencedObjectNames, err
	}

	var unreferencedObjects []client.Object
	unreferencedObjects, unreferencedObjectNames, objectsReplicatedInNamespaceNames = getUnreferencedObjects(
		project.UID, project.Name, allLabeledObjectsInNamespace, objectsReplicatedInNamespace)
	err = handler.deleteObjects(ctx, unreferencedObjects, gvk)
	return objectsReplicatedInNamespaceNames, unreferencedObjectNames, err
}

func (handler *ClusterWideReplicatedResourceHandler) deleteObjects(
	ctx context.Context, unreferencedObjects []client.Object, gvk schema.GroupVersionKind) error {
	var err error
	for _, objectToDelete := range unreferencedObjects {
		log.Ctx(ctx).Info().Msgf("Deleting unreferenced object. gvk %s, object %s, namespace %s",
			gvk.String(), objectToDelete.GetName(), objectToDelete.GetNamespace())
		innerErr := handler.Client.Delete(ctx, objectToDelete)
		if innerErr != nil {
			log.Ctx(ctx).Error().Msgf("Failed to delete unreferenced object. gvk %s, object %s, namespace %s. error: %s",
				gvk.String(), objectToDelete.GetName(), objectToDelete.GetNamespace(), innerErr.Error())
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

func (handler *ClusterWideReplicatedResourceHandler) createNamespacedObjectFromExisting(
	ctx context.Context, project kaiorg.Project, namespace string,
	clusterWideObject client.Object, gvk schema.GroupVersionKind,
) (client.Object, error) {
	gvkObject, err := scheme.Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	newObject, ok := gvkObject.(client.Object)
	if !ok {
		return nil, fmt.Errorf("failed to parse gvkObject to client.Object")
	}

	newObject.SetName(clusterWideObject.GetName())
	newObject.SetNamespace(namespace)

	newObjectLabels := map[string]string{}
	copyRunaiLabels(clusterWideObject.GetLabels(), &newObjectLabels)
	replicatedProjectLabel, found := newObjectLabels[replicatedProjectNameLabelKey]
	if found && replicatedProjectLabel != "" {
		return nil, fmt.Errorf("object already has projectName replicated label: %s:%s", replicatedProjectNameLabelKey, replicatedProjectLabel)
	}

	newObjectLabels[replicatedProjectNameLabelKey] = project.Name
	newObjectLabels[replicatedProjectIdLabelKey] = string(project.UID)
	newObject.SetLabels(newObjectLabels)

	newObjectAnnotations := map[string]string{}
	copyRunaiAnnotations(clusterWideObject.GetAnnotations(), &newObjectAnnotations)
	newObject.SetAnnotations(newObjectAnnotations)

	handler.clusterWideGVKCommon.DeepCopySpec(ctx, clusterWideObject, newObject)
	return newObject, nil
}

func (handler *ClusterWideReplicatedResourceHandler) updateExistingWithNewObject(ctx context.Context, existingObj, newObj client.Object) {
	existingObjLabels := existingObj.GetLabels()
	copyRunaiLabels(newObj.GetLabels(), &existingObjLabels)
	existingObj.SetLabels(existingObjLabels)

	existingObjectAnnotations := existingObj.GetAnnotations()
	copyRunaiAnnotations(newObj.GetAnnotations(), &existingObjectAnnotations)
	existingObj.SetAnnotations(existingObjectAnnotations)

	handler.clusterWideGVKCommon.DeepCopySpec(ctx, newObj, existingObj)
}

func (handler *ClusterWideReplicatedResourceHandler) listScopeWideObjectsToReplicate(ctx context.Context, project kaiorg.Project,
	gvk schema.GroupVersionKind, objectConfig ObjectConfig) ([]client.Object, error) {
	return handler.listScopeWideObjectsInNamespace(ctx, project, gvk, clustercommonconstants.NamespaceRunAi, true, objectConfig)
}

func (handler *ClusterWideReplicatedResourceHandler) listAllScopeWideObjectsInProjectsNamespace(ctx context.Context, project kaiorg.Project,
	gvk schema.GroupVersionKind, namespace string, objectConfig ObjectConfig) ([]client.Object, error) {
	return handler.listScopeWideObjectsInNamespace(ctx, project, gvk, namespace, false, objectConfig)
}

// listScopeWideObjectsInNamespace
// lists cluster-wide or department-wide objects in runai namespace.
// if cluster-wide - objects in runai namespace that match the cluster-wide (scope-wide) label.
// if department-wide:
//
//		if onlyProjectMatchingObjects - objects in runai namespace that match
//					the department-wide (scope-wide) label with the projects' departmentId value.
//	    if not - objects in runai namespace that have the department-wide (scope-wide) label.
func (handler *ClusterWideReplicatedResourceHandler) listScopeWideObjectsInNamespace(ctx context.Context, project kaiorg.Project,
	gvk schema.GroupVersionKind, namespace string, onlyProjectMatchingObjects bool,
	objectConfig ObjectConfig) ([]client.Object, error) {
	if objectConfig.Scope == clusterScope {
		return handler.clusterWideGVKCommon.ListObjectsWithLabel(ctx, namespace, objectConfig.ScopeWideLabel,
			"true", gvk, objectConfig.AdditionalObjectFilter)
	}

	// department scoped
	if onlyProjectMatchingObjects {
		return handler.clusterWideGVKCommon.ListObjectsWithLabel(ctx, namespace, objectConfig.ScopeWideLabel,
			schedulingv2alpha2.GetDepartmentId(project.Labels), gvk, objectConfig.AdditionalObjectFilter)
	}
	return handler.clusterWideGVKCommon.ListObjectsWithLabel(ctx, namespace, objectConfig.ScopeWideLabel,
		"", gvk, objectConfig.AdditionalObjectFilter)
}

func (handler *ClusterWideReplicatedResourceHandler) patchAnnotations(ctx context.Context, obj client.Object, annotations map[string]interface{}) error {
	// Skip patching in case the current annotations match the annotations to patch
	if !ShouldPatchAnnotations(obj.GetAnnotations(), annotations) {
		return nil
	}

	metadata, err := json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"annotations": annotations}})
	if err != nil {
		return err
	}
	patch := client.RawPatch(types.MergePatchType, metadata)
	return handler.Client.Patch(ctx, obj, patch)
}

func getUnreferencedObjects(
	projectUID types.UID, projectName string, allLabeledObjectsInNamespace, objectsReplicatedInNamespace []client.Object,
) (unreferencedObjects []client.Object, unreferencedObjectNames []string, objectsReplicatedInNamespaceNames []string) {
	// Since there's no field selector support for OwnerReferences, have to do it the hard way.
	oldObjectsList := funk.Filter(allLabeledObjectsInNamespace, func(object client.Object) bool {
		ownerProjectName := object.GetLabels()[replicatedProjectNameLabelKey]
		ownerProjectId := object.GetLabels()[replicatedProjectIdLabelKey]
		hasLabel := ownerProjectName == projectName && ownerProjectId == string(projectUID)
		// -------------------------------------------------------------------------------
		// TODO: this should be removed after 2.27 and after new project crd upgrade!
		ownerReferences := object.GetOwnerReferences()
		hasOwner := len(ownerReferences) == 1 && ownerReferences[0].UID == projectUID
		// -------------------------------------------------------------------------------
		return hasLabel || hasOwner
	})

	oldObjectNames, ok := getObjectsNames(oldObjectsList)
	if !ok {
		return
	}
	objectsReplicatedInNamespaceNames, ok = getObjectsNames(objectsReplicatedInNamespace)
	if !ok {
		return
	}
	diff, _ := funk.Difference(oldObjectNames, objectsReplicatedInNamespaceNames)
	unreferencedObjectNames, ok = diff.([]string)
	if !ok {
		return
	}

	for _, oldSecret := range oldObjectsList.([]client.Object) {
		if funk.Contains(diff, oldSecret.GetName()) {
			unreferencedObjects = append(unreferencedObjects, oldSecret)
		}
	}
	return unreferencedObjects, unreferencedObjectNames, objectsReplicatedInNamespaceNames
}

var objectNameMapper = func(object client.Object) string {
	return object.GetName()
}

func getObjectsNames(objects interface{}) ([]string, bool) {
	names, ok := (funk.Map(objects, objectNameMapper)).([]string)
	return names, ok
}

func copyRunaiLabels(from map[string]string, to *map[string]string) {
	if from == nil || to == nil {
		return
	}

	if *to == nil {
		*to = map[string]string{}
	}

	for key, value := range from {
		if strings.HasPrefix(key, "runai/") || strings.HasPrefix(key, "run.ai/") {
			(*to)[key] = value
		}
	}
}

func copyRunaiAnnotations(from map[string]string, to *map[string]string) {
	if from == nil || to == nil {
		return
	}

	if *to == nil {
		*to = map[string]string{}
	}

	for key, value := range from {
		if strings.HasPrefix(key, "run.ai/") || strings.HasPrefix(key, "attributes.run.ai/") {
			(*to)[key] = value
		}
	}
}

func IsManuallyOverridden(meta metav1.Object) bool {
	return strings.ToLower(meta.GetLabels()[ResourceManualOverrideLabel]) == "true"
}

// ShouldPatchAnnotations reports whether the patch contains an update to the annotations.
func ShouldPatchAnnotations(annotations map[string]string, patch map[string]interface{}) bool {
	for key, patchValue := range patch {
		curValue, ok := annotations[key]
		if patchValue == nil && ok {
			return true
		}
		if patchValue != nil && !ok {
			return true
		}
		if patchValue != nil && ok && patchValue != curValue {
			return true
		}
	}
	return false
}
