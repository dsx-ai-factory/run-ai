package controller_utils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ctrl "sigs.k8s.io/controller-runtime"
)

// CreateOrUpdateIfNeeded creates or updates the given resource,
// upon a successful operation, resource is updated
func CreateOrUpdateWithOptionsIfNeeded(ctx context.Context, k8sClient client.Client,
	logger logr.Logger, resource client.Object, createOptions client.CreateOptions, updateOption client.UpdateOptions) error {
	// Creating a new instance of the resource type to allocate
	// new memory for the Get method to populate
	existingResource := reflect.New(reflect.TypeOf(resource).Elem()).Interface().(client.Object)
	found := true

	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}, existingResource)
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}

	if found {
		if specsDiffer(existingResource, resource) {
			logger.Info(fmt.Sprintf("updating resource [%s/%s/%s]",
				resource.GetNamespace(),
				utils.ResourceStructType(resource),
				resource.GetName()))
			resource.SetResourceVersion(existingResource.GetResourceVersion())
			err = k8sClient.Update(ctx, resource, &updateOption)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info(fmt.Sprintf("creating resource [%s/%s/%s]",
			resource.GetNamespace(),
			utils.ResourceStructType(resource),
			resource.GetName()))
		err = k8sClient.Create(ctx, resource, &createOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateOrUpdateIfNeeded creates or updates the given resource,
// upon a successful operation, resource is updated
func CreateOrUpdateIfNeeded(k8sClient client.Client, ctx context.Context,
	logger logr.Logger, resource client.Object) error {
	return CreateOrUpdateWithOptionsIfNeeded(ctx, k8sClient, logger, resource, client.CreateOptions{}, client.UpdateOptions{})
}

// specsDiffer reports whether the operator should issue an Update by comparing
// only the Spec field using equality.Semantic.DeepEqual. Everything else on
// the object is intentionally ignored because Spec is the only thing the
// operator authoritatively owns and writes via client.Update.
//
// The motivating problem: server-controlled fields (resourceVersion,
// managedFields, status, uid, creationTimestamp) change on every concurrent
// write by other managers — most notably engine-operator's ~1/sec status writes
// to engine-config. A full-object DeepEqual treats those as a diff every
// reconcile, triggers a no-op Update, and the Update races engine-operator's
// next write into a 409 Conflict.
//
// Semantic.DeepEqual (not reflect.DeepEqual) is needed because K8s types like
// resource.Quantity have multiple equivalent internal representations with
// custom Equal() methods that only Semantic honors; engine-config's Spec is
// full of Quantity values, so reflect.DeepEqual would still always diff.
//
// Trade-off worth knowing: writable metadata the operator does set (labels,
// annotations, ownerReferences, finalizers) is also not compared. If any of
// those drift on the live object (manual kubectl edit, third party stripping
// them, etc.) the operator will NOT notice and will NOT restore them on its
// own. If you see operator-managed resources missing the labels / annotations /
// ownerReferences the operator is supposed to set, or carrying finalizers it
// did not add, that is this trade-off surfacing — we should find the root
// cause and re-assess this solution accordingly.
//
// For resource kinds without a Spec field (ConfigMap, Secret, RBAC) we fall
// back to whole-object Semantic.DeepEqual; those kinds lack a fast concurrent
// writer like engine-operator, so the spurious-Update problem is less severe
// there.
func specsDiffer(existing, desired client.Object) bool {
	existingSpec, hasSpec := specField(existing)
	if !hasSpec {
		return !equality.Semantic.DeepEqual(existing, desired)
	}
	desiredSpec, _ := specField(desired)
	return !equality.Semantic.DeepEqual(existingSpec, desiredSpec)
}

// specField returns obj.Spec via reflection, or (nil, false) if the type has no
// Spec field (ConfigMap, Secret, Role, RoleBinding, ServiceAccount, etc.).
func specField(obj client.Object) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	spec := v.FieldByName("Spec")
	if !spec.IsValid() {
		return nil, false
	}
	return spec.Interface(), true
}

// CreateOrUpdateResources creates or updates a slice of resources
func CreateOrUpdateResources(k8sClient client.Client, ctx context.Context,
	resources []operands.Resource, runaiConfig *runaiv1.RunaiConfig, cluster *runaiv1.Cluster, ownResources bool) error {
	logger := log.FromContext(ctx, "action", "CreateOrUpdateResources")
	errs := me.New()
	var wg sync.WaitGroup
	for i := range resources {
		wg.Add(1)
		go func(resource operands.Resource) {
			defer wg.Done()
			resourceObject, err := resource.Object(ctx)
			if err != nil {
				errs.Append(err)
				return
			}

			err = SetOwnerAnnotations(runaiConfig, []client.Object{resourceObject})
			if err != nil {
				errs.Append(fmt.Errorf("owner annotations setting error: %v", err))
			}

			if ownResources {
				err = SetOwnerReferenceForApplicableResources(cluster, []client.Object{resourceObject}, k8sClient.Scheme())
				if err != nil {
					errs.Append(fmt.Errorf("failed setting cluster as owner of resources: %v", err))
				}

				err = SetControllerReferenceForApplicableResources(runaiConfig, []client.Object{resourceObject}, k8sClient.Scheme())
				if err != nil {
					errs.Append(fmt.Errorf("failed setting runaiconfig as owner of resources: %v", err))
				}
			}

			logger.Info(fmt.Sprintf("handling [%s] operand resource [%s/%s/%s]",
				resource.OperandName(), resourceObject.GetNamespace(),
				utils.ResourceStructType(resourceObject),
				resourceObject.GetName()))

			err = CreateOrUpdateIfNeeded(k8sClient, ctx, logger, resourceObject)
			if err != nil {
				if meta.IsNoMatchError(err) {
					logger.Info(fmt.Sprintf("resource [%s/%s/%s] CRD not installed, gracefully skipping create/update",
						resourceObject.GetNamespace(),
						utils.ResourceStructType(resourceObject),
						resourceObject.GetName()))
				} else {
					logger.Error(err, fmt.Sprintf("failed to handle [%s] operand resource [%s/%s/%s]: %v",
						resource.OperandName(), resourceObject.GetNamespace(),
						utils.ResourceStructType(resourceObject),
						resourceObject.GetName(), err))
					errs.Append(err)
				}
			}
		}(resources[i])
	}

	logger.Info("waiting for resources to be created or updated")
	wg.Wait()
	logger.Info("all resources were created or updated")

	if !errs.Empty() {
		return errs
	}

	return nil
}

func DeleteResources(k8sClient client.Client, ctx context.Context, resources []operands.Resource) error {
	logger := log.FromContext(ctx, "action", "DeleteResources")
	errs := me.New()
	var wg sync.WaitGroup
	for i := range resources {
		wg.Add(1)
		go func(resource operands.Resource) {
			defer wg.Done()

			resourceObject, err := resource.Object(ctx)
			if err != nil {
				errs.Append(err)
				return
			}

			annotations := resourceObject.GetAnnotations()
			if annotations[runaiv1.DeletionPolicyAnnotationKey] == string(runaiv1.DeletionPolicyRetain) {
				return
			}

			existingResource := resourceObject.DeepCopyObject().(client.Object)
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(resourceObject), existingResource)
			if err != nil {
				if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
					logger.Info(fmt.Sprintf("resource [%s/%s/%s] not found, gracefully skipping deletion",
						resourceObject.GetNamespace(),
						utils.ResourceStructType(resourceObject),
						resourceObject.GetName()))
					return
				}
				errs.Append(err)
				return
			}

			if existingResource != nil {
				logger.Info(fmt.Sprintf("deleting [%s] operand resource [%s/%s/%s]",
					resource.OperandName(), resourceObject.GetNamespace(),
					utils.ResourceStructType(resourceObject),
					resourceObject.GetName()))
				err = k8sClient.Delete(ctx, resourceObject)
				if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
					logger.Error(err, fmt.Sprintf("failed to delete [%s] operand resource [%s/%s/%s]",
						resource.OperandName(), resourceObject.GetNamespace(),
						utils.ResourceStructType(resourceObject),
						resourceObject.GetName()))
					errs.Append(err)
					return
				}
			}
		}(resources[i])
	}

	logger.Info("waiting for resources to be deleted")
	wg.Wait()
	logger.Info("all resources were deleted")

	return errs.ErrorOrNil()
}

func GetRunaiConfigForRequest(k8sReader client.Reader,
	ctx context.Context, req ctrl.Request) (*runaiv1.RunaiConfig, error) {
	runaiConfig := &runaiv1.RunaiConfig{}
	err := k8sReader.Get(ctx, req.NamespacedName, runaiConfig)
	if err != nil {
		return nil, err
	}

	return runaiConfig, nil
}

func GetClusterResource(k8sReader client.Reader, ctx context.Context) (*runaiv1.Cluster, error) {
	cluster := &runaiv1.Cluster{}
	err := k8sReader.Get(ctx, client.ObjectKey{
		Name: runaiv1.ClusterName,
	}, cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func ExtractRunaiConfigForRead(ctx context.Context, k8sClient client.Reader) (*runaiv1.RunaiConfig, error) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfigObjectKey := client.ObjectKey{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}
	err := k8sClient.Get(ctx, runaiConfigObjectKey, runaiConfig)
	if err != nil {
		return nil, err
	}
	runaiConfigWithDefaultsSet := runaiConfig.DeepCopy()
	// Safe-guard to prevent consumers of runaiConfigWithDefaultsSet to update it
	runaiConfigWithDefaultsSet.SetResourceVersion("")
	runaiConfigWithDefaultsSet.SetDefaultsWhereNeeded()
	return runaiConfigWithDefaultsSet, nil
}

func ExtractRemoteConfig(ctx context.Context, k8sClient client.Reader) (*runaiv1.RemoteConfig, error) {
	remoteConfig := &runaiv1.RemoteConfig{}
	remoteConfigObjectKey := client.ObjectKey{
		Name: runaiv1.RemoteConfigName,
	}
	err := k8sClient.Get(ctx, remoteConfigObjectKey, remoteConfig)
	if err != nil {
		return nil, err
	}
	return remoteConfig, nil
}

func GetRemoteConfigForRequest(ctx context.Context, k8sReader client.Reader, req ctrl.Request) (*runaiv1.RemoteConfig, error) {
	remoteConfigResource := &runaiv1.RemoteConfig{}
	err := k8sReader.Get(ctx, req.NamespacedName, remoteConfigResource)
	if err != nil {
		return nil, err
	}
	remoteConfigResource.SetDefaultsWhereNeeded()
	return remoteConfigResource, nil
}

func GetRemoteConfigSyncForRequest(ctx context.Context, k8sReader client.Reader, req ctrl.Request) (*runaiv1.RemoteConfigSync, error) {
	remoteConfigSyncResource := &runaiv1.RemoteConfigSync{}
	err := k8sReader.Get(ctx, req.NamespacedName, remoteConfigSyncResource)
	if err != nil {
		return nil, err
	}
	remoteConfigSyncResource.SetDefaultsWhereNeeded()
	return remoteConfigSyncResource, nil
}

func IsRemoteConfigSyncTerminated(rcs *runaiv1.RemoteConfigSync) bool {
	terminatedCond := rcs.Status.GetCondition(runaiv1.ConditionTypeConfigTerminated)
	return terminatedCond != nil && terminatedCond.Status == metav1.ConditionTrue
}

func NewResource(resource client.Object) client.Object {
	return reflect.New(reflect.TypeOf(resource).Elem()).Interface().(client.Object)
}

func ObjectMetaFromRuntimeObject(obj runtime.Object) (*metav1.ObjectMeta, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	type metaHaver struct {
		ObejctMeta metav1.ObjectMeta `json:"metadata"`
	}

	metaHolder := &metaHaver{}
	err = json.Unmarshal(jsonData, metaHolder)
	if err != nil {
		return nil, err
	}

	return &metaHolder.ObejctMeta, nil
}

// GetObjectsForResources extracts the client.Object from each operands.Resource in the given slice.
// It returns a slice of client.Object and an aggregated error if any resource fails to provide its object.
func GetObjectsForResources(ctx context.Context, resources []operands.Resource) ([]client.Object, error) {
	logger := log.FromContext(ctx, "action", "GetObjectsForResources")
	type result struct {
		obj client.Object
		err error
	}

	resultsCh := make(chan result, len(resources))
	var wg sync.WaitGroup

	for _, resource := range resources {
		wg.Add(1)
		go func(res operands.Resource) {
			defer wg.Done()
			obj, err := res.Object(ctx)
			if err != nil {
				resultsCh <- result{nil, fmt.Errorf("failed to get object for a resource in operand %s: %w", res.OperandName(), err)}
			}

			resultsCh <- result{obj, nil}
		}(resource)
	}

	go func() {
		logger.Info("waiting for resources to be process objects")
		wg.Wait()
		logger.Info("all resource objects were processed")
		close(resultsCh)
	}()

	objects := make([]client.Object, 0, len(resources))
	errs := me.New()

	for res := range resultsCh {
		if res.err != nil {
			errs.Append(res.err)
		} else if res.obj != nil {
			objects = append(objects, res.obj)
		}
	}

	return objects, errs.ErrorOrNil()
}
