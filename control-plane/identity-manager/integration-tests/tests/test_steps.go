package tests

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	serviceUtils "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
)

var realmCreationMu sync.Mutex

const (
	realmNamePrefix          = "realm"
	serviceAccountNamePrefix = "service-account"
	accessKeyNamePrefix      = "access-key"
	createdByUser            = "test@run.ai"
)

type TestSteps struct {
	iamClient     *iam.IamServiceContainer
	openapiClient *openapi.APIClient
	hostHeader    string
}

func NewTestSteps(iamClient *iam.IamServiceContainer, openapiClient *openapi.APIClient, hostHeader string) *TestSteps {
	return &TestSteps{
		iamClient:     iamClient,
		openapiClient: openapiClient,
		hostHeader:    hostHeader,
	}
}

func (t *TestSteps) CreateRealm(ctx context.Context, name *string, tenantId *string) (string, string, error) {
	if name == nil {
		randname := GenerateRandomString(realmNamePrefix)
		name = &randname
	}
	if tenantId == nil {
		tenantIdStr := GenerateRandomTenantID()
		tenantId = &tenantIdStr
	}
	realmReq := openapi.RealmCreationRequest{
		RealmName: *name,
		TenantId:  *tenantId,
	}
	_, resp, err := t.openapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmReq).Execute()
	if err != nil && resp != nil && resp.StatusCode != http.StatusConflict {
		return "", "", err
	}
	return *name, *tenantId, nil
}

func (t *TestSteps) DeleteRealm(ctx context.Context, realmName string, isHardDelete bool) error {
	_, err := t.openapiClient.RealmAPI.DeleteRealm(ctx, realmName).IsHardDelete(isHardDelete).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (t *TestSteps) DeleteIDP(ctx context.Context, idpAlias string) error {
	_, err := t.openapiClient.IdpsAPI.DeleteIdp(ctx, idpAlias).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (t *TestSteps) CreateIDP(ctx context.Context, idpRequest openapi.IdpCreationRequest) error {
	_, _, err := t.openapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (t *TestSteps) CreateServiceAccount(ctx context.Context, realm string, tenantId string, appName *string) (string, error) {
	if appName == nil {
		randname := GenerateRandomString(serviceAccountNamePrefix)
		appName = &randname
	}
	serviceAccount := &models.Client{
		Name:     appName,
		ClientID: appName,
		Enabled:  ptr.To(true),
		TenantId: &tenantId,
		Attributes: &models.ClientAttributes{
			CreatedBy: ptr.To(createdByUser),
		},
	}
	realm_client.ApplyClientProfile(serviceAccount, realm_client.ClientTypeTenantApp)
	return t.iamClient.ClientService.CreateClient(ctx, realm, serviceAccount)
}

func (t *TestSteps) CreateAccessKey(ctx context.Context, realm, tenantId string, appName, createdBy *string) (string, error) {
	if appName == nil {
		randname := GenerateRandomString(accessKeyNamePrefix)
		appName = &randname
	}
	if createdBy == nil {
		createdBy = ptr.To(createdByUser)
	}
	accessKey := &models.Client{
		Name:     appName,
		ClientID: ptr.To("clientID" + uuid.New().String()),
		Enabled:  ptr.To(true),
		TenantId: &tenantId,
		Attributes: &models.ClientAttributes{
			CreatedBy: createdBy,
		},
	}
	realm_client.ApplyClientProfile(accessKey, realm_client.ClientTypeUserApp)
	return t.iamClient.ClientService.CreateClient(ctx, realm, accessKey)
}

func (t *TestSteps) DeleteClient(ctx context.Context, realm string, appId string) error {
	return t.iamClient.ClientService.DeleteClient(ctx, realm, appId)
}

func (t *TestSteps) CreateUser(ctx context.Context, realm string, username *string) (string, string, string, error) {
	if username == nil {
		username = ptr.To(GenerateUsername())
	}
	password := serviceUtils.GenerateTemporaryUserPassword()
	userID, err := t.iamClient.UserService.CreateUser(ctx, realm, &models.User{
		Username:      username,
		Email:         username,
		Enabled:       ptr.To(true),
		ResetPassword: ptr.To(false),
		CreatedBy:     ptr.To(createdByUser),
		TempPassword:  &password,
	})
	return userID, *username, password, err
}

func (t *TestSteps) NewTestContextWithRealm(prefix string) (context.Context, string, string) {
	realm := RandomRealmName(prefix)
	tenantID := GenerateRandomTenantID()
	realmPtr := &realm
	tenantIDPtr := &tenantID

	realmCreationMu.Lock()
	defer realmCreationMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _, err := t.CreateRealm(ctx, realmPtr, tenantIDPtr)
	if err != nil {
		panic(err)
	}
	ctx = t.SetContextValues(context.Background(), tenantID)
	return ctx, realm, tenantID
}

func (t *TestSteps) SetContextValues(ctx context.Context, tenantID string) context.Context {
	return t.SetContextValuesWithUser(ctx, tenantID, "test@run.ai")
}

func (t *TestSteps) SetContextValuesWithUser(ctx context.Context, tenantID string, username string) context.Context {
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, username)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)
	return ctx
}

func (t *TestSteps) CreateInternalClient(ctx context.Context, tenantId string, clientName *string) string {
	if clientName == nil {
		clientName = ptr.To(GenerateRandomString("client"))
	}

	clientRequest := openapi.ClientCreationRequest{
		Type:     openapi.CLIENTTYPE_AGENT,
		ClientId: *clientName,
		TenantId: ptr.To(tenantId),
	}

	resp, _, err := t.openapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()
	if err != nil {
		panic(err)
	}
	return *resp.Id
}

func (t *TestSteps) DeleteInternalClient(ctx context.Context, clientId string) error {
	_, err := t.openapiClient.ClientsAPI.DeleteClient(ctx, clientId).Execute()
	return err
}

func (t *TestSteps) ReconcileRealm(ctx context.Context, realmName string, reconcilerType *openapi.ReconcilerType) error {
	_, err := t.openapiClient.RealmAPI.ReconcileRealm(ctx, realmName).RealmReconcileRequest(openapi.RealmReconcileRequest{
		Force:          ptr.To(false),
		ReconcilerType: *openapi.NewNullableReconcilerType(reconcilerType),
	}).Execute()
	if err != nil {
		return err
	}
	return nil
}

type fakeHostHeaderCtxKey struct{}

func (t *TestSteps) SetFakeHostHeader(ctx context.Context, openapiClient *openapi.APIClient, realmName string) context.Context {
	fakeHost := fmt.Sprintf("%s.fake.run.ai", realmName)
	return context.WithValue(ctx, fakeHostHeaderCtxKey{}, fakeHost)
}

type FakeHostHeaderTransport struct {
	HeaderName string
	Base       http.RoundTripper
}

func (rt *FakeHostHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if value, ok := req.Context().Value(fakeHostHeaderCtxKey{}).(string); ok {
		req.Header.Set(rt.HeaderName, value)
	}
	return rt.Base.RoundTrip(req)
}
