package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

type mockSettingsManager struct {
	mock.Mock
}

func (m *mockSettingsManager) GetAllSettings(tenantID string) ([]manager.Settinger, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]manager.Settinger), args.Error(1)
}

func (m *mockSettingsManager) GetSettingByKey(key api.WorkloadsSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	args := m.Called(key, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(manager.Settinger), args.Error(1)
}

func (m *mockSettingsManager) UpdateSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantID string, updateRequest api.UpdateWorkloadsSettingByKeyRequest) error {
	args := m.Called(ctx, key, tenantID, updateRequest)
	return args.Error(0)
}

func (m *mockSettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.WorkloadsSettingsGetResponse {
	args := m.Called(settingsList)
	return args.Get(0).(*api.WorkloadsSettingsGetResponse)
}

func (m *mockSettingsManager) TransformSettingByKeyToResponse(setting manager.Settinger) (interface{}, error) {
	args := m.Called(setting)
	return args.Get(0), args.Error(1)
}

const testTenantId = "tenant-123"

func ctxWithTenant(tenantId string) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
}

func managementCtxWithTenant(tenantId string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	return ctx
}

func newTestSettingsService(mgr settings.SettingsManagerInterface) *SettingsService {
	return &SettingsService{settingsManager: mgr}
}

func assertErrorResponse(t *testing.T, resp api.ImplResponse, expectedCode int) {
	t.Helper()
	assert.Equal(t, expectedCode, resp.Code)
	apiErr, ok := resp.Body.(*api.Error)
	require.True(t, ok, "response body should be *api.Error")
	assert.Equal(t, int32(expectedCode), apiErr.Code)
	assert.NotEmpty(t, apiErr.Message)
}

type transformTestCase struct {
	name        string
	setting     manager.Settinger
	enabled     bool
	displayName string
	description string
	stage       string
}

func allTransformTestCases() []transformTestCase {
	return []transformTestCase{
		{
			name:        "enablePolicies",
			setting:     &settings.EnablePolicies{Enabled: true},
			enabled:     true,
			displayName: "Policies",
			description: "Enable to impose restrictions and default values on workloads' fields. For more information, see the Policies guide",
			stage:       "GA",
		},
		{
			name:        "workloadSubmissionV2Enabled",
			setting:     &settings.WorkloadSubmissionV2Enabled{Enabled: false},
			enabled:     false,
			displayName: "Workload submission v2",
			description: "Enable submission of workload with the new API",
			stage:       "Alpha",
		},
		{
			name:        "enableTemplatesV2",
			setting:     &settings.EnableTemplatesV2{Enabled: true},
			enabled:     true,
			displayName: "Flexible workload templates",
			description: "Enable to create workload templates via the new flexible form. For more information, see the Workload Templates guide",
			stage:       "GA",
		},
		{
			name:        "enableHybridSubmission",
			setting:     &settings.EnableHybridSubmission{Enabled: false},
			enabled:     false,
			displayName: "Flexible workload submission",
			description: "Enable to submit workloads via the new flexible form. Allow selecting an existing setup or starting from scratch, reviewing existing setups and understanding policy definitions.",
			stage:       "GA",
		},
		{
			name:        "gpuOverProvisioning",
			setting:     &settings.GpuOverProvisioning{Enabled: true},
			enabled:     true,
			displayName: "GPU resource optimization",
			description: "Enable to set a GPU memory limit for workloads. To enable this on each Run:ai cluster, see the Dynamic GPU fractions guide",
			stage:       "Beta",
		},
	}
}

// unknownSetting is a stub that satisfies manager.Settinger but isn't handled by the transform switch.
type unknownSetting struct{}

func (u *unknownSetting) Category() string            { return "unknown" }
func (u *unknownSetting) Description() string         { return "" }
func (u *unknownSetting) DisplayName() string         { return "" }
func (u *unknownSetting) Key() string                 { return "unknown" }
func (u *unknownSetting) SetDefault()                 {}
func (u *unknownSetting) Stage() manager.Stage        { return manager.GA }
func (u *unknownSetting) Validate(_ interface{}) bool { return true }

// --- TransformSettingByKeyToResponse (parameterized) ---

func TestTransformSettingByKeyToResponse(t *testing.T) {
	mgr := settings.NewSettingsManager()

	for _, tc := range allTransformTestCases() {
		tc := tc
		t.Run("should transform "+tc.name+" correctly", func(t *testing.T) {
			result, err := mgr.TransformSettingByKeyToResponse(tc.setting)
			require.NoError(t, err)
			require.NotNil(t, result)

			switch resp := result.(type) {
			case api.GetEnablePolicies:
				require.NotNil(t, resp.Value.Enabled)
				assert.Equal(t, tc.enabled, *resp.Value.Enabled)
				assert.Equal(t, tc.displayName, resp.DisplayName)
				assert.Equal(t, tc.description, resp.Description)
				assert.Equal(t, tc.stage, resp.Stage)
			case api.GetWorkloadSubmissionV2Enabled:
				require.NotNil(t, resp.Value.Enabled)
				assert.Equal(t, tc.enabled, *resp.Value.Enabled)
				assert.Equal(t, tc.displayName, resp.DisplayName)
				assert.Equal(t, tc.description, resp.Description)
				assert.Equal(t, tc.stage, resp.Stage)
			case api.GetEnableTemplatesV2:
				require.NotNil(t, resp.Value.Enabled)
				assert.Equal(t, tc.enabled, *resp.Value.Enabled)
				assert.Equal(t, tc.displayName, resp.DisplayName)
				assert.Equal(t, tc.description, resp.Description)
				assert.Equal(t, tc.stage, resp.Stage)
			case api.GetEnableHybridSubmission:
				require.NotNil(t, resp.Value.Enabled)
				assert.Equal(t, tc.enabled, *resp.Value.Enabled)
				assert.Equal(t, tc.displayName, resp.DisplayName)
				assert.Equal(t, tc.description, resp.Description)
				assert.Equal(t, tc.stage, resp.Stage)
			case api.GetGpuOverProvisioning:
				require.NotNil(t, resp.Value.Enabled)
				assert.Equal(t, tc.enabled, *resp.Value.Enabled)
				assert.Equal(t, tc.displayName, resp.DisplayName)
				assert.Equal(t, tc.description, resp.Description)
				assert.Equal(t, tc.stage, resp.Stage)
			default:
				t.Fatalf("unexpected response type: %T", result)
			}
		})
	}

	t.Run("should return error for unknown setting type", func(t *testing.T) {
		_, err := mgr.TransformSettingByKeyToResponse(&unknownSetting{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), settings.ErrUnknownSettingKey)
	})
}

// --- TransformSettingsToResponse ---

func TestTransformSettingsToResponse(t *testing.T) {
	mgr := settings.NewSettingsManager()
	allCases := allTransformTestCases()

	settingsList := make([]manager.Settinger, len(allCases))
	for i, tc := range allCases {
		settingsList[i] = tc.setting
	}

	response := mgr.TransformSettingsToResponse(settingsList)
	require.NotNil(t, response)

	require.NotNil(t, response.EnablePolicies.Value.Enabled)
	assert.Equal(t, true, *response.EnablePolicies.Value.Enabled)
	assert.Equal(t, "Policies", response.EnablePolicies.DisplayName)

	require.NotNil(t, response.WorkloadSubmissionV2Enabled.Value.Enabled)
	assert.Equal(t, false, *response.WorkloadSubmissionV2Enabled.Value.Enabled)
	assert.Equal(t, "Workload submission v2", response.WorkloadSubmissionV2Enabled.DisplayName)

	require.NotNil(t, response.EnableTemplatesV2.Value.Enabled)
	assert.Equal(t, true, *response.EnableTemplatesV2.Value.Enabled)
	assert.Equal(t, "Flexible workload templates", response.EnableTemplatesV2.DisplayName)

	require.NotNil(t, response.EnableHybridSubmission.Value.Enabled)
	assert.Equal(t, false, *response.EnableHybridSubmission.Value.Enabled)
	assert.Equal(t, "Flexible workload submission", response.EnableHybridSubmission.DisplayName)

	require.NotNil(t, response.GpuOverProvisioning.Value.Enabled)
	assert.Equal(t, true, *response.GpuOverProvisioning.Value.Enabled)
	assert.Equal(t, "GPU resource optimization", response.GpuOverProvisioning.DisplayName)
}

// --- resolveTenantId ---

func TestResolveTenantId(t *testing.T) {
	t.Run("should resolve tenant id from context", func(t *testing.T) {
		ctx := ctxWithTenant(testTenantId)
		resolved, apiErr := resolveTenantId(ctx, "")
		assert.Nil(t, apiErr)
		assert.Equal(t, testTenantId, resolved)
	})

	t.Run("should return 400 when tenant id is missing from context", func(t *testing.T) {
		_, apiErr := resolveTenantId(context.Background(), "")
		require.NotNil(t, apiErr)
		assert.Equal(t, int32(http.StatusBadRequest), apiErr.Code)
	})

	t.Run("should allow management tenant to use explicit tenant id", func(t *testing.T) {
		ctx := managementCtxWithTenant("mgmt-tenant")
		resolved, apiErr := resolveTenantId(ctx, "target-tenant")
		assert.Nil(t, apiErr)
		assert.Equal(t, "target-tenant", resolved)
	})

	t.Run("should return 401 when non-management tenant uses explicit tenant id", func(t *testing.T) {
		ctx := ctxWithTenant(testTenantId)
		_, apiErr := resolveTenantId(ctx, "other-tenant")
		require.NotNil(t, apiErr)
		assert.Equal(t, int32(http.StatusUnauthorized), apiErr.Code)
	})
}

func TestGetWorkloadsServiceSettings(t *testing.T) {
	t.Run("should return all settings successfully", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		settingsList := []manager.Settinger{
			&settings.EnablePolicies{Enabled: true},
			&settings.GpuOverProvisioning{Enabled: false},
		}
		expectedResponse := &api.WorkloadsSettingsGetResponse{
			EnablePolicies: api.GetEnablePolicies{
				Value: api.GetEnableSettingsValue{Enabled: ptr.To(true)},
			},
			GpuOverProvisioning: api.GetGpuOverProvisioning{
				Value: api.GetEnableSettingsValue{Enabled: ptr.To(false)},
			},
		}
		mgr.On("GetAllSettings", testTenantId).Return(settingsList, nil)
		mgr.On("TransformSettingsToResponse", settingsList).Return(expectedResponse)

		resp, err := svc.GetWorkloadsServiceSettings(ctx, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Body)
		mgr.AssertExpectations(t)
	})

	t.Run("should return 400 when tenant id is missing", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)

		resp, err := svc.GetWorkloadsServiceSettings(context.Background(), "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 500 when GetAllSettings fails", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		mgr.On("GetAllSettings", testTenantId).Return(nil, errors.New("db error"))

		resp, err := svc.GetWorkloadsServiceSettings(ctx, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusInternalServerError)
	})

	t.Run("should return 401 when non-management tenant uses explicit tenant id", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		resp, err := svc.GetWorkloadsServiceSettings(ctx, "other-tenant")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusUnauthorized)
	})

	t.Run("should allow management tenant to query another tenant", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := managementCtxWithTenant("mgmt-tenant")

		settingsList := []manager.Settinger{}
		expectedResponse := &api.WorkloadsSettingsGetResponse{}
		mgr.On("GetAllSettings", "target-tenant").Return(settingsList, nil)
		mgr.On("TransformSettingsToResponse", settingsList).Return(expectedResponse)

		resp, err := svc.GetWorkloadsServiceSettings(ctx, "target-tenant")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		mgr.AssertExpectations(t)
	})
}

func TestGetWorkloadsServiceSettingByKey(t *testing.T) {
	t.Run("should return setting by key successfully", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		settingValue := &settings.EnablePolicies{Enabled: true}
		expectedBody := api.GetEnablePolicies{
			Value:       api.GetEnableSettingsValue{Enabled: ptr.To(true)},
			DisplayName: "Policies",
		}
		mgr.On("GetSettingByKey", api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES, testTenantId).Return(settingValue, nil)
		mgr.On("TransformSettingByKeyToResponse", settingValue).Return(expectedBody, nil)

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedBody, resp.Body)
		mgr.AssertExpectations(t)
	})

	t.Run("should return 400 when tenant id is missing", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)

		resp, err := svc.GetWorkloadsServiceSettingByKey(context.Background(), api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 400 for unknown setting key", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		mgr.On("GetSettingByKey", api.WorkloadsSettingsKeyEnum("invalidKey"), testTenantId).
			Return(nil, errors.New(settings.ErrUnknownSettingKey+": invalidKey"))

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, "invalidKey", "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 500 when GetSettingByKey fails", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		mgr.On("GetSettingByKey", api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED, testTenantId).
			Return(nil, errors.New("db connection failed"))

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusInternalServerError)
	})

	t.Run("should return 500 when TransformSettingByKeyToResponse fails", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		settingValue := &settings.EnableHybridSubmission{Enabled: false}
		mgr.On("GetSettingByKey", api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION, testTenantId).Return(settingValue, nil)
		mgr.On("TransformSettingByKeyToResponse", settingValue).Return(nil, errors.New("transform failed"))

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusInternalServerError)
	})

	t.Run("should return 401 when non-management tenant uses explicit tenant id", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2, "other-tenant")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusUnauthorized)
	})

	t.Run("should allow management tenant to query another tenant", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := managementCtxWithTenant("mgmt-tenant")

		settingValue := &settings.GpuOverProvisioning{Enabled: true}
		expectedBody := api.GetGpuOverProvisioning{
			Value:       api.GetEnableSettingsValue{Enabled: ptr.To(true)},
			DisplayName: "GPU resource optimization",
		}
		mgr.On("GetSettingByKey", api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING, "target-tenant").Return(settingValue, nil)
		mgr.On("TransformSettingByKeyToResponse", settingValue).Return(expectedBody, nil)

		resp, err := svc.GetWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING, "target-tenant")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedBody, resp.Body)
		mgr.AssertExpectations(t)
	})
}

func TestUpdateWorkloadsServiceSettingByKey(t *testing.T) {
	t.Run("should update setting successfully", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		mgr.On("UpdateSettingByKey", ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES, testTenantId, updateReq).Return(nil)

		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES, updateReq, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, updateReq, resp.Body)
		mgr.AssertExpectations(t)
	})

	t.Run("should return 400 when tenant id is missing", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		resp, err := svc.UpdateWorkloadsServiceSettingByKey(context.Background(), api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION, updateReq, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 400 for unknown setting key", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		mgr.On("UpdateSettingByKey", ctx, api.WorkloadsSettingsKeyEnum("invalidKey"), testTenantId, updateReq).
			Return(errors.New(settings.ErrUnknownSettingKey + ": invalidKey"))

		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, "invalidKey", updateReq, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 400 for validation error", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		mgr.On("UpdateSettingByKey", ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2, testTenantId, updateReq).
			Return(errors.New(settings.ErrValidationFailed + ": cannot enable enableTemplatesV2 while enableHybridSubmission is disabled"))

		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2, updateReq, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusBadRequest)
	})

	t.Run("should return 500 when UpdateSettingByKey fails", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(false)}
		mgr.On("UpdateSettingByKey", ctx, api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING, testTenantId, updateReq).
			Return(errors.New("db write failed"))

		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING, updateReq, "")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusInternalServerError)
	})

	t.Run("should return 401 when non-management tenant uses explicit tenant id", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := ctxWithTenant(testTenantId)

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED, updateReq, "other-tenant")
		assert.NoError(t, err)
		assertErrorResponse(t, resp, http.StatusUnauthorized)
	})

	t.Run("should allow management tenant to update another tenant", func(t *testing.T) {
		mgr := new(mockSettingsManager)
		svc := newTestSettingsService(mgr)
		ctx := managementCtxWithTenant("mgmt-tenant")

		updateReq := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
		mgr.On("UpdateSettingByKey", ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2, "target-tenant", updateReq).Return(nil)

		resp, err := svc.UpdateWorkloadsServiceSettingByKey(ctx, api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2, updateReq, "target-tenant")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, updateReq, resp.Body)
		mgr.AssertExpectations(t)
	})
}
