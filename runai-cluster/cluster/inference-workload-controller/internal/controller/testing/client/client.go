package client

import (
	safeFake "github.com/run-ai/runai/runai-cluster/cluster/common/utils/client"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func InitFakeClientWithObjectNoStatus(scheme *runtime.Scheme, objects ...client.Object) client.Client {
	return InitFakeClientWithObjectWithStatus(scheme, nil, objects...)
}

func InitFakeClientWithObjectWithStatus(scheme *runtime.Scheme, statusObjects []client.Object, objects ...client.Object) client.Client {
	testClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithStatusSubresource(statusObjects...).Build()
	return safeFake.NewFakeClientConcurrentSafe(testClient)
}
