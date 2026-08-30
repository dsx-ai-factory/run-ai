package step

import (
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockK8sClients implements the K8sClients interface for testing
type mockK8sClients struct {
	client client.Client
	config *rest.Config
}

func (m *mockK8sClients) GetClient() client.Client {
	return m.client
}

func (m *mockK8sClients) GetConfig() *rest.Config {
	return m.config
}

func SetupMockK8sClients(objects ...client.Object) k8s.K8sClients {
	fakeClient := fake.NewClientBuilder().WithObjects(objects...).Build()
	mock := &mocks.MockK8sClients{}
	mock.On("GetClient").Return(fakeClient).Maybe()
	mock.On("GetConfig").Return(&rest.Config{}).Maybe()
	return mock
}
