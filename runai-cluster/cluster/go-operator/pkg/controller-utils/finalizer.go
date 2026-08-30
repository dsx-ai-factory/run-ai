package controller_utils

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"

	json_patch "github.com/run-ai/runai/runai-cluster/common/utils/json-patch"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finalizersKey = "/metadata/finalizers"
)

// HandleRunAIFinalizer adds the runai finalizer to a resource,
// or executes finalizer logic if the resource is set to be deleted.
// The passed in resource pointer will be updated after the finalizer
// assignment or removal.
func HandleRunAIFinalizer(ctx context.Context, managementClusterClient client.Client, remoteClusterClient client.Client,
	runaiConfig *runv1.RunaiConfig,
	resourcesToCleanup []operands.Resource,
	disableFinalizer bool) (bool, error) {
	logger := log.FromContext(ctx, "action", "HandleRunAIFinalizer")

	if !runaiConfig.GetDeletionTimestamp().IsZero() {
		// Resource is set to be deleted, finalize
		logger.Info(fmt.Sprintf("[%s/%s/%s] was scheduled for deletion, finalizing...",
			runaiConfig.GetNamespace(), utils.ResourceStructType(runaiConfig),
			runaiConfig.GetName()))

		err := Finalize(ctx, remoteClusterClient, resourcesToCleanup, logger)
		if err != nil {
			return false, err
		}

		logger.Info(fmt.Sprintf("[%s/%s/%s] finalized",
			runaiConfig.GetNamespace(), utils.ResourceStructType(runaiConfig),
			runaiConfig.GetName()))

		logger.Info(fmt.Sprintf("removing finalizer from [%s/%s/%s]",
			runaiConfig.GetNamespace(), utils.ResourceStructType(runaiConfig),
			runaiConfig.GetName()))

		// Remove finalizer once finalizer logic is complete
		RemoveFinalizer(runaiConfig, runv1.Finalizer)

		err = PatchRunaiConfigFinalizers(ctx, managementClusterClient, runaiConfig)
		if err != nil {
			return false, fmt.Errorf("failed removing RunaiConfig finalizers: %v", err)
		}

		return true, nil
	} else {
		originalFinalizers := runaiConfig.GetFinalizers()

		if disableFinalizer {
			RemoveFinalizer(runaiConfig, runv1.Finalizer)
		} else {
			AddFinalizer(runaiConfig, runv1.Finalizer)
		}

		if !slices.Equal(originalFinalizers, runaiConfig.GetFinalizers()) {
			logger.Info(fmt.Sprintf("updating RunaiConfig finalizers: %v -> %v", originalFinalizers, runaiConfig.GetFinalizers()))
			err := PatchRunaiConfigFinalizers(ctx, managementClusterClient, runaiConfig)
			if err != nil {
				return false, fmt.Errorf("failed patching RunaiConfig finalizers: %v", err)
			}
		}
	}

	return false, nil
}

// finalize the RunaiConfig cleanup
func Finalize(ctx context.Context, kubeClient client.Client, resources []operands.Resource,
	logger logr.Logger) error {
	merr := me.New()

	err := deleteCrs(ctx, kubeClient, logger)
	if err != nil {
		logger.Error(err, "failed deleting crds")
		merr.Append(err)
		return merr
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(resources))

	// Clean up all resources owned by RunaiConfig concurrently
	for _, resource := range resources {
		wg.Add(1)
		go func(resource operands.Resource) {
			defer wg.Done()
			resourceObject, err := resource.Object(ctx)
			if err != nil {
				logger.Error(err, "error getting resource object", "operandName", resource.OperandName())
				// If we can't get the object, treat as error and return
				errCh <- err
				return
			}
			logger.Info(fmt.Sprintf("deleting resource [%s/%s/%s] for operand [%s]",
				resourceObject.GetNamespace(), utils.ResourceStructType(resourceObject),
				resourceObject.GetName(), resource.OperandName()))

			annotations := resourceObject.GetAnnotations()
			if annotations[runv1.DeletionPolicyAnnotationKey] == string(runv1.DeletionPolicyRetain) {
				return
			}

			err = kubeClient.Delete(ctx, resourceObject)
			if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
				errCh <- err
				return
			}
		}(resource)
	}

	logger.Info("waiting for resources to be finalized")
	wg.Wait()
	logger.Info("all resources were finalized")
	close(errCh)

	// Return the first error encountered, if any
	for err := range errCh {
		if err != nil {
			merr.Append(err)
		}
	}

	return merr.ErrorOrNil()
}

func HasFinalizer(resource client.Object, expectedFinalizer string) bool {
	finalizers := resource.GetFinalizers()

	for _, finalizer := range finalizers {
		if finalizer == expectedFinalizer {
			return true
		}
	}

	return false
}

func ShouldFinalizeRunaiConfig(runaiConfig *runv1.RunaiConfig) bool {
	return !runaiConfig.GetDeletionTimestamp().IsZero() && slices.Contains(runaiConfig.GetFinalizers(), runv1.Finalizer)
}

func AddFinalizer(resource client.Object, finalizerToAdd string) {
	finalizers := resource.GetFinalizers()

	found := false
	for _, finalizer := range finalizers {
		if finalizer == finalizerToAdd {
			found = true
			break
		}
	}

	if !found {
		resource.SetFinalizers(append(resource.GetFinalizers(), finalizerToAdd))
	}
}

func RemoveFinalizer(resource client.Object, finalizerToRemove string) {
	finalizers := resource.GetFinalizers()
	finalizersWithoutFinalizer := []string{}

	for _, finalizer := range finalizers {
		if finalizer == finalizerToRemove {
			continue
		}

		finalizersWithoutFinalizer = append(finalizersWithoutFinalizer, finalizer)
	}

	resource.SetFinalizers(finalizersWithoutFinalizer)
}

func PatchRunaiConfigFinalizers(ctx context.Context, kubeClient client.Client, runaiConfig *runv1.RunaiConfig) error {
	finalizerPatch := json_patch.NewPatch()
	finalizerPatch.AddOperations(json_patch.NewReplaceOperation(finalizersKey,
		runaiConfig.GetFinalizers()))
	rawFinalizerPatch, err := finalizerPatch.Bytes()
	if err != nil {
		return fmt.Errorf("finalizer patch build error: %v", err)
	}

	err = kubeClient.Patch(ctx, runaiConfig,
		client.RawPatch(types.JSONPatchType, rawFinalizerPatch))
	if err != nil {
		return fmt.Errorf("finalizer patch error: %v", err)
	}

	return nil
}

func deleteCrs(ctx context.Context, kubeClient client.Client,
	logger logr.Logger) error {
	var err error
	innerErr := deleteNodePools(ctx, kubeClient, logger)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = deleteAccessRules(ctx, kubeClient, logger)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	return err
}

func deleteAllOfResource(ctx context.Context, kubeClient client.Client,
	logger logr.Logger, obj client.Object) error {
	err := kubeClient.DeleteAllOf(ctx, obj)
	if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		logger.Error(err, "failed deleting all resources",
			"kind", reflect.TypeOf(obj).Elem().Name())
		return err
	}
	return nil
}

func deleteNodePools(ctx context.Context, kubeClient client.Client,
	logger logr.Logger) error {
	nodePool := &v1alpha1.NodePool{}
	return deleteAllOfResource(ctx, kubeClient, logger, nodePool)
}

func deleteAccessRules(ctx context.Context, kubeClient client.Client,
	logger logr.Logger) error {
	accessRule := &v1alpha1.AccessRule{}
	return deleteAllOfResource(ctx, kubeClient, logger, accessRule)
}
