/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"fmt"
	"reflect"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kube_client "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	me "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/multi-error"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RemoteConfigSyncReconciler reconciles a RemoteConfigSync object
type RemoteConfigSyncReconciler struct {
	runaiReconciler
	clusterServiceClient clients.ClusterServiceClient
}

func NewRemoteConfigSyncReconciler(k8s kube_client.HybridClient, verbosity int, clusterClient clients.ClusterServiceClient) *RemoteConfigSyncReconciler {
	reconciler := &RemoteConfigSyncReconciler{
		runaiReconciler: runaiReconciler{
			localClusterClient: k8s,
			scheme:             k8s.Scheme(),
			verbosity:          verbosity,
		},
		clusterServiceClient: clusterClient,
	}

	return reconciler
}

//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigsyncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigsyncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=remoteconfigsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RemoteConfigSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *RemoteConfigSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	logger := log.FromContext(ctx)
	multiErr := &me.MultiError{}

	remoteSyncResource, err := cu.GetRemoteConfigSyncForRequest(ctx, r.localClusterClient, req)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("[%s/%s/%s] not found, probably deleted",
				req.Namespace, utils.ResourceStructType(remoteSyncResource), req.Name))
			return noRequeue(), nil
		}
		return requeue(), err
	}

	if cu.IsRemoteConfigSyncTerminated(remoteSyncResource) {
		logger.Info(fmt.Sprintf("RemoteConfigSync %s is terminated", utils.ResourceUniqueID(remoteSyncResource)))
		return ctrl.Result{}, nil
	}

	interval := *remoteSyncResource.Spec.SyncIntervalInMs
	if interval == 0 || remoteSyncResource.Spec.SyncIntervalInMs == nil {
		logger.Info(fmt.Sprintf("RemoteConfigSync %s is disabled", utils.ResourceUniqueID(remoteSyncResource)))
		return ctrl.Result{}, nil
	}

	res = requeue()
	defer func() {
		res, retErr = r.controllerDeferFunction(ctx, remoteSyncResource, multiErr, res)
		return
	}()

	runaiConfig, err := cu.ExtractRunaiConfigForRead(ctx, r.localClusterClient)
	if err != nil {
		multiErr.Append(err)
		return res, multiErr
	}

	err = r.setClusterServiceClientWithCpDetails(ctx, runaiConfig)
	if err != nil {
		multiErr.Append(err)
		return res, multiErr
	}

	configResponse, err := r.pullRemoteConfigFromCP(remoteSyncResource)
	if err != nil {
		multiErr.Append(err)
	} else {
		err = r.updateRemoteConfig(ctx, configResponse)
		if err != nil {
			multiErr.Append(err)
		}
	}

	err = r.pushRemoteConfigToCP(remoteSyncResource)
	if err != nil {
		multiErr.Append(err)
	}
	if !multiErr.Empty() {
		return
	}
	res = requeueAfter(time.Duration(interval) * time.Millisecond)
	return
}

func (r *RemoteConfigSyncReconciler) pullRemoteConfigFromCP(remoteSyncResource *runaiv1.RemoteConfigSync) (*openapi.GetInternalClusterConfigResponse, error) {
	configResponse, err := r.clusterServiceClient.GetClusterInternalConfig(context.Background())
	if err != nil {
		cu.SetCondition(remoteSyncResource, runaiv1.ConditionTypeConfigPulled,
			metav1.ConditionFalse, runaiv1.PullFailed, err.Error(), remoteSyncResource.Generation)
		return nil, err
	}
	cu.SetCondition(remoteSyncResource, runaiv1.ConditionTypeConfigPulled,
		metav1.ConditionTrue, runaiv1.PullSuccessful, "The RemoteConfig has been successfully pulled from the ControlPlane.", remoteSyncResource.Generation)
	return configResponse, nil
}

func (r *RemoteConfigSyncReconciler) updateRemoteConfig(ctx context.Context, configResponse *openapi.GetInternalClusterConfigResponse) error {
	remoteConfig, err := cu.ExtractRemoteConfig(ctx, r.localClusterClient)
	if err != nil {
		return fmt.Errorf("error while getting remote config resource from the cluster: %v", err)
	}
	shouldPushRemoteConfig := r.remoteConfigRequirePush(&configResponse.Config, remoteConfig)
	if !shouldPushRemoteConfig {
		return nil
	}
	remoteConfig.Spec.UserConfig = configResponse.Config.UserConfig
	remoteConfig.Spec.SystemConfig = configResponse.Config.SystemConfig
	err = r.localClusterClient.Update(ctx, remoteConfig)
	if err != nil {
		return fmt.Errorf("error while updating remote config resource from the cluster: %v", err)
	}
	return nil
}

func (r *RemoteConfigSyncReconciler) setClusterServiceClientWithCpDetails(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) (err error) {
	configurationSecret, err := secrets.GetConfigurationSecret(ctx, r.localClusterClient)
	if err != nil {
		return err
	}
	configurationMapper, err := secrets.NewInstallationConfigurationMapper(configurationSecret)
	if err != nil {
		return err
	}
	cpDetails := &clients.ControlPlaneDetails{
		URL:          runaiConfig.Spec.Internal.Global.ControlPlane.URL,
		ClientSecret: configurationMapper.GetKeyBySP(secrets.ClientSecret),
		ClientId:     configurationMapper.GetKeyBySP(secrets.ClientName),
		ClusterUUID:  configurationMapper.GetKeyBySP(secrets.ClusterID),
	}

	configuredClient, err := r.clusterServiceClient.SetCpDetails(cpDetails)
	if err != nil {
		return
	}
	r.clusterServiceClient = configuredClient
	return nil
}

func (r *RemoteConfigSyncReconciler) remoteConfigRequirePush(cpInternalConfig *openapi.ClusterFullConfigData, remoteConfig *runaiv1.RemoteConfig) bool {
	userConfigEqual := reflect.DeepEqual(cpInternalConfig.UserConfig, remoteConfig.Status.ActualConfig.UserConfig)
	systemConfigEqual := reflect.DeepEqual(cpInternalConfig.SystemConfig, remoteConfig.Status.ActualConfig.SystemConfig)
	if userConfigEqual && systemConfigEqual {
		return false
	}
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *RemoteConfigSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runaiv1.RemoteConfigSync{}).
		Watches(&runaiv1.RemoteConfig{}, handler.EnqueueRequestsFromMapFunc(r.generateRemoteConfigWatchReconcileRequest)).
		Complete(r)
}

func (r *RemoteConfigSyncReconciler) pushRemoteConfigToCP(resource *runaiv1.RemoteConfigSync) error {
	ctx := context.Background()
	remoteConfig, err := cu.ExtractRemoteConfig(ctx, r.localClusterClient)
	if err != nil {
		return fmt.Errorf("error while getting remote config resource from the cluster: %v", err)
	}

	err = r.clusterServiceClient.PutRemoteConfigStatus(ctx, &openapi.PutClusterConfigStatusRequest{
		Status: *r.getPutRemoteConfigStatusRequest(remoteConfig),
	})
	if err != nil {
		cu.SetCondition(resource, runaiv1.ConditionTypeConfigPushed,
			metav1.ConditionFalse, runaiv1.PushFailed, err.Error(), resource.Generation)
		return err
	}

	cu.SetCondition(resource, runaiv1.ConditionTypeConfigPushed,
		metav1.ConditionTrue, runaiv1.PushSuccessful, "The RemoteConfig Status has been successfully pushed to the ControlPlane.", resource.Generation)

	if remoteConfig.Status.ActualConfig.GetSystemConfig().Deactivated {
		cu.SetCondition(resource, runaiv1.ConditionTypeConfigTerminated, metav1.ConditionTrue,
			runaiv1.TerminatedDueToDeactivation, "The sync has been terminated due to cluster deactivation.", resource.Generation)
	}

	return nil
}

func (r *RemoteConfigSyncReconciler) getPutRemoteConfigStatusRequest(remoteConfig *runaiv1.RemoteConfig) *openapi.PutClusterConfigStatusRequestStatus {
	requestConditions := []openapi.Condition{}
	for _, condition := range remoteConfig.Status.Conditions {
		requestConditions = append(requestConditions, openapi.Condition{
			Type:               condition.Type,
			Status:             string(condition.Status),
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime.Time,
		})
	}
	userConfig := remoteConfig.Status.ActualConfig.UserConfig
	if userConfig == nil {
		userConfig = &openapi.ClusterUserConfigData{}
	}
	systemConfig := remoteConfig.Status.ActualConfig.SystemConfig
	if systemConfig == nil {
		systemConfig = &openapi.ClusterSystemConfigData{}
	}
	return &openapi.PutClusterConfigStatusRequestStatus{
		Conditions: requestConditions,
		ActualConfig: &openapi.ClusterFullConfigData{
			UserConfig:   userConfig,
			SystemConfig: systemConfig,
		},
	}
}

func (r *RemoteConfigSyncReconciler) generateRemoteConfigWatchReconcileRequest(ctx context.Context, object client.Object) []reconcile.Request {
	log := log.FromContext(ctx)
	_, ok := object.(*runaiv1.RemoteConfig)
	if !ok {
		log.Error(fmt.Errorf("expected a RemoteConfig but got a %T", object), "")
		return []reconcile.Request{}
	}
	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Name: runaiv1.RemoteConfigName,
			},
		},
	}
}
