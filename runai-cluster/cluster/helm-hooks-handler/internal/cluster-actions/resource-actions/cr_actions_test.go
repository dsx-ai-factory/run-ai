package resource_actions

import (
	"context"
	"testing"

	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateFlowCrs(t *testing.T) {
	k8s := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
	remoteConfig := &runaiv1.RemoteConfig{}
	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	cluster := &runaiv1.Cluster{}
	err := k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.ClusterName,
	}, cluster)
	assert.NotEqual(t, err, nil)
	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remoteConfig)
	assert.NotEqual(t, err, nil)
	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.NotEqual(t, err, nil)

	// Create Cluster CR
	err = createClusterCR(k8s)
	assert.Equal(t, err, nil)

	cluster, err = HandleClusterCR(context.Background(), k8s, false)
	assert.Equal(t, err, nil)
	assert.NotNil(t, cluster)
	checkThatAllCrsAreInPlace(t, k8s)
}

func TestUpdateFlowCrs(t *testing.T) {
	k8s := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()

	// Create Cluster CR
	err := createClusterCR(k8s)
	assert.Equal(t, err, nil)

	_, err = HandleClusterCR(context.Background(), k8s, false)
	assert.Equal(t, err, nil)

	remoteConfig := &runaiv1.RemoteConfig{}
	remoteConfigSync := &runaiv1.RemoteConfigSync{}

	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remoteConfig)
	assert.Equal(t, err, nil)

	err = deleteCR(k8s, remoteConfig)
	assert.Equal(t, err, nil)

	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remoteConfig)
	assert.NotEqual(t, err, nil)

	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Equal(t, err, nil)

	_, err = HandleClusterCR(context.Background(), k8s, false)
	assert.Equal(t, err, nil)
	checkThatAllCrsAreInPlace(t, k8s)
}

func checkThatAllCrsAreInPlace(t *testing.T, k8s client.Client) {
	remoteConfig := &runaiv1.RemoteConfig{}
	remoteConfigSync := &runaiv1.RemoteConfigSync{}
	cluster := &runaiv1.Cluster{}

	err := k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.ClusterName,
	}, cluster)
	assert.Equal(t, err, nil)

	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigName,
	}, remoteConfig)
	assert.Equal(t, err, nil)
	assert.Equal(t, remoteConfig.ObjectMeta.OwnerReferences[0].Name, runaiv1.ClusterName)
	assert.Equal(t, remoteConfig.ObjectMeta.OwnerReferences[0].UID, cluster.UID)

	err = k8s.Get(context.TODO(), types.NamespacedName{
		Name: runaiv1.RemoteConfigSyncName,
	}, remoteConfigSync)
	assert.Equal(t, err, nil)
	assert.Equal(t, remoteConfigSync.ObjectMeta.OwnerReferences[0].Name, runaiv1.ClusterName)
	assert.Equal(t, remoteConfigSync.ObjectMeta.OwnerReferences[0].UID, cluster.UID)
}

func deleteCR(k8s client.Client, cr runaiv1.RunaiCustomTypeI) error {
	for {
		deleteOptions := client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
		}
		err := k8s.Delete(context.Background(), cr, &deleteOptions)
		if k8sErrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			break
		} else if err != nil {
			return err
		}

	}
	return nil
}

func createClusterCR(k8s client.Client) error {
	return k8s.Create(context.Background(), &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.ClusterName,
		},
	})
}
