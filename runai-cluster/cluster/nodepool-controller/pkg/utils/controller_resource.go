package utils

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SetControllerReference adds an owner as a controller owner reference to resource.
func SetControllerReference(owner, resource client.Object, scheme *runtime.Scheme) error {
	return controllerutil.SetControllerReference(owner, resource, scheme)
}

// CreateOrUpdateIfNeeded creates the resource, or updates it when its spec drifted.
func CreateOrUpdateIfNeeded(k8sClient client.Client, ctx context.Context,
	logger logr.Logger, resource client.Object) error {
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

	if !found {
		logger.Info(fmt.Sprintf("creating resource [%s/%s/%s]",
			resource.GetNamespace(), resourceStructType(resource), resource.GetName()))
		return k8sClient.Create(ctx, resource)
	}

	if specsDiffer(existingResource, resource) {
		logger.Info(fmt.Sprintf("updating resource [%s/%s/%s]",
			resource.GetNamespace(), resourceStructType(resource), resource.GetName()))
		resource.SetResourceVersion(existingResource.GetResourceVersion())
		return k8sClient.Update(ctx, resource)
	}

	return nil
}

// specsDiffer compares only the Spec field, because Spec is the only thing the
// controller authoritatively owns and writes. Comparing the whole object would
// diff on server-controlled fields (resourceVersion, managedFields, status) that
// other writers change constantly, causing no-op Updates that race into 409s.
//
// equality.Semantic is required rather than reflect.DeepEqual because types like
// resource.Quantity have several equivalent internal representations and only
// Semantic honours their Equal() methods.
//
// Kinds without a Spec (ConfigMap, Secret, RBAC) fall back to whole-object
// comparison; they have no fast concurrent writer so spurious Updates are benign.
func specsDiffer(existing, desired client.Object) bool {
	existingSpec, hasSpec := specField(existing)
	if !hasSpec {
		return !equality.Semantic.DeepEqual(existing, desired)
	}
	desiredSpec, _ := specField(desired)
	return !equality.Semantic.DeepEqual(existingSpec, desiredSpec)
}

// specField returns obj.Spec via reflection, or (nil, false) if the type has no Spec field.
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

// DeleteResources deletes the given operand resources concurrently, skipping any
// resource annotated with the Retain deletion policy.
func DeleteResources(k8sClient client.Client, ctx context.Context, resources []operands.Resource) error {
	logger := log.FromContext(ctx, "action", "DeleteResources")

	var (
		wg      sync.WaitGroup
		errsMu  sync.Mutex
		errsAcc error
	)
	appendErr := func(err error) {
		errsMu.Lock()
		defer errsMu.Unlock()
		errsAcc = multierror.Append(errsAcc, err)
	}

	for i := range resources {
		wg.Add(1)
		go func(resource operands.Resource) {
			defer wg.Done()

			resourceObject, err := resource.Object(ctx)
			if err != nil {
				appendErr(err)
				return
			}

			existingResource := resourceObject.DeepCopyObject().(client.Object)
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(resourceObject), existingResource)
			if err != nil {
				if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
					logger.Info(fmt.Sprintf("resource [%s/%s/%s] not found, gracefully skipping deletion",
						resourceObject.GetNamespace(), resourceStructType(resourceObject), resourceObject.GetName()))
					return
				}
				appendErr(err)
				return
			}

			logger.Info(fmt.Sprintf("deleting [%s] operand resource [%s/%s/%s]",
				resource.OperandName(), resourceObject.GetNamespace(),
				resourceStructType(resourceObject), resourceObject.GetName()))
			err = k8sClient.Delete(ctx, resourceObject)
			if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
				logger.Error(err, fmt.Sprintf("failed to delete [%s] operand resource [%s/%s/%s]",
					resource.OperandName(), resourceObject.GetNamespace(),
					resourceStructType(resourceObject), resourceObject.GetName()))
				appendErr(err)
			}
		}(resources[i])
	}

	logger.Info("waiting for resources to be deleted")
	wg.Wait()
	logger.Info("all resources were deleted")

	return errsAcc
}

func resourceStructType(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}
