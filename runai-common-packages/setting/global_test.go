package setting

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	dal "github.com/run-ai/runai/runai-common-packages/setting/database"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
	managerMock "github.com/run-ai/runai/runai-common-packages/setting/manager/mocks"
	settingMocks "github.com/run-ai/runai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	tenantID = "test-tenant-123"
)

var (
	settingsDal         *managerMock.SettingsDal
	backendClientMock   *settingMocks.BackendClient
	settingsManagerMock *settingMocks.SettingsManager
)

func initTest(t *testing.T) {
	once = sync.Once{}
	settingsManager = nil
	backendClient = nil
	settingsDal = managerMock.NewSettingsDal(t)
	settingsManagerMock = settingMocks.NewSettingsManager(t)
	backendClientMock = settingMocks.NewBackendClient(t)

	config := &SettingConfig{
		BackendClient: backendClientMock,
	}
	Init(config, settingsDal)
}

func initTestWithoutBackendClient(t *testing.T) {
	once = sync.Once{}
	settingsManager = nil
	backendClient = nil
	settingsDal = managerMock.NewSettingsDal(t)
	settingsManagerMock = settingMocks.NewSettingsManager(t)
	backendClientMock = settingMocks.NewBackendClient(t)

	config := &SettingConfig{BackendClient: nil}
	Init(config, settingsDal)
}

func initTestWithoutSettingsManager(t *testing.T) {
	once = sync.Once{}
	settingsManager = nil
}

func TestUpdateBuilder_Creation(t *testing.T) {
	value := TestSetting1{Name: "test", Count: 42}

	builder := Update[*TestSetting1](tenantID, value)

	assert.NotNil(t, builder)
	assert.Equal(t, tenantID, builder.tenantId)
	assert.Equal(t, value, builder.value)
	assert.True(t, builder.updateByUser)
}

func TestUpdateBuilder_UpdatedByUser(t *testing.T) {
	value := TestSetting1{Name: "test", Count: 42}

	builder := Update[*TestSetting1](tenantID, value).UpdatedByUser()

	assert.NotNil(t, builder)
	assert.Equal(t, tenantID, builder.tenantId)
	assert.Equal(t, value, builder.value)
	assert.True(t, builder.updateByUser)
}

func TestUpdateBuilder_UpdatedBySystem(t *testing.T) {
	value := TestSetting1{Name: "test", Count: 42}

	builder := Update[*TestSetting1](tenantID, value).UpdatedBySystem()

	assert.NotNil(t, builder)
	assert.Equal(t, tenantID, builder.tenantId)
	assert.Equal(t, value, builder.value)
	assert.False(t, builder.updateByUser)
}

func TestUpdateBuilder_MethodChaining(t *testing.T) {
	value := TestSetting1{Name: "test", Count: 42}

	builder := Update[*TestSetting1](tenantID, value).UpdatedByUser().UpdatedBySystem()
	assert.False(t, builder.updateByUser)

	builder = Update[*TestSetting1](tenantID, value).UpdatedBySystem().UpdatedByUser()
	assert.True(t, builder.updateByUser)
}

func TestUpdateBuilder_Execute_Success_UserUpdate(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	value := TestSetting1{Name: "test", Count: 42}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, valueBytes).Return(nil).Once()
	err = Update[*TestSetting1](tenantID, value).UpdatedByUser().Execute()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Success_SystemUpdate(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	value := TestSetting1{Name: "system-update", Count: 100}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, false, valueBytes).Return(nil).Once()
	err = Update[*TestSetting1](tenantID, value).UpdatedBySystem().Execute()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Success_DefaultBehavior(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	value := TestSetting1{Name: "default-test", Count: 50}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, valueBytes).Return(nil).Once()
	err = Update[*TestSetting1](tenantID, value).Execute()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_UnregisteredSetting(t *testing.T) {
	initTest(t)
	value := TestSetting1{Name: "test", Count: 42}

	err := Update[*TestSetting1](tenantID, value).Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting is not registered")
	assert.True(t, isBadRequest(err), "expected bad request error")
}

func TestUpdateBuilder_Execute_ValidationFailure(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	value := TestSetting1{Name: "", Count: -1}
	err = Update[*TestSetting1](tenantID, value).Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not valid")
	assert.True(t, isBadRequest(err), "expected bad request error")
}

func TestUpdateBuilder_Execute_DatabaseError(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	value := TestSetting1{Name: "test", Count: 42}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, valueBytes).Return(errors.New("database connection failed")).Once()
	err = Update[*TestSetting1](tenantID, value).Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_MultipleSettingTypes(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	value1 := TestSetting1{Name: "test1", Count: 42}
	value1Bytes, _ := json.Marshal(value1)
	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, value1Bytes).Return(nil).Once()

	err = Update[*TestSetting1](tenantID, value1).UpdatedByUser().Execute()
	assert.NoError(t, err)

	value2 := TestSetting2{Enabled: true, Message: "hello"}
	value2Bytes, _ := json.Marshal(value2)
	settingsDal.On("SaveSetting", "test-setting2", tenantID, false, value2Bytes).Return(nil).Once()

	err = Update[*TestSetting2](tenantID, value2).UpdatedBySystem().Execute()
	assert.NoError(t, err)

	settingsDal.AssertExpectations(t)
}

func TestUpdateAndGet_Integration(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	updateValue := TestSetting1{Name: "updated-name", Count: 99}
	updateValueBytes, _ := json.Marshal(updateValue)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, updateValueBytes).Return(nil).Once()
	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             updateValueBytes,
		TenantID:          tenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", tenantID).Return(dbSetting, nil).Once()

	err = Update[*TestSetting1](tenantID, updateValue).Execute()
	assert.NoError(t, err)

	result, err := Get[*TestSetting1](tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "updated-name", (*result).Name, "Name should match updated value")
	assert.Equal(t, 99, (*result).Count, "Count should match updated value")

	settingsDal.AssertExpectations(t)
}

func TestUpdateAndGet_Integration_SystemUpdate(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting2]()
	require.NoError(t, err)

	updateValue := TestSetting2{Enabled: false, Message: "system-updated"}
	updateValueBytes, _ := json.Marshal(updateValue)
	settingsDal.On("SaveSetting", "test-setting2", tenantID, false, updateValueBytes).Return(nil).Once()

	dbSetting := &dal.Setting{
		Key:               "test-setting2",
		Value:             updateValueBytes,
		TenantID:          tenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: false,
	}
	settingsDal.On("GetSettingByKey", "test-setting2", tenantID).Return(dbSetting, nil).Once()

	err = Update[*TestSetting2](tenantID, updateValue).UpdatedBySystem().Execute()
	assert.NoError(t, err)

	result, err := Get[*TestSetting2](tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, false, (*result).Enabled, "Enabled should match updated value")
	assert.Equal(t, "system-updated", (*result).Message, "Message should match updated value")

	settingsDal.AssertExpectations(t)
}

func TestUpdateAndGet_Integration_MultipleUpdates(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	firstUpdate := TestSetting1{Name: "first-update", Count: 10}
	firstUpdateBytes, _ := json.Marshal(firstUpdate)

	secondUpdate := TestSetting1{Name: "second-update", Count: 20}
	secondUpdateBytes, _ := json.Marshal(secondUpdate)

	settingsDal.On("SaveSetting", "test-setting1", tenantID, true, firstUpdateBytes).Return(nil).Once()
	settingsDal.On("SaveSetting", "test-setting1", tenantID, false, secondUpdateBytes).Return(nil).Once()

	finalDbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             secondUpdateBytes,
		TenantID:          tenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: false,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", tenantID).Return(finalDbSetting, nil).Once()

	err = Update[*TestSetting1](tenantID, firstUpdate).UpdatedByUser().Execute()
	assert.NoError(t, err)

	err = Update[*TestSetting1](tenantID, secondUpdate).UpdatedBySystem().Execute()
	assert.NoError(t, err)

	result, err := Get[*TestSetting1](tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "second-update", (*result).Name, "Name should match final updated value")
	assert.Equal(t, 20, (*result).Count, "Count should match final updated value")
	settingsDal.AssertExpectations(t)
}

func TestGet_Success_WithDatabaseValue(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-db-value")
	settingValue := TestSetting1{Name: "from-db", Count: 99}
	valueBytes, _ := json.Marshal(settingValue)

	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             valueBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}

	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil).Once()
	result, err := Get[*TestSetting1](uniqueTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, *result)

	assert.Equal(t, "from-db", (*result).Name, "result should have Name field with correct value")
	assert.Equal(t, 99, (*result).Count, "result should have Count field with correct value")

	settingsDal.AssertExpectations(t)
}

func TestGet_Success_NotFoundInDatabase_UsesDefaults(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-not-found")
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(nil, gorm.ErrRecordNotFound).Once()
	result, err := Get[*TestSetting1](uniqueTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, *result)

	assert.Equal(t, "default name", (*result).Name, "result should have default Name when using defaults")
	assert.Equal(t, 1, (*result).Count, "result should have default Count when using defaults")

	settingsDal.AssertExpectations(t)
}

func TestGet_Success_NilValueInDatabase_UsesDefaults(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-nil-value")
	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             nil,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: false,
	}

	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil).Once()
	result, err := Get[*TestSetting1](uniqueTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, *result)

	assert.Equal(t, "default name", (*result).Name, "result should have default Name when database value is nil")
	assert.Equal(t, 1, (*result).Count, "result should have default Count when database value is nil")

	settingsDal.AssertExpectations(t)
}

func TestGet_Error_UnregisteredSetting(t *testing.T) {
	initTest(t)

	uniqueTenantID := generateUniqueTenantID("test-unregistered")
	result, err := Get[*TestSetting1](uniqueTenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "setting is not registered")
	assert.True(t, isBadRequest(err), "expected bad request error")
}

func TestGet_Error_DatabaseError(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-db-error")
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(nil, errors.New("database connection failed")).Once()

	result, err := Get[*TestSetting1](uniqueTenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
	assert.Contains(t, err.Error(), "database connection failed")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	settingsDal.AssertExpectations(t)
}

func TestGetAll_Success_AllSettings_NoCategory(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	setting1Value := TestSetting1{Name: "setting1", Count: 10}
	setting1Bytes, _ := json.Marshal(setting1Value)

	setting2Value := TestSetting2{Enabled: true, Message: "test"}
	setting2Bytes, _ := json.Marshal(setting2Value)

	dbSettings := &[]dal.Setting{
		{
			Key:               "test-setting1",
			Value:             setting1Bytes,
			TenantID:          tenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: true,
		},
		{
			Key:               "test-setting2",
			Value:             setting2Bytes,
			TenantID:          tenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: false,
		},
	}

	settingsDal.On("GetSettingsByTenantId", tenantID).Return(dbSettings, nil).Once()

	results, err := GetAll(tenantID, "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	var setting1 *TestSetting1
	var setting2 *TestSetting2
	for _, result := range results {
		if s1, ok := result.(*TestSetting1); ok {
			setting1 = s1
		} else if s2, ok := result.(*TestSetting2); ok {
			setting2 = s2
		}
	}

	assert.NotNil(t, setting1, "TestSetting1 should be found in results")
	assert.Equal(t, "setting1", setting1.Name, "TestSetting1 should have correct Name from database")
	assert.Equal(t, 10, setting1.Count, "TestSetting1 should have correct Count from database")

	assert.NotNil(t, setting2, "TestSetting2 should be found in results")
	assert.Equal(t, true, setting2.Enabled, "TestSetting2 should have correct Enabled from database")
	assert.Equal(t, "test", setting2.Message, "TestSetting2 should have correct Message from database")

	settingsDal.AssertExpectations(t)
}

func TestGetAll_Success_WithCategoryFilter(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	dbSettings := &[]dal.Setting{
		{
			Key:      "test-setting1",
			Value:    []byte(`{"name":"test","count":1}`),
			TenantID: tenantID,
		},
		{
			Key:      "test-setting2",
			Value:    []byte(`{"enabled":true,"message":"test"}`),
			TenantID: tenantID,
		},
	}

	settingsDal.On("GetSettingsByTenantId", tenantID).Return(dbSettings, nil).Once()

	results, err := GetAll(tenantID, "test")

	assert.NoError(t, err)
	assert.Len(t, results, 1)

	setting1, ok := results[0].(*TestSetting1)
	assert.True(t, ok, "Result should be *TestSetting1 type")
	assert.Equal(t, "test", setting1.Name, "TestSetting1 should have correct Name from database")
	assert.Equal(t, 1, setting1.Count, "TestSetting1 should have correct Count from database")

	settingsDal.AssertExpectations(t)
}

func TestGetAll_Success_EmptyDatabase(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-empty-db")
	dbSettings := &[]dal.Setting{}

	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(dbSettings, nil).Once()

	results, err := GetAll(uniqueTenantID, "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	var setting1 *TestSetting1
	var setting2 *TestSetting2
	for _, result := range results {
		if s1, ok := result.(*TestSetting1); ok {
			setting1 = s1
		} else if s2, ok := result.(*TestSetting2); ok {
			setting2 = s2
		}
	}

	assert.NotNil(t, setting1, "TestSetting1 should be found in results")
	assert.Equal(t, "default name", setting1.Name, "TestSetting1 should have default Name when using defaults")
	assert.Equal(t, 1, setting1.Count, "TestSetting1 should have default Count when using defaults")

	assert.NotNil(t, setting2, "TestSetting2 should be found in results")
	assert.Equal(t, true, setting2.Enabled, "TestSetting2 should have default Enabled when using defaults")
	assert.Equal(t, "default message", setting2.Message, "TestSetting2 should have default Message when using defaults")

	settingsDal.AssertExpectations(t)
}

func TestGetAll_Error_DatabaseError(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-getall-error")
	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(nil, errors.New("database connection failed")).Once()
	results, err := GetAll(uniqueTenantID, "")

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to get settings by tenantId")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	assert.Contains(t, err.Error(), "database connection failed")
	settingsDal.AssertExpectations(t)
}

func TestRegister_Success(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	assert.NoError(t, err)

	err = Register[*TestSetting1]()
	assert.NoError(t, err)
}

func TestRegister_MultipleSettings(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	assert.NoError(t, err)

	err = Register[*TestSetting2]()
	assert.NoError(t, err)
}

func TestValidateValue_Success_TestSetting1_ValidValue(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	validValue := &TestSetting1{Name: "test", Count: 10}
	result := Validate[*TestSetting1](validValue)

	assert.True(t, result, "Valid value should pass validation")
}

func TestValidateValue_Failure_TestSetting1_BothInvalid(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	invalidValue := &TestSetting1{Name: "", Count: -1}
	result := Validate[*TestSetting1](invalidValue)

	assert.False(t, result, "Value with both empty name and negative count should fail validation")
}

func TestValidateValue_Failure_WrongType(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	wrongTypeSetting := &TestSetting1{Name: "test", Count: 10}
	result := Validate[*TestSetting2](wrongTypeSetting)

	assert.False(t, result, "Validation should fail for wrong type")
}

// ============================================================================
// CACHE INTEGRATION TESTS WITH GLOBAL API
// ============================================================================
// Note: Cache tests are in cache_test.go. These tests verify cache behavior
// through the public API by checking the number of DB calls.

func TestGlobalAPI_Get_AfterUpdate_WithCache_FetchesNewValue(t *testing.T) {
	initTestWithCacheForGlobalAPI(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-get-after-update")

	initialValue := TestSetting1{Name: "initial", Count: 1}
	initialBytes, _ := json.Marshal(initialValue)
	dbSetting1 := &dal.Setting{
		Key:               "test-setting1",
		Value:             initialBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting1, nil).Once()

	result1, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "initial", (*result1).Name)

	updateValue := TestSetting1{Name: "updated", Count: 2}
	updateValueBytes, _ := json.Marshal(updateValue)
	settingsDal.On("SaveSetting", "test-setting1", uniqueTenantID, true, updateValueBytes).Return(nil).Once()

	err = Update[*TestSetting1](uniqueTenantID, updateValue).Execute()
	assert.NoError(t, err)

	updatedBytes, _ := json.Marshal(updateValue)
	dbSetting2 := &dal.Setting{
		Key:               "test-setting1",
		Value:             updatedBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting2, nil).Once()

	result2, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "updated", (*result2).Name)
	assert.Equal(t, 2, (*result2).Count)

	settingsDal.AssertExpectations(t)
	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 2)
}

func TestGlobalAPI_GetAll_AfterUpdate_WithCache_FetchesNewList(t *testing.T) {
	initTestWithCacheForGlobalAPI(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-getall-after-update")

	initialSettings := &[]dal.Setting{
		{
			Key:      "test-setting1",
			Value:    []byte(`{"name":"initial","count":1}`),
			TenantID: uniqueTenantID,
		},
	}
	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(initialSettings, nil).Once()

	results, err := GetAll(uniqueTenantID, "")
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	updateValue := TestSetting1{Name: "updated", Count: 99}
	updateValueBytes, _ := json.Marshal(updateValue)
	settingsDal.On("SaveSetting", "test-setting1", uniqueTenantID, false, updateValueBytes).Return(nil).Once()

	err = Update[*TestSetting1](uniqueTenantID, updateValue).UpdatedBySystem().Execute()
	assert.NoError(t, err)

	updatedSettings := &[]dal.Setting{
		{
			Key:      "test-setting1",
			Value:    []byte(`{"name":"updated","count":99}`),
			TenantID: uniqueTenantID,
		},
	}
	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(updatedSettings, nil).Once()

	results, err = GetAll(uniqueTenantID, "")
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	settingsDal.AssertExpectations(t)
	settingsDal.AssertNumberOfCalls(t, "GetSettingsByTenantId", 2)
}

func TestGlobalAPI_Cache_DifferentTenantsHaveSeparateCaches(t *testing.T) {
	initTestWithCacheForGlobalAPI(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	tenant1 := generateUniqueTenantID("tenant-1")
	tenant2 := generateUniqueTenantID("tenant-2")

	value1 := TestSetting1{Name: "tenant1-value", Count: 1}
	bytes1, _ := json.Marshal(value1)
	dbSetting1 := &dal.Setting{
		Key:      "test-setting1",
		Value:    bytes1,
		TenantID: tenant1,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", tenant1).Return(dbSetting1, nil).Once()

	result1, err := Get[*TestSetting1](tenant1)
	assert.NoError(t, err)
	assert.Equal(t, "tenant1-value", (*result1).Name)

	value2 := TestSetting1{Name: "tenant2-value", Count: 2}
	bytes2, _ := json.Marshal(value2)
	dbSetting2 := &dal.Setting{
		Key:      "test-setting1",
		Value:    bytes2,
		TenantID: tenant2,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", tenant2).Return(dbSetting2, nil).Once()

	result2, err := Get[*TestSetting1](tenant2)
	assert.NoError(t, err)
	assert.Equal(t, "tenant2-value", (*result2).Name)

	result1Again, err := Get[*TestSetting1](tenant1)
	assert.NoError(t, err)
	assert.Equal(t, "tenant1-value", (*result1Again).Name)

	result2Again, err := Get[*TestSetting1](tenant2)
	assert.NoError(t, err)
	assert.Equal(t, "tenant2-value", (*result2Again).Name)

	settingsDal.AssertExpectations(t)
	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 2)
}

func TestCheckInitialized_Register_BeforeInit(t *testing.T) {
	initTestWithoutSettingsManager(t)
	err := Register[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting package is not initialized")
	assert.True(t, isInternalServerError(err), "expected internal server error")
}

func TestCheckInitialized_Get_BeforeInit(t *testing.T) {
	initTestWithoutSettingsManager(t)
	result, err := Get[*TestSetting1](tenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "setting package is not initialized")
	assert.True(t, isInternalServerError(err), "expected internal server error")
}

func TestCheckInitialized_GetAll_BeforeInit(t *testing.T) {
	initTestWithoutSettingsManager(t)
	results, err := GetAll(tenantID, "")

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "setting package is not initialized")
	assert.True(t, isInternalServerError(err), "expected internal server error")
}

func TestCheckInitialized_Update_BeforeInit(t *testing.T) {
	initTestWithoutSettingsManager(t)
	value := TestSetting1{Name: "test", Count: 42}
	err := Update[*TestSetting1](tenantID, value).Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting package is not initialized")
	assert.True(t, isInternalServerError(err), "expected internal server error")
}

func TestMigrate_Success_ShouldNotMigrate(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(false, nil).Once()

	err = Migrate[*TestSetting1]()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
	settingsDal.AssertNotCalled(t, "SaveSettingByKeyForTenantsList")
}

func TestMigrate_Error_ShouldMigrateSettingFails(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(false, errors.New("database connection error")).Once()

	err = Migrate[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection error")
	settingsDal.AssertExpectations(t)
	settingsDal.AssertNotCalled(t, "SaveSettingByKeyForTenantsList")
}

func TestMigrate_Error_BackendClientNotConfigured(t *testing.T) {
	initTestWithoutBackendClient(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	err = Migrate[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend client not configured")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Error_GetSettingByKeyForAllTenantsFails(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(nil, errors.New("backend service unavailable")).Once()

	err = Migrate[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to migrate setting by key")
	assert.Contains(t, err.Error(), "backend service unavailable")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Success_EmptyTenantSettings(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(&manager.SettingValuesForAllTenantsBackendResponse{Values: []manager.TenantSettingValueForBackendResponse{}}, nil).Once()

	err = Migrate[*TestSetting1]()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
	settingsDal.AssertNotCalled(t, "SaveSettingByKeyForTenantsList")
}

func TestMigrate_Success_WithTenantSettings(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	tenantSettings := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting1",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting1{Name: "tenant1-setting", Count: 10}},
			{TenantId: 2, Value: &TestSetting1{Name: "tenant2-setting", Count: 20}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(tenantSettings, nil).Once()

	settingsDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.AnythingOfType("[]database.TenantSettingInput")).Return(nil).Once()

	err = Migrate[*TestSetting1]()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Failure_NoTransformation(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting3]()
	require.NoError(t, err)

	tenantSettings := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting3",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting3{Enabled: true, Message: "tenant1-setting"}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting3").Return(tenantSettings, nil).Once()

	settingsDal.On("ShouldMigrateSetting", "test-setting3").Return(true, nil).Once()

	err = Migrate[*TestSetting3]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting does not support migration transformation")
	assert.True(t, isBadRequest(err), "expected bad request error")
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Error_UpdateAllTenantsByKeyFails(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	tenantSettings := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting1",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting1{Name: "tenant1-setting", Count: 10}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(tenantSettings, nil).Once()

	settingsDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.AnythingOfType("[]database.TenantSettingInput")).Return(errors.New("database write failed")).Once()

	err = Migrate[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to batch update setting")
	assert.True(t, isInternalServerError(err), "expected internal server error")
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Error_ValidationFails(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	tenantSettings := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting1",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting1{Name: "", Count: -1}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(tenantSettings, nil).Once()

	err = Migrate[*TestSetting1]()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid")
	assert.True(t, isBadRequest(err), "expected bad request error")
	settingsDal.AssertExpectations(t)
}

func TestMigrate_Success_MultipleTenants(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	tenantSettings := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting1",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting1{Name: "setting1", Count: 1}},
			{TenantId: 2, Value: &TestSetting1{Name: "setting2", Count: 2}},
			{TenantId: 3, Value: &TestSetting1{Name: "setting3", Count: 3}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(tenantSettings, nil).Once()

	settingsDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		return len(inputs) == 3
	})).Return(nil).Once()

	err = Migrate[*TestSetting1]()

	assert.NoError(t, err)
	settingsDal.AssertExpectations(t)
}

func TestMigrate_DifferentSettingTypes(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()
	setting1Values := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting1",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting1{Name: "migrated", Count: 100}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting1").Return(setting1Values, nil).Once()
	settingsDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.AnythingOfType("[]database.TenantSettingInput")).Return(nil).Once()

	err = Migrate[*TestSetting1]()
	assert.NoError(t, err)

	settingsDal.On("ShouldMigrateSetting", "test-setting2").Return(true, nil).Once()
	setting2Values := &manager.SettingValuesForAllTenantsBackendResponse{
		Key: "test-setting2",
		Values: []manager.TenantSettingValueForBackendResponse{
			{TenantId: 1, Value: &TestSetting2{Enabled: true, Message: "migrated message"}},
		},
	}
	backendClientMock.On("GetSettingByKeyForAllTenants", mock.Anything, "test-setting2").Return(setting2Values, nil).Once()
	settingsDal.On("SaveSettingByKeyForTenantsList", "test-setting2", mock.AnythingOfType("[]database.TenantSettingInput")).Return(nil).Once()

	err = Migrate[*TestSetting2]()
	assert.NoError(t, err)

	settingsDal.AssertExpectations(t)
}
