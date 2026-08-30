package handlers

import (
	"context"
	new_error "errors"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/resourcing"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeletionPredicate func(object client.Object) bool

type CommonHandler struct {
	backendClient     *runaiBackendClient.BackendClient
	runaiClient       client.Client
	cachedClient      client.Reader
	gvk               schema.GroupVersionKind
	objectInspector   ObjectInspector
	deletionPredicate DeletionPredicate
	HandlingOptions
	isUnstructured       bool
	expectedObjectGetter ExpectedObjectGetter
}

type HandlingOptions struct {
	deletePropagationOrphan bool
}

type ResourceHandler interface {
	Handle() error
	MetricLabel() string
}

type ResourceHandlerInitFunc func(b *runaiBackendClient.BackendClient, r client.Client, c client.Reader) (ResourceHandler, error)

type ObjectInspector interface {
	DeepCopySpec(from, to client.Object)

	IsUpToDate(left, right client.Object) bool
}

// ExpectedObjectGetter is an interface for getting expected objects.
// Sometimes, if the parsing from backend object to cluster object is a heavy operation,
// we want to do it only when we actually need to create or update the object in the cluster.
type ExpectedObjectGetter interface {
	// GetExpectedObject
	// params: name, namespace - of the object we want to get
	// returns: the expected object to create/update, and error if occurred
	GetExpectedObject(name, namespace string) (client.Object, error)

	// ShouldReCreate
	// params: expected - the expected object to create/update
	//         existing - the existing object in the cluster
	// returns: true if we should delete and re-create the object, false if we can update it in-place
	ShouldReCreate(expected, existing client.Object) bool
}

// GVK = group-version-kind (schema.GroupVersionKind)
// a convenient way to create a generic object with a type.

func NewCommonHandler(backendClient *runaiBackendClient.BackendClient,
	runaiClient client.Client, cachedClient client.Reader,
	object client.Object, objectInspector ObjectInspector, handlingOptions HandlingOptions) (*CommonHandler, error) {
	if backendClient == nil || runaiClient == nil || object == nil {
		log.Error().Msgf("Bad params for NewCommonHandler")
		return nil, fmt.Errorf("bad params for NewCommonHandler")
	}

	gvk, err := getGvkForObject(object, runaiClient.Scheme())
	if err != nil {
		log.Error().Msgf("Failed to get GVK for object <%s>, error: %s",
			object.GetObjectKind().GroupVersionKind().String(), err.Error())
		return nil, err
	}

	return &CommonHandler{
		backendClient:        backendClient,
		runaiClient:          runaiClient,
		cachedClient:         cachedClient,
		gvk:                  gvk,
		objectInspector:      objectInspector,
		HandlingOptions:      handlingOptions,
		deletionPredicate:    func(object client.Object) bool { return true },
		expectedObjectGetter: nil,
	}, nil
}

func NewUnstructuredHandler(backendClient *runaiBackendClient.BackendClient,
	runaiClient client.Client, cachedClient client.Reader,
	object client.Object, objectInspector ObjectInspector, handlingOptions HandlingOptions) (*CommonHandler, error) {
	if backendClient == nil || runaiClient == nil || object == nil {
		log.Error().Msgf("Bad params for NewCommonHandler")
		return nil, fmt.Errorf("bad params for NewCommonHandler")
	}

	return &CommonHandler{
		backendClient:     backendClient,
		runaiClient:       runaiClient,
		cachedClient:      cachedClient,
		objectInspector:   objectInspector,
		HandlingOptions:   handlingOptions,
		isUnstructured:    true,
		deletionPredicate: func(object client.Object) bool { return true },
	}, nil
}

// CreateOrUpdateObjectsIgnoreConflicts creates or updates the 'objects' in the cluster.
// this function update existing only if label "conflictLabel" is the same in both original and existing, else, existing is not changed
func (h *CommonHandler) CreateOrUpdateObjectsIgnoreConflicts(objects []client.Object, conflictLabel string) (map[string]bool, []error, error) {
	existingObjectNames := map[string]bool{}
	var err error = nil
	perObjectError := make([]error, len(objects))

	for idx, clusterObject := range objects {
		innerErr := h.CreateOrUpdateObject(clusterObject, conflictLabel)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			perObjectError[idx] = multierror.Append(perObjectError[idx], innerErr)

			log.Error().Msgf("Failed to reconcile %s <%s>, error: %s",
				h.objectTypeStr(), clusterObject.GetName(), err.Error())
			continue
		}

		existingObjectNames[clusterObject.GetName()] = true
	}

	return existingObjectNames, perObjectError, err
}

// CreateOrUpdateObjects creates or updates the 'objects' in the cluster.
func (h *CommonHandler) CreateOrUpdateObjects(objects []client.Object) (map[string]bool, []error, error) {
	return h.CreateOrUpdateObjectsIgnoreConflicts(objects, "")
}

// ProjectToNamespaceList returns a map from project ID to namespace.
func (h *CommonHandler) ProjectToNamespaceList() (map[string]string, error) {
	projectIdToNamespace := map[string]string{}

	projectList := &kaiorg.ProjectList{}
	err := h.cachedClient.List(context.Background(), projectList)
	if err != nil {
		return nil, err
	}

	for _, project := range projectList.Items {
		if project.Status.Namespace != "" {
			projectId, found := project.Labels[runaiv2.ProjectIdLabel]
			if found {
				projectIdToNamespace[projectId] = project.Status.Namespace
			} else {
				log.Warn().Msgf("Project <%s> does not have a project ID label", project.Name)
			}
		}
	}

	return projectIdToNamespace, nil
}

// ReconcileObjects creates or updates the 'objects' in the cluster;
// then syncs the objects that should be deleted - that are found in the cluster, but are not a part of 'objects' list.
func (h *CommonHandler) ReconcileObjects(objects []client.Object, skipDeletedSync bool) error {
	existingObjectNames, _, err := h.CreateOrUpdateObjects(objects)
	if err != nil {
		log.Error().Msgf("Had an error while creating or updating - skipping the sync of deleted %ss",
			h.objectTypeStr())
		return err
	}

	if skipDeletedSync {
		return nil
	}

	err = h.syncDeletedObjects(existingObjectNames)
	if err != nil {
		log.Error().Msgf("Failed to sync deleted %ss in cluster, error: %s", h.objectTypeStr(), err.Error())
		return err
	}

	log.Debug().Msgf("Successfully handled %ss", h.objectTypeStr())
	return nil
}

// DeleteObjects deletes the objects if needed.
// if namespaced is set (!= ""), the objects are expected to be namespaced.
func (h *CommonHandler) DeleteObjects(objectNames []types.NamespacedName) ([]error, error) {
	var err error
	perObjectError := make([]error, len(objectNames))

	for idx, objectName := range objectNames {
		object, innerErr := h.getObject(objectName.Name, objectName.Namespace)
		if innerErr != nil {
			if !errors.IsNotFound(innerErr) {
				log.Error().Msgf("Failed to get %s <%s/%s>, error: %s", h.objectTypeStr(), objectName.Namespace, objectName.Name, innerErr.Error())
				err = multierror.Append(err, innerErr)
				perObjectError[idx] = multierror.Append(perObjectError[idx], innerErr)
			} else {
				log.Debug().Msgf("%s <%s/%s> is already deleted from the cluster - doing nothing",
					h.objectTypeStr(), objectName.Namespace, objectName.Name)

				perObjectError[idx] = errors.NewNotFound(schema.GroupResource{Resource: h.gvk.Kind}, objectName.Name)
			}
			continue
		}

		innerErr = h.deleteObject(object)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			perObjectError[idx] = multierror.Append(perObjectError[idx], innerErr)
			h.createDeletionEvent(object, innerErr)
		}
	}

	return perObjectError, err
}

// DeleteObjectsById deletes the objects if needed also based on the id.
// if namespaced is set (!= ""), the objects are expected to be namespaced.
func (h *CommonHandler) DeleteObjectsById(objectsIdsToName map[string]types.NamespacedName, idLabel string) (map[string]error, error) {
	var err error
	perObjectError := make(map[string]error, len(objectsIdsToName))

	for objectId, objectName := range objectsIdsToName {
		object, innerErr := h.getObject(objectName.Name, objectName.Namespace)
		if innerErr != nil {
			if !errors.IsNotFound(innerErr) {
				log.Error().Msgf("Failed to get %s <%s/%s>, error: %s", h.objectTypeStr(), objectName.Namespace, objectName.Name, innerErr.Error())
				err = multierror.Append(err, innerErr)
				perObjectError[objectId] = innerErr
			} else {
				log.Debug().Msgf("%s <%s/%s> is already deleted from the cluster - doing nothing",
					h.objectTypeStr(), objectName.Namespace, objectName.Name)

				perObjectError[objectId] = errors.NewNotFound(schema.GroupResource{Resource: h.gvk.Kind}, objectName.Name)
			}
			continue
		}
		id, found := object.GetLabels()[idLabel]
		if !found || id == "" {
			errorMsg := fmt.Sprintf("%s <%s/%s> cannot get id from label %s",
				h.objectTypeStr(), objectName.Namespace, objectName.Name, idLabel)
			log.Debug().Msgf("%s", errorMsg)

			perObjectError[objectId] = errors.NewInternalError(new_error.New(errorMsg))
			continue
		}
		if id != objectId {
			log.Debug().Msgf("%s <%s/%s> did not delete object since cluster object id %s is different than deleted object id %s",
				h.objectTypeStr(), objectName.Namespace, objectName.Name, id, objectId)

			continue
		}
		innerErr = h.deleteObject(object)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			perObjectError[objectId] = innerErr
			h.createDeletionEvent(object, innerErr)
		}
	}

	return perObjectError, err
}

// createDeletionEvent creates a Kubernetes event for the failed deletion
func (h *CommonHandler) createDeletionEvent(object client.Object, objectErr error) {
	involvedObject := corev1.ObjectReference{
		Kind:      object.GetObjectKind().GroupVersionKind().Kind,
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
		UID:       object.GetUID(),
	}

	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: object.GetNamespace(),
			Name:      object.GetName(),
		},
		InvolvedObject: involvedObject,
		Reason:         "Deletion failed",
		Message:        fmt.Sprintf("Failed to delete %s: %s", involvedObject.Kind, objectErr.Error()), // Use the message field available in corev1.Event
		Type:           corev1.EventTypeWarning,
		Source: corev1.EventSource{
			Component: "runai-agent",
		},
		LastTimestamp: metav1.Now(),
	}

	if err := h.runaiClient.Create(context.Background(), event); err != nil {
		log.Error().Msgf("Failed to create event: %s", err.Error())
	}
}

func (h *CommonHandler) CreateOrUpdateObject(objectToReconcile client.Object, conflictLabel string) error {
	log.Debug().Msgf("Reconciling %s %s", h.objectTypeStr(), objectToReconcile.GetName())

	var existingObject client.Object
	var err error

	if h.isUnstructured {
		clientObjectKey := client.ObjectKey{Name: objectToReconcile.GetName(), Namespace: objectToReconcile.GetNamespace()}

		u := unstructured.Unstructured{}

		u.SetGroupVersionKind(objectToReconcile.GetObjectKind().GroupVersionKind())
		err = h.cachedClient.Get(context.Background(), clientObjectKey, &u)
		if err == nil {
			existingObject = &u
		}
	} else {
		existingObject, err = h.getObject(objectToReconcile.GetName(), objectToReconcile.GetNamespace())
	}

	if err != nil {
		if errors.IsNotFound(err) {
			return h.handleNonExisting(objectToReconcile)
		}
		log.Error().Msgf("Failed getting %s <%s>, error: %s",
			h.objectTypeStr(), objectToReconcile.GetName(), err.Error())
		return err
	}

	return h.handleExisting(objectToReconcile, existingObject, conflictLabel)
}

// SetExpectedObjectGetter sets the getter for expected object before creation.
// mainly used for objects that don't support Update, like Karta -
// so we won't have to json-parse the object every time we see it in the list;
// but only when we actually need to create it.
func (h *CommonHandler) SetExpectedObjectGetter(expectedObjectGetter ExpectedObjectGetter) {
	h.expectedObjectGetter = expectedObjectGetter
}

func (h *CommonHandler) syncDeletedObjects(existingObjectNames map[string]bool) error {
	log.Debug().Msgf("Syncing deleted %ss", h.objectTypeStr())

	allObjectsInCluster, err := h.listObjects()
	if err != nil {
		log.Error().Msgf("Failed listing %ss, error: %s", h.objectTypeStr(), err.Error())
		return err
	}

	for _, objectInCluster := range allObjectsInCluster {
		if _, found := existingObjectNames[objectInCluster.GetName()]; found {
			continue
		}
		if !h.deletionPredicate(objectInCluster) {
			continue
		}
		innerErr := h.deleteObject(objectInCluster)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

func (h *CommonHandler) handleNonExisting(objectToReconcile client.Object) error {
	log.Info().Msgf("Creating %s <%s/%s>", h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())

	if h.expectedObjectGetter != nil {
		expectedObject, err := h.expectedObjectGetter.GetExpectedObject(objectToReconcile.GetName(), objectToReconcile.GetNamespace())
		if err != nil {
			log.Error().Msgf("Failed to get expected object for %s <%s/%s>, error: %s",
				h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName(), err.Error())
			return err
		}
		objectToReconcile = expectedObject
	}

	err := h.runaiClient.Create(context.Background(), objectToReconcile)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error().Msgf("Failed creating %s <%s>, error: %s",
			h.objectTypeStr(), objectToReconcile.GetName(), err.Error())
		return err
	}
	return nil
}

func (h *CommonHandler) handleExisting(objectToReconcile, existingObject client.Object, conflictLabel string) error {
	if h.objectInspector.IsUpToDate(objectToReconcile, existingObject) {
		log.Debug().Msgf("%s <%s> in cluster is up to date", h.objectTypeStr(), existingObject.GetName())
		return nil
	}

	if conflictLabel != "" && objectToReconcile.GetLabels()[conflictLabel] != existingObject.GetLabels()[conflictLabel] {
		return errors.NewInternalError(fmt.Errorf("%s <%s> exists in cluster from a different source", h.objectTypeStr(), existingObject.GetName()))
	}

	if h.expectedObjectGetter != nil {
		return h.handleExistingPossibleRecreation(objectToReconcile, existingObject)
	}

	// Retry the update operation with exponential backoff in case of conflicts
	log.Info().Msgf("Updating %s <%s/%s/>", h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Refetch the latest version of the object to ensure we have the most recent resourceVersion
		latestObject, getErr := h.getObjectByGVK(
			objectToReconcile.GetName(),
			objectToReconcile.GetNamespace(),
			objectToReconcile.GetObjectKind().GroupVersionKind(),
		)
		if getErr != nil {
			return getErr
		}

		// deep copy will run over the entire object, this will copy attributes that needed to be kept from the existing object.
		existingPreCopy(objectToReconcile, latestObject)

		h.objectInspector.DeepCopySpec(objectToReconcile, latestObject)

		// Try to update with the latest version
		return h.runaiClient.Update(context.Background(), latestObject)
	})

	if err != nil {
		log.Error().Msgf("Failed updating %s <%s>, error: %s",
			h.objectTypeStr(), existingObject.GetName(), err.Error())
		return err
	}
	return nil
}

func (h *CommonHandler) handleExistingPossibleRecreation(objectToReconcile, existingObject client.Object) error {
	if h.expectedObjectGetter == nil {
		return nil
	}

	expectedObject, err := h.expectedObjectGetter.GetExpectedObject(objectToReconcile.GetName(), objectToReconcile.GetNamespace())
	if err != nil {
		log.Error().Msgf("Failed to get expected object for %s <%s/%s>, error: %s",
			h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName(), err.Error())
		return err
	}
	objectToReconcile = expectedObject

	shouldReCreate := h.expectedObjectGetter.ShouldReCreate(expectedObject, existingObject)

	// Update object - either regular update, or only the labels before deletion
	// Retry the update operation with exponential backoff in case of conflicts
	log.Info().Msgf("Updating %s <%s/%s/>", h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Refetch the latest version of the object to ensure we have the most recent resourceVersion
		latestObject, getErr := h.getObjectByGVK(
			objectToReconcile.GetName(),
			objectToReconcile.GetNamespace(),
			objectToReconcile.GetObjectKind().GroupVersionKind(),
		)
		if getErr != nil {
			return getErr
		}

		if !shouldReCreate {
			// regular update - just copy the whole object
			h.objectInspector.DeepCopySpec(objectToReconcile, latestObject)
		} else {
			// will perform deletion and recreation - but do update only labels
			latestObject.SetLabels(objectToReconcile.GetLabels())
		}

		// Try to update with the latest version
		return h.runaiClient.Update(context.Background(), latestObject)
	})

	if err != nil {
		log.Error().Msgf("Failed updating %s <%s>, error: %s",
			h.objectTypeStr(), existingObject.GetName(), err.Error())
		return err
	}
	if !shouldReCreate {
		return nil
	}

	// now, delete the object, and it will be re-created later during the next reconciliation

	// Refetch one more time to get the latest state for deletion checks
	currentObject, refetchErr := h.getObjectByGVK(
		objectToReconcile.GetName(),
		objectToReconcile.GetNamespace(),
		objectToReconcile.GetObjectKind().GroupVersionKind(),
	)
	if refetchErr != nil {
		if errors.IsNotFound(refetchErr) {
			// Object was already deleted, nothing to do
			log.Debug().Msgf("%s <%s/%s> was already deleted",
				h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())
			return nil
		}
		log.Error().Msgf("Failed to refetch %s <%s/%s> for deletion check, error: %s",
			h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName(), refetchErr.Error())
		return refetchErr
	}

	if !currentObject.GetDeletionTimestamp().IsZero() {
		log.Info().Msgf("%s <%s/%s> is already being deleted, waiting for it to be deleted",
			h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())
		return nil
	}

	log.Info().Msgf("%s <%s/%s> requires re-creation, deleting it - will be created later during the next reconciliation",
		h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName())
	err = h.runaiClient.Delete(context.Background(), currentObject)
	if err != nil && !errors.IsNotFound(err) {
		log.Error().Msgf("Failed deleting %s <%s/%s>, error: %s",
			h.objectTypeStr(), objectToReconcile.GetNamespace(), objectToReconcile.GetName(), err.Error())
		return err
	}
	return nil
}

func (h *CommonHandler) listObjects() ([]client.Object, error) {
	objectList := &unstructured.UnstructuredList{}
	objectList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   h.gvk.Group,
		Kind:    h.gvk.Kind,
		Version: h.gvk.Version,
	})

	err := h.cachedClient.List(context.Background(), objectList)
	if err != nil {
		log.Error().Msgf("Failed to list %ss, error: %s", h.objectTypeStr(), err.Error())
		return []client.Object{}, err
	}

	objectItems := []client.Object{}
	err = objectList.EachListItem(func(item runtime.Object) error {
		gvkObject, innerErr := h.runaiClient.Scheme().New(h.gvk)
		if innerErr != nil {
			return innerErr
		}
		innerErr = h.runaiClient.Scheme().Convert(item, gvkObject, nil)
		if innerErr != nil {
			return innerErr
		}

		objectItems = append(objectItems, gvkObject.(client.Object))
		return nil
	})
	if err != nil {
		log.Error().Msgf("Failed to convert objects of type <%s> in objectList, error: %s",
			h.objectTypeStr(), err.Error())
		return []client.Object{}, err
	}

	return objectItems, nil
}

func (h *CommonHandler) getObject(name string, namespace string) (client.Object, error) {
	// Use the common getObjectByGVK function with the handler's GVK
	return h.getObjectByGVK(name, namespace, h.gvk)
}

// getObjectByGVK fetches an object using the provided GVK.
// For unstructured handlers, it uses the unstructured get path.
// For typed handlers, it creates a typed object from the provided GVK.
func (h *CommonHandler) getObjectByGVK(name, namespace string, gvk schema.GroupVersionKind) (client.Object, error) {
	var object client.Object
	var err error

	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if h.isUnstructured {
		// For unstructured handlers, use an unstructured.Unstructured and the unstructured get path
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		err = h.cachedClient.Get(
			context.Background(),
			namespacedName,
			u,
		)
		object = u
	} else {
		// For typed handlers, create an empty typed object from the provided GVK
		existingObject, createErr := h.createEmptyObject()
		if createErr != nil {
			return nil, createErr
		}

		err = h.cachedClient.Get(context.Background(), namespacedName, existingObject)
		object = existingObject
	}

	return object, err
}

func (h *CommonHandler) deleteObject(object client.Object) error {
	deleteOptions := &client.DeleteOptions{}
	deleteOptionsMsg := ""
	if h.deletePropagationOrphan {
		deleteOptionsMsg = "; with delete propagation = Orphan"
		propagationPolicy := metav1.DeletePropagationOrphan
		deleteOptions.PropagationPolicy = &propagationPolicy
	}

	log.Info().Msgf("%s <%s> does not exist in the control-plane anymore, deleting it from the cluster%s",
		h.objectTypeStr(), object.GetName(), deleteOptionsMsg)
	err := h.runaiClient.Delete(context.Background(), object, deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		log.Error().Msgf("Failed deleting %s <%s>, error: %s",
			h.objectTypeStr(), object.GetName(), err.Error())
		return err
	}
	return nil
}

func (h *CommonHandler) createEmptyObject() (client.Object, error) {
	existingRuntimeObject, err := h.runaiClient.Scheme().New(h.gvk)
	if err != nil {
		log.Error().Msgf("Failed creating new runtime object from gvk <%s>, error: %s",
			h.gvk.String(), err.Error())
		return nil, err
	}

	existingObject, ok := existingRuntimeObject.(client.Object)
	if !ok {
		return nil,
			fmt.Errorf("failed to parse existingRuntimeObject of kind <%s> to client.Object", h.objectTypeStr())
	}
	return existingObject, nil
}

func (h *CommonHandler) objectTypeStr() string {
	return h.gvk.Kind
}

func getGvkForObject(object runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) == 0 {
		err = fmt.Errorf("no kinds registered for object <%s> in scheme",
			object.GetObjectKind().GroupVersionKind().String())
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) > 1 {
		err = fmt.Errorf("too many kinds registered for object <%s> in scheme; make sure only one is registered",
			object.GetObjectKind().GroupVersionKind().String())
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func GetQueueName(orgUnitName, nodePoolName string) string {
	if nodePoolName == v1alpha1.DefaultNodePoolName {
		return orgUnitName
	}
	return fmt.Sprintf("%s-%s", orgUnitName, nodePoolName)
}

func ConvertNodePoolResources(nodePoolsResources orgUnitApi.ResourcesSync) *resourcing.QueueResources {
	return &resourcing.QueueResources{
		GPU:    convertGPUTypeResources(nodePoolsResources.Gpu),
		CPU:    convertCPUTypeResources(nodePoolsResources.Cpu),
		Memory: convertMemoryTypeResources(nodePoolsResources.Memory),
	}
}

func convertGPUTypeResources(r orgUnitApi.ResourcesSyncGpu) resourcing.SystemResource {
	return resourcing.SystemResource{
		Deserved:        r.Deserved,
		OverQuotaWeight: common.GetFloatValDefault(r.OverQuotaWeight.Get(), DefaultOverQuotaWeight),
		Limit:           common.GetFloatValDefault(r.Limit.Get(), NoMaxAllowed),
	}
}

func convertCPUTypeResources(r orgUnitApi.NullableResourcesCpu) resourcing.SystemResource {
	if !r.IsSet() || r.Get() == nil {
		return defaultResources()
	}

	return convertCpuMemoryNonNullResources(r.Get().Deserved, r.Get().Limit, r.Get().OverQuotaWeight)
}

func convertMemoryTypeResources(r orgUnitApi.NullableResourcesSyncMemory) resourcing.SystemResource {
	if !r.IsSet() || r.Get() == nil {
		return defaultResources()
	}
	return convertCpuMemoryNonNullResources(r.Get().Deserved, r.Get().Limit, r.Get().OverQuotaWeight)
}

func convertCpuMemoryNonNullResources(deserved orgUnitApi.NullableFloat64, limit orgUnitApi.NullableFloat64, overQuotaWeight orgUnitApi.NullableFloat64) resourcing.SystemResource {
	return resourcing.SystemResource{
		Deserved:        common.GetFloatValDefault(deserved.Get(), NoMaxAllowed),
		OverQuotaWeight: common.GetFloatValDefault(overQuotaWeight.Get(), DefaultOverQuotaWeight),
		Limit:           common.GetFloatValDefault(limit.Get(), NoMaxAllowed),
	}
}

func defaultResources() resourcing.SystemResource {
	return resourcing.SystemResource{
		Deserved:        NoMaxAllowed,
		OverQuotaWeight: DefaultOverQuotaWeight,
		Limit:           NoMaxAllowed,
	}
}

// existingPreCopy TODO when workload supports workload metadata this should be handled in the BE and not here, so this should be deleted.
func existingPreCopy(origin client.Object, existing client.Object) {
	// annotation pre-copy, this is added because the origin annotation runs over the existing annotation (not a good practice, but this is what we can do for now)
	annotationPreCopy(origin, existing)
}

func annotationPreCopy(origin client.Object, existing client.Object) {
	existingAnnotations := existing.GetAnnotations()
	if len(existingAnnotations) == 0 {
		return
	}

	name := existingAnnotations[topology.KAITopologyAnnotationKey]
	preferred := existingAnnotations[topology.KAITopologyAnnotationPreferredKey]
	required := existingAnnotations[topology.KAITopologyAnnotationRequiredKey]
	if len(name) == 0 && len(preferred) == 0 && len(required) == 0 {
		return
	}

	originAnnotations := origin.GetAnnotations()
	if originAnnotations == nil {
		originAnnotations = make(map[string]string)
	}

	if len(name) != 0 {
		originAnnotations[topology.KAITopologyAnnotationKey] = name
	}

	if len(preferred) != 0 {
		originAnnotations[topology.KAITopologyAnnotationPreferredKey] = preferred
	}

	if len(required) != 0 {
		originAnnotations[topology.KAITopologyAnnotationRequiredKey] = required
	}
	origin.SetAnnotations(originAnnotations)
}
