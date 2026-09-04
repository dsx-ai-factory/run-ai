/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package run

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kube_client "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	me "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/multi-error"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RemoteConfigReconciler reconciles a RemoteConfig object
type RemoteConfigReconciler struct {
	runaiReconciler
}

//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigs/finalizers,verbs=update

// NewRemoteConfigReconciler creates a new RemoteConfigReconciler
func NewRemoteConfigReconciler(k8s kube_client.HybridClient, verbosity int) *RemoteConfigReconciler {
	reconciler := &RemoteConfigReconciler{
		runaiReconciler: runaiReconciler{
			localClusterClient: k8s,
			scheme:             k8s.Scheme(),
			verbosity:          verbosity,
		},
	}

	return reconciler
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RemoteConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *RemoteConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	logger := log.FromContext(ctx).V(r.verbosity)

	multiErr := &me.MultiError{}
	retErr = multiErr

	res = requeue()
	remoteConfigResource, err := cu.GetRemoteConfigForRequest(ctx, r.localClusterClient, req)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("[%s/%s/%s] not found, probably deleted",
				req.Namespace, utils.ResourceStructType(remoteConfigResource), req.Name))
			return noRequeue(), nil
		}

		return res, err
	}
	defer func() {
		res, retErr = r.controllerDeferFunction(ctx, remoteConfigResource, multiErr, res)
		return
	}()
	runaiConfig, err := cu.ExtractRunaiConfigForRead(ctx, r.localClusterClient)
	if err != nil {
		multiErr.Append(err)
		return
	}
	err = r.applyRemoteConfig(ctx, runaiConfig, remoteConfigResource)
	if err != nil {
		multiErr.Append(err)
		return
	}
	res = ctrl.Result{}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *RemoteConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runaiv1.RemoteConfig{}).
		Watches(&runaiv1.RunaiConfig{}, handler.EnqueueRequestsFromMapFunc(r.generateRunaiConfigWatchReconcileRequest)).
		Complete(r)
}

func (r *RemoteConfigReconciler) applyRemoteConfig(ctx context.Context, runaiConfig *runaiv1.RunaiConfig, remoteConfigResource *runaiv1.RemoteConfig) error {
	if err := r.runaiRemoteConfigPatchRequest(ctx, runaiConfig, remoteConfigResource); err != nil {
		err = fmt.Errorf("error while updating runaiconfig with new remote config: %v", err)
		cu.SetCondition(remoteConfigResource, runaiv1.ConditionTypeRemoteConfigApplying,
			metav1.ConditionFalse,
			runaiv1.ConfigActionFailed,
			err.Error(), remoteConfigResource.Generation)
		cu.SetCondition(remoteConfigResource, runaiv1.ConditionTypeRemoteConfigApplied,
			metav1.ConditionFalse,
			runaiv1.ConfigActionFailed,
			"RemoteConfig failed to propagate to cluster", remoteConfigResource.Generation)
		return err
	}
	cu.SetCondition(remoteConfigResource, runaiv1.ConditionTypeRemoteConfigApplying,
		metav1.ConditionTrue,
		runaiv1.ConfigActionSuccess,
		"RemoteConfig applied to runaiconfig", remoteConfigResource.Generation)
	r.handleRemoteConfigAvailabilityStatus(runaiConfig, remoteConfigResource)
	return nil
}

func (r *RemoteConfigReconciler) handleRemoteConfigAvailabilityStatus(runaiConfig *runaiv1.RunaiConfig, remoteConfigResource *runaiv1.RemoteConfig) {
	if isRemoteConfigAppliedInCluster(runaiConfig) {
		cu.SetCondition(remoteConfigResource, runaiv1.ConditionTypeRemoteConfigApplied,
			metav1.ConditionTrue,
			runaiv1.ConfigActionSuccess,
			"RemoteConfig is applied to cluster", remoteConfigResource.Generation)
	} else {
		message := runaiConfig.ExtractReconciledConditionMessage()
		cu.SetCondition(remoteConfigResource, runaiv1.ConditionTypeRemoteConfigApplied,
			metav1.ConditionFalse,
			runaiv1.ConfigActionFailed,
			message, remoteConfigResource.Generation)
	}
	//We are updating the actual config in the status to reflect the actual config in the runaiconfig
	if runaiConfig.Status.RemoteConfig != nil {
		remoteConfigResource.Status.ActualConfig.UserConfig = runaiConfig.Status.RemoteConfig.UserConfig
		remoteConfigResource.Status.ActualConfig.SystemConfig = runaiConfig.Status.RemoteConfig.SystemConfig
	}
}

func (r *RemoteConfigReconciler) generateRunaiConfigWatchReconcileRequest(ctx context.Context, object client.Object) []reconcile.Request {
	log := log.FromContext(ctx)
	_, ok := object.(*runaiv1.RunaiConfig)
	if !ok {
		log.Error(fmt.Errorf("expected a Runaiconfig but got a %T", object), "")
		return []reconcile.Request{}
	}
	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Namespace: runaiv1.DefaultNamespace,
				Name:      runaiv1.DefaultName,
			},
		},
	}
}

func isRemoteConfigAppliedInCluster(runaiConfig *runaiv1.RunaiConfig) bool {
	if runaiConfig.Status.RemoteConfig == nil {
		return false
	}
	userConfigEqual := reflect.DeepEqual(runaiConfig.Spec.Internal.RemoteConfig.UserConfig, runaiConfig.Status.RemoteConfig.UserConfig)
	systemConfigEqual := reflect.DeepEqual(runaiConfig.Spec.Internal.RemoteConfig.SystemConfig, runaiConfig.Status.RemoteConfig.SystemConfig)
	if !userConfigEqual || !systemConfigEqual {
		return false
	}
	return runaiConfig.IsRunaiConfigAvailableAndReconciled()
}

func (r *RemoteConfigReconciler) runaiRemoteConfigPatchRequest(ctx context.Context, runaiConfig *runaiv1.RunaiConfig, remoteConfigResource *runaiv1.RemoteConfig) error {
	remoteConfigToPatch := runaiv1.RunaiConfigSpec{
		Internal: &runaiv1internal.Internal{
			RemoteConfig: &runaiv1internal.ClusterFullConfig{
				ClusterFullConfigData: openapi.ClusterFullConfigData{
					UserConfig:   remoteConfigResource.Spec.UserConfig,
					SystemConfig: remoteConfigResource.Spec.SystemConfig,
				},
			},
		},
	}
	patchBytes, err := json.Marshal(remoteConfigToPatch)
	if err != nil {
		return fmt.Errorf("error while marshalling remote config to patch: %v", err)
	}
	unmarshalObj := map[string]interface{}{}
	err = json.Unmarshal(patchBytes, &unmarshalObj)
	patchBytes, err = json.Marshal(map[string]interface{}{"spec": unmarshalObj})
	if err != nil {
		return fmt.Errorf("error while marshalling remote config to patch: %v", err)
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	return r.localClusterClient.Patch(ctx, runaiConfig, patch)
}
