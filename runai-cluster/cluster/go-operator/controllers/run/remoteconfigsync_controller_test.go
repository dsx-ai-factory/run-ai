/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/clients"
	clientsmocks "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/clients/mocks"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/secrets"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type RemoteConfigSyncTestTypes struct {
	reconciler           *run.RemoteConfigSyncReconciler
	clusterServiceClient *clientsmocks.ClusterServiceClient
}

func TestReconcileRemoteConfigSyncHappyBasicFlow(t *testing.T) {
	kbClient := createEmptyKubeClientRemoteConfigSyncMock()
	testTypes := initReconcilerAndMocks(t, kbClient)
	testTypes.clusterServiceClient.On("GetClusterInternalConfig", context.Background()).Return(&cpclusterclient.GetInternalClusterConfigResponse{
		Config: cpclusterclient.ClusterFullConfigData{
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: false,
			},
		},
	}, nil)

	testTypes.clusterServiceClient.On("PutRemoteConfigStatus", mock.Anything, mock.Anything).Return(nil)

	cpDetails := &clients.ControlPlaneDetails{
		URL:          "URL",
		ClientId:     "name",
		ClientSecret: "secret",
		ClusterUUID:  "UUID",
	}
	testTypes.clusterServiceClient.On("SetCpDetails", cpDetails).Return(testTypes.clusterServiceClient, nil)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigSyncName,
		},
	}
	res, err := testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, res.Requeue, true)
	assert.Equal(t, res.RequeueAfter, time.Duration(runaiv1.DefaultSyncIntervalMs)*time.Millisecond)
	remote := &runaiv1.RemoteConfig{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Nil(t, err)
	assert.Equal(t, remote.Spec.SystemConfig.Deactivated, false)

	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)

	conditions := remoteConfigSync.Status.Conditions
	assert.Len(t, conditions, 3)

	reconciled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)

	pulled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPulled)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pulled.Status)

	pushed := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPushed)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pushed.Status)
}

func TestReconcileRemoteConfigSyncRecoversAfterConfigurationSecretIsCreated(t *testing.T) {
	kbClient := createEmptyKubeClientRemoteConfigSyncMock()
	testTypes := initReconcilerAndMocks(t, kbClient)
	configurationSecret := getInstallationConfigurationSecret()
	err := kbClient.Delete(context.Background(), configurationSecret)
	assert.NoError(t, err)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigSyncName,
		},
	}

	res, err := testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(0), res.RequeueAfter)

	err = kbClient.Create(context.Background(), configurationSecret)
	assert.NoError(t, err)
	cpDetails := &clients.ControlPlaneDetails{
		URL:          "URL",
		ClientId:     "name",
		ClientSecret: "secret",
		ClusterUUID:  "UUID",
	}
	testTypes.clusterServiceClient.On("SetCpDetails", cpDetails).Return(testTypes.clusterServiceClient, nil).Once()
	testTypes.clusterServiceClient.On("GetClusterInternalConfig", context.Background()).Return(&cpclusterclient.GetInternalClusterConfigResponse{
		Config: cpclusterclient.ClusterFullConfigData{
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: false,
			},
		},
	}, nil).Once()
	testTypes.clusterServiceClient.On("PutRemoteConfigStatus", mock.Anything, mock.Anything).Return(nil).Once()

	res, err = testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(runaiv1.DefaultSyncIntervalMs)*time.Millisecond, res.RequeueAfter)

	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)
	reconciled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)
}

func TestReconcileRemoteConfigSyncFailedToPullFromCP(t *testing.T) {
	kbClient := createEmptyKubeClientRemoteConfigSyncMock()
	testTypes := initReconcilerAndMocks(t, kbClient)
	testTypes.clusterServiceClient.On("GetClusterInternalConfig", context.Background()).Return(nil, errors.New("failure"))

	testTypes.clusterServiceClient.On("PutRemoteConfigStatus", mock.Anything, mock.Anything).Return(nil)

	cpDetails := &clients.ControlPlaneDetails{
		URL:          "URL",
		ClientId:     "name",
		ClientSecret: "secret",
		ClusterUUID:  "UUID",
	}
	testTypes.clusterServiceClient.On("SetCpDetails", cpDetails).Return(testTypes.clusterServiceClient, nil)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigSyncName,
		},
	}
	res, err := testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, res.Requeue, true)
	assert.Equal(t, res.RequeueAfter, time.Duration(0))

	remote := &runaiv1.RemoteConfig{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Nil(t, err)
	assert.True(t, remote.Spec.SystemConfig == nil)

	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)

	conditions := remoteConfigSync.Status.Conditions
	assert.Len(t, conditions, 3)

	reconciled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionFalse, reconciled.Status)

	pulled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPulled)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionFalse, pulled.Status)

	pushed := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPushed)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pushed.Status)
}

func TestReconcileRemoteConfigSyncFailedToPushToCP(t *testing.T) {
	kbClient := createEmptyKubeClientRemoteConfigSyncMock()
	testTypes := initReconcilerAndMocks(t, kbClient)
	testTypes.clusterServiceClient.On("GetClusterInternalConfig", context.Background()).Return(&cpclusterclient.GetInternalClusterConfigResponse{
		Config: cpclusterclient.ClusterFullConfigData{
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}, nil)
	testTypes.clusterServiceClient.On("PutRemoteConfigStatus", mock.Anything, mock.Anything).Return(errors.New("failure"))

	cpDetails := &clients.ControlPlaneDetails{
		URL:          "URL",
		ClientId:     "name",
		ClientSecret: "secret",
		ClusterUUID:  "UUID",
	}
	testTypes.clusterServiceClient.On("SetCpDetails", cpDetails).Return(testTypes.clusterServiceClient, nil)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigSyncName,
		},
	}
	res, err := testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, res.Requeue, true)
	assert.Equal(t, res.RequeueAfter, time.Duration(0))

	remote := &runaiv1.RemoteConfig{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Nil(t, err)
	assert.Equal(t, remote.Spec.SystemConfig.Deactivated, true)

	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)

	conditions := remoteConfigSync.Status.Conditions
	assert.Len(t, conditions, 3)

	reconciled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionFalse, reconciled.Status)

	pulled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPulled)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pulled.Status)

	pushed := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPushed)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionFalse, pushed.Status)
}

// Test that the remote config sync is being terminated when the system config is deactivated
func TestReconcileRemoteConfigSyncTermination(t *testing.T) {
	kbClient := createEmptyKubeClientRemoteConfigSyncMock()
	testTypes := initReconcilerAndMocks(t, kbClient)
	testTypes.clusterServiceClient.On("GetClusterInternalConfig", context.Background()).Return(&cpclusterclient.GetInternalClusterConfigResponse{
		Config: cpclusterclient.ClusterFullConfigData{
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}, nil)

	testTypes.clusterServiceClient.On("PutRemoteConfigStatus", mock.Anything, mock.Anything).Return(nil)

	cpDetails := &clients.ControlPlaneDetails{
		URL:          "URL",
		ClientId:     "name",
		ClientSecret: "secret",
		ClusterUUID:  "UUID",
	}
	testTypes.clusterServiceClient.On("SetCpDetails", cpDetails).Return(testTypes.clusterServiceClient, nil)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigSyncName,
		},
	}

	res, err := testTypes.reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, res.Requeue, true)
	assert.Equal(t, res.RequeueAfter, time.Duration(runaiv1.DefaultSyncIntervalMs)*time.Millisecond)
	remote := &runaiv1.RemoteConfig{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)

	assert.Nil(t, err)
	assert.Equal(t, remote.Spec.SystemConfig.Deactivated, true)

	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)

	conditions := remoteConfigSync.Status.Conditions
	assert.Len(t, conditions, 3)

	reconciled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)

	pulled := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPulled)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pulled.Status)

	pushed := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigPushed)
	assert.NotNil(t, pulled)
	assert.Equal(t, metav1.ConditionTrue, pushed.Status)

	remote.Status.ActualConfig = runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	err = kbClient.Status().Update(context.Background(), remote)
	assert.Nil(t, err)

	res, err = testTypes.reconciler.Reconcile(context.Background(), req)

	assert.Nil(t, err)
	remoteConfigSync = &runaiv1.RemoteConfigSync{}
	err = kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Nil(t, err)

	terminated := remoteConfigSync.Status.GetCondition(runaiv1.ConditionTypeConfigTerminated)
	assert.NotNil(t, terminated)
	assert.Equal(t, metav1.ConditionTrue, terminated.Status)

	res, err = testTypes.reconciler.Reconcile(context.Background(), req)

	assert.Nil(t, err)
	assert.Equal(t, res.Requeue, false)
}

func initReconcilerAndMocks(t *testing.T, k8sClient kc.HybridClient) *RemoteConfigSyncTestTypes {
	clusterServiceClient := clientsmocks.NewClusterServiceClient(t)
	reconciler := run.NewRemoteConfigSyncReconciler(k8sClient, 1, clusterServiceClient)
	return &RemoteConfigSyncTestTypes{
		reconciler:           reconciler,
		clusterServiceClient: clusterServiceClient,
	}
}

func createEmptyKubeClientRemoteConfigSyncMock() kc.HybridClient {
	remoteConfig := getDefaultRemoteConfig()
	remoteConfigSync := getDefaultRemoteConfigSync()
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.ObjectMeta = metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Internal.Global.ControlPlane = &runaiv1internal.ControlPlane{
		Config:      &runaiv1internal.ClusterControlPlaneConfig{},
		URL:         "URL",
		ClusterUUID: "runai-config-uuid",
	}
	configurationSecret := getInstallationConfigurationSecret()
	fc := fakeClient.NewClientBuilder().WithScheme(scheme_init.Scheme()).
		WithRuntimeObjects(remoteConfig, runaiConfig, remoteConfigSync, configurationSecret).
		WithStatusSubresource(remoteConfig, runaiConfig, remoteConfigSync).
		Build()
	return kc.New(fc, fc, nil)

}

func getInstallationConfigurationSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secrets.InstallationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
		Data: map[string][]byte{
			secrets.InstallationConfigurationSecretName: []byte("global:\n  controlPlane:\n    clusterUUID: UUID\n    config:\n      agent:\n        client:\n          name: name\n          secret: secret\n"),
		},
	}
}

func getDefaultRemoteConfigSync() *runaiv1.RemoteConfigSync {
	return &runaiv1.RemoteConfigSync{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.RemoteConfigSyncName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       runaiv1.RemoteConfigSyncKind,
			APIVersion: runaiv1.GroupVersion.String(),
		},
	}
}
