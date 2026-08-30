package run_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	cu "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestConditionUpdate(t *testing.T) {
	kbClient := createEmptyKubeClientMock()
	runaiConfig, err := cu.ExtractRunaiConfigForRead(context.Background(), kbClient)
	assert.Nil(t, err)
	assert.False(t, runaiConfig.Spec.Internal.RemoteConfig.SystemConfig.Deactivated)
	remote := &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Nil(t, remote.Spec.SystemConfig)
	remote.Spec.ClusterFullConfig = runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	kbClient.Update(context.Background(), remote)
	cu.SetCondition(remote, runaiv1.ConditionTypeRemoteConfigApplying,
		metav1.ConditionFalse,
		runaiv1.ConfigActionFailed,
		"RemoteConfig failed to propagate to cluster", remote.GetGeneration())
	cu.SetCondition(remote, runaiv1.ConditionTypeReconciled,
		metav1.ConditionFalse,
		runaiv1.NotReconciled,
		"err", remote.GetGeneration())

	kbClient.Status().Update(context.Background(), remote)
	reconciler := run.NewRemoteConfigReconciler(kbClient, 0)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigName,
		},
	}
	_, err = reconciler.Reconcile(context.Background(), req)
	remote = &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Equal(t, 3, len(remote.Status.Conditions))
	reconciled := remote.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)

	applied := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplying)
	assert.NotNil(t, applied)
	assert.Equal(t, metav1.ConditionTrue, applied.Status)

	available := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplied)
	assert.NotNil(t, available)
	assert.Equal(t, metav1.ConditionFalse, available.Status)
	assert.Nil(t, remote.Status.ActualConfig.SystemConfig)
	assert.Nil(t, remote.Status.ActualConfig.UserConfig)
}

func TestReconcileRunaiConfigPatchWorks(t *testing.T) {
	kbClient := createEmptyKubeClientMock()
	runaiConfig, err := cu.ExtractRunaiConfigForRead(context.Background(), kbClient)
	assert.Nil(t, err)
	assert.False(t, runaiConfig.Spec.Internal.RemoteConfig.SystemConfig.Deactivated)
	remote := &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Nil(t, remote.Spec.SystemConfig)
	remote.Spec.ClusterFullConfig = runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	kbClient.Update(context.Background(), remote)
	reconciler := run.NewRemoteConfigReconciler(kbClient, 0)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigName,
		},
	}
	_, err = reconciler.Reconcile(context.Background(), req)
	runaiConfig, err = cu.ExtractRunaiConfigForRead(context.Background(), kbClient)
	assert.Nil(t, err)
	assert.True(t, runaiConfig.Spec.Internal.RemoteConfig.SystemConfig.Deactivated)
}
func TestReconcileRemoteConfigApplyToCluster(t *testing.T) {
	kbClient := createEmptyKubeClientMock()
	remote := &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	remote.Spec.ClusterFullConfig = runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	kbClient.Update(context.Background(), remote)
	reconciler := run.NewRemoteConfigReconciler(kbClient, 0)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigName,
		},
	}
	_, err := reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	remote = &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Equal(t, 3, len(remote.Status.Conditions))
	reconciled := remote.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)

	applied := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplying)
	assert.NotNil(t, applied)
	assert.Equal(t, metav1.ConditionTrue, applied.Status)

	available := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplied)
	assert.NotNil(t, available)
	assert.Equal(t, metav1.ConditionFalse, available.Status)
	assert.Nil(t, remote.Status.ActualConfig.SystemConfig)
	assert.Nil(t, remote.Status.ActualConfig.UserConfig)
}

func TestReconcileRemoteConfigAvailableToCluster(t *testing.T) {
	//prepare remote config & runai config objects
	fullConfig := &runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	kbClient := createEmptyKubeClientMock()
	remote := &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	remote.Spec.ClusterFullConfig = *fullConfig
	kbClient.Update(context.Background(), remote)

	runaiConfig := &runaiv1.RunaiConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}, runaiConfig)
	runaiConfig.Spec.Internal.RemoteConfig = fullConfig
	err := kbClient.Update(context.Background(), runaiConfig)
	runaiConfig.Status.RemoteConfig = &runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	cu.SetCondition(runaiConfig, runaiv1.ConditionTypeReconciled, metav1.ConditionTrue, runaiv1.Reconciled, "Reconciliation completed successfully", runaiConfig.GetGeneration())
	cu.SetCondition(runaiConfig, runaiv1.ConditionTypeAvailable, metav1.ConditionTrue, runaiv1.Available, "System Available", runaiConfig.Generation)
	err = kbClient.Status().Update(context.Background(), runaiConfig)

	//Test
	reconciler := run.NewRemoteConfigReconciler(kbClient, 0)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigName,
		},
	}
	_, err = reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	remote = &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Equal(t, 3, len(remote.Status.Conditions))
	reconciled := remote.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionTrue, reconciled.Status)

	applied := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplying)
	assert.NotNil(t, applied)
	assert.Equal(t, metav1.ConditionTrue, applied.Status)

	available := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplied)
	assert.NotNil(t, available)
	assert.Equal(t, metav1.ConditionTrue, available.Status)

	userConfigEqual := reflect.DeepEqual(remote.Spec.UserConfig, remote.Status.ActualConfig.UserConfig)
	systemConfigEqual := reflect.DeepEqual(remote.Spec.SystemConfig, remote.Status.ActualConfig.SystemConfig)
	assert.True(t, userConfigEqual && systemConfigEqual)
}

func TestReconcileRemoteConfigErrorOnPatch(t *testing.T) {
	//prepare remote config & runai config objects
	fullConfig := &runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	kbClient := createKubeClientMockErrorOnPatch()
	remote := &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	remote.Spec.ClusterFullConfig = *fullConfig
	kbClient.Update(context.Background(), remote)

	runaiConfig := &runaiv1.RunaiConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}, runaiConfig)
	runaiConfig.Spec.Internal.RemoteConfig = fullConfig
	err := kbClient.Update(context.Background(), runaiConfig)
	runaiConfig.Status.RemoteConfig = &runaiv1internal.ClusterFullConfig{
		ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
			UserConfig: &cpclusterclient.ClusterUserConfigData{},
			SystemConfig: &cpclusterclient.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
	}
	cu.SetCondition(runaiConfig, runaiv1.ConditionTypeReconciled, metav1.ConditionTrue, runaiv1.Reconciled, "Reconciliation completed successfully", runaiConfig.GetGeneration())
	cu.SetCondition(runaiConfig, runaiv1.ConditionTypeAvailable, metav1.ConditionTrue, runaiv1.Available, "System Available", runaiConfig.Generation)
	err = kbClient.Status().Update(context.Background(), runaiConfig)

	//Test
	reconciler := run.NewRemoteConfigReconciler(kbClient, 0)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: runaiv1.RemoteConfigName,
		},
	}
	_, err = reconciler.Reconcile(context.Background(), req)
	assert.Nil(t, err)
	remote = &runaiv1.RemoteConfig{}
	kbClient.Get(context.Background(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remote)
	assert.Equal(t, 3, len(remote.Status.Conditions))
	reconciled := remote.Status.GetCondition(runaiv1.ConditionTypeReconciled)
	assert.NotNil(t, reconciled)
	assert.Equal(t, metav1.ConditionFalse, reconciled.Status)

	applied := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplying)
	assert.NotNil(t, applied)
	assert.Equal(t, metav1.ConditionFalse, applied.Status)

	available := remote.Status.GetCondition(runaiv1.ConditionTypeRemoteConfigApplied)
	assert.NotNil(t, available)
	assert.Equal(t, metav1.ConditionFalse, available.Status)

	assert.Nil(t, remote.Status.ActualConfig.SystemConfig)
	assert.Nil(t, remote.Status.ActualConfig.UserConfig)
}

func createEmptyKubeClientMock() kc.HybridClient {
	remoteConfig := getDefaultRemoteConfig()
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.ObjectMeta = metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}
	runaiConfig.SetDefaultsWhereNeeded()
	fc := fakeClient.NewClientBuilder().WithScheme(scheme_init.Scheme()).
		WithRuntimeObjects(remoteConfig, runaiConfig).
		WithStatusSubresource(remoteConfig, runaiConfig).
		Build()
	return kc.New(fc, fc, nil)

}

func createKubeClientMockErrorOnPatch() kc.HybridClient {
	remoteConfig := getDefaultRemoteConfig()
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.ObjectMeta = metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}
	runaiConfig.SetDefaultsWhereNeeded()
	fc := fakeClient.NewClientBuilder().WithScheme(scheme_init.Scheme()).
		WithRuntimeObjects(remoteConfig, runaiConfig).
		WithStatusSubresource(remoteConfig, runaiConfig).
		WithInterceptorFuncs(interceptor.Funcs{
			Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
				return errors.New("errorPatching runaiconfig")
			},
		}).
		Build()
	return kc.New(fc, fc, nil)
}

func getDefaultRemoteConfig() *runaiv1.RemoteConfig {
	return &runaiv1.RemoteConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.RemoteConfigName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       runaiv1.RemoteConfigKind,
			APIVersion: runaiv1.GroupVersion.String(),
		},
	}
}
