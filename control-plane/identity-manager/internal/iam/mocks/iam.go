package mocks

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/services/mocks"
)

type MockServicesContainer struct {
	MockUserService     mocks.UserService
	MockRealmService    mocks.RealmService
	MockAuthService     mocks.AuthService
	MockIdpService      mocks.IdpService
	MockClientService   mocks.ClientService
	MockAuthFlowService mocks.AuthFlowService
}

func NewMockServicesContainer() *MockServicesContainer {
	return &MockServicesContainer{
		MockUserService:     *new(mocks.UserService),
		MockRealmService:    *new(mocks.RealmService),
		MockAuthService:     *new(mocks.AuthService),
		MockIdpService:      *new(mocks.IdpService),
		MockClientService:   *new(mocks.ClientService),
		MockAuthFlowService: *new(mocks.AuthFlowService),
	}
}

func (m *MockServicesContainer) ToServicesContainer() *iam.IamServiceContainer {
	return &iam.IamServiceContainer{
		UserService:     &m.MockUserService,
		RealmService:    &m.MockRealmService,
		AuthService:     &m.MockAuthService,
		IdpService:      &m.MockIdpService,
		ClientService:   &m.MockClientService,
		AuthFlowService: &m.MockAuthFlowService,
	}
}
