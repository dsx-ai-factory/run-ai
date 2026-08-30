package manager

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	cachetesting "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	dal "github.com/run-ai/runai/runai-common-packages/setting/database"
	mockManager "github.com/run-ai/runai/runai-common-packages/setting/manager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

const testTenantID = "test-tenant-123"

func TestNewManager_WithCustomDal(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}

	manager := NewManager(mockDal, nil)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.settingsMap)
	assert.Equal(t, mockDal, manager.settingsDal)
}

func TestNewManager_WithNilDal(t *testing.T) {
	manager := NewManager(nil, nil)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.settingsMap)
	assert.NotNil(t, manager.settingsDal)
}

func TestManager_Register_Success(t *testing.T) {
	manager := NewManager(nil, nil)
	setting := &TestSetting1{}

	err := manager.Register(setting)

	assert.NoError(t, err)
	assert.Contains(t, manager.settingsMap, "test-setting1")
	assert.Equal(t, setting, manager.settingsMap["test-setting1"])
}

func TestManager_Register_Idempotent(t *testing.T) {
	manager := NewManager(nil, nil)
	setting1 := &TestSetting1{Name: "first", Count: 1}
	setting2 := &TestSetting1{Name: "second", Count: 2}

	err1 := manager.Register(setting1)
	err2 := manager.Register(setting2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, setting1, manager.settingsMap["test-setting1"])
}

func TestManager_Register_MultipleSettings(t *testing.T) {
	manager := NewManager(nil, nil)
	setting1 := &TestSetting1{}
	setting2 := &TestSetting2{}

	err1 := manager.Register(setting1)
	err2 := manager.Register(setting2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, manager.settingsMap, 2)
	assert.Contains(t, manager.settingsMap, "test-setting1")
	assert.Contains(t, manager.settingsMap, "test-setting2")
}

func TestManager_Get_UnregisteredSetting(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	result, err := manager.Get(ctx, testTenantID, "nonexistent-key")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "setting is not registered")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for unregistered setting")
}

func TestManager_Get_DatabaseError(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	setting := &TestSetting1{}
	manager.Register(setting)
	ctx := context.Background()

	mockDal.On("GetSettingByKey", "test-setting1", testTenantID).Return(nil, errors.New("database connection failed")).Once()
	result, err := manager.Get(ctx, testTenantID, "test-setting1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
	assert.Contains(t, err.Error(), "database connection failed")
	assert.True(t, isInternalServerError(err), "expected ErrInternal for database error")
	mockDal.AssertExpectations(t)
}

func TestManager_Get_RecordNotFound_UsesDefaults(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	setting := &TestSetting1{}
	manager.Register(setting)
	ctx := context.Background()

	mockDal.On("GetSettingByKey", "test-setting1", testTenantID).Return(nil, gorm.ErrRecordNotFound).Once()
	result, err := manager.Get(ctx, testTenantID, "test-setting1")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	testResult := result.(*TestSetting1)
	assert.Equal(t, "default name", testResult.Name)
	assert.Equal(t, 1, testResult.Count)

	mockDal.AssertExpectations(t)
}

func TestManager_Get_WithDatabaseValue_Success(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	setting := &TestSetting1{}
	manager.Register(setting)
	ctx := context.Background()

	settingValue := TestSetting1{Name: "from-db", Count: 99}
	valueBytes, _ := json.Marshal(settingValue)
	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             valueBytes,
		TenantID:          testTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}

	mockDal.On("GetSettingByKey", "test-setting1", testTenantID).Return(dbSetting, nil).Once()
	result, err := manager.Get(ctx, testTenantID, "test-setting1")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	testResult := result.(*TestSetting1)
	assert.Equal(t, "from-db", testResult.Name)
	assert.Equal(t, 99, testResult.Count)

	mockDal.AssertExpectations(t)
}

func TestManager_Get_WithNilValue_UsesDefaults(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	setting := &TestSetting1{}
	manager.Register(setting)
	ctx := context.Background()

	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             nil,
		TenantID:          testTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: false,
	}

	mockDal.On("GetSettingByKey", "test-setting1", testTenantID).Return(dbSetting, nil).Once()
	result, err := manager.Get(ctx, testTenantID, "test-setting1")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	testResult := result.(*TestSetting1)
	assert.Equal(t, "default name", testResult.Name)
	assert.Equal(t, 1, testResult.Count)
	mockDal.AssertExpectations(t)
}

func TestManager_Get_InvalidJSON_Error(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	setting := &TestSetting1{}
	manager.Register(setting)
	ctx := context.Background()

	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             []byte("invalid json"),
		TenantID:          testTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}

	mockDal.On("GetSettingByKey", "test-setting1", testTenantID).Return(dbSetting, nil).Once()
	result, err := manager.Get(ctx, testTenantID, "test-setting1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to unmarshal value")
	assert.True(t, isInternalServerError(err), "expected ErrInternal for unmarshal error")
	mockDal.AssertExpectations(t)
}

func TestManager_GetAll_DatabaseError(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	mockDal.On("GetSettingsByTenantId", testTenantID).Return(nil, errors.New("database error")).Once()
	results, err := manager.GetAll(ctx, testTenantID, "")

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to get settings by tenantId")
	assert.True(t, isInternalServerError(err), "expected ErrInternal for database error")
	mockDal.AssertExpectations(t)
}

func TestManager_GetAll_EmptyDatabase_UsesDefaults(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting1 := &TestSetting1{}
	setting2 := &TestSetting2{}
	manager.Register(setting1)
	manager.Register(setting2)

	dbSettings := &[]dal.Setting{}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbSettings, nil).Once()

	results, err := manager.GetAll(ctx, testTenantID, "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	var testResult1 *TestSetting1
	var testResult2 *TestSetting2
	for _, result := range results {
		if s1, ok := result.(*TestSetting1); ok {
			testResult1 = s1
		} else if s2, ok := result.(*TestSetting2); ok {
			testResult2 = s2
		}
	}

	assert.NotNil(t, testResult1)
	assert.Equal(t, "default name", testResult1.Name)
	assert.Equal(t, 1, testResult1.Count)

	assert.NotNil(t, testResult2)
	assert.Equal(t, true, testResult2.Enabled)
	assert.Equal(t, "default message", testResult2.Message)
	mockDal.AssertExpectations(t)
}

func TestManager_GetAll_WithDatabaseValues_Success(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting1 := &TestSetting1{}
	setting2 := &TestSetting2{}
	manager.Register(setting1)
	manager.Register(setting2)

	setting1Value := TestSetting1{Name: "db-name", Count: 42}
	setting1Bytes, _ := json.Marshal(setting1Value)

	setting2Value := TestSetting2{Enabled: false, Message: "db-message"}
	setting2Bytes, _ := json.Marshal(setting2Value)

	dbSettings := &[]dal.Setting{
		{
			Key:               "test-setting1",
			Value:             setting1Bytes,
			TenantID:          testTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: true,
		},
		{
			Key:               "test-setting2",
			Value:             setting2Bytes,
			TenantID:          testTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: false,
		},
	}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbSettings, nil).Once()

	results, err := manager.GetAll(ctx, testTenantID, "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	var testResult1 *TestSetting1
	var testResult2 *TestSetting2
	for _, result := range results {
		if s1, ok := result.(*TestSetting1); ok {
			testResult1 = s1
		} else if s2, ok := result.(*TestSetting2); ok {
			testResult2 = s2
		}
	}

	assert.NotNil(t, testResult1)
	assert.Equal(t, "db-name", testResult1.Name)
	assert.Equal(t, 42, testResult1.Count)

	assert.NotNil(t, testResult2)
	assert.Equal(t, false, testResult2.Enabled)
	assert.Equal(t, "db-message", testResult2.Message)
	mockDal.AssertExpectations(t)
}

func TestManager_GetAll_WithCategoryFilter(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting1 := &TestSetting1{}
	setting2 := &TestSetting2{}
	manager.Register(setting1)
	manager.Register(setting2)

	dbSettings := &[]dal.Setting{}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbSettings, nil).Once()

	results, err := manager.GetAll(ctx, testTenantID, "test")

	assert.NoError(t, err)
	assert.Len(t, results, 1)

	testResult := results[0].(*TestSetting1)
	assert.Equal(t, "default name", testResult.Name)
	assert.Equal(t, 1, testResult.Count)
	mockDal.AssertExpectations(t)
}

func TestManager_GetAll_WithInvalidJSON_UsesDefaults(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting1 := &TestSetting1{}
	manager.Register(setting1)

	dbSettings := &[]dal.Setting{
		{
			Key:               "test-setting1",
			Value:             []byte("invalid json"),
			TenantID:          testTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: true,
		},
	}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbSettings, nil).Once()

	results, err := manager.GetAll(ctx, testTenantID, "")

	assert.NoError(t, err)
	assert.Len(t, results, 1)

	testResult := results[0].(*TestSetting1)
	assert.Equal(t, "default name", testResult.Name)
	assert.Equal(t, 1, testResult.Count)
	mockDal.AssertExpectations(t)
}

func TestManager_Update_UnregisteredSetting(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	value := TestSetting1{Name: "test", Count: 1}
	err := manager.Update(ctx, testTenantID, "nonexistent-key", true, value)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting is not registered")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for unregistered setting")
}

func TestManager_Update_ValidationFailure(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting := &TestSetting1{}
	manager.Register(setting)

	value := TestSetting1{Name: "", Count: -1}
	err := manager.Update(ctx, testTenantID, "test-setting1", true, value)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for validation failure")
}

func TestManager_Update_DatabaseSaveError(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting := &TestSetting1{}
	manager.Register(setting)

	value := TestSetting1{Name: "test", Count: 1}
	valueBytes, _ := json.Marshal(value)

	mockDal.On("SaveSetting", "test-setting1", testTenantID, true, valueBytes).Return(errors.New("database save failed")).Once()

	err := manager.Update(ctx, testTenantID, "test-setting1", true, value)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Contains(t, err.Error(), "database save failed")
	assert.True(t, isInternalServerError(err), "expected ErrInternal for database error")
	mockDal.AssertExpectations(t)
}

func TestManager_Update_CategoryGetAllReturnsFreshValue(t *testing.T) {
	cachetesting.EnableCache()
	defer cachetesting.DisableCache()

	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, &CacheConfig{Enabled: true})
	manager.Register(&TestSetting1{})
	ctx := context.Background()

	oldValueBytes, _ := json.Marshal(TestSetting1{Name: "old", Count: 1})
	newValue := TestSetting1{Name: "new", Count: 2}
	newValueBytes, _ := json.Marshal(newValue)

	dbBeforeUpdate := &[]dal.Setting{{Key: "test-setting1", Value: oldValueBytes}}
	dbAfterUpdate := &[]dal.Setting{{Key: "test-setting1", Value: newValueBytes}}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbBeforeUpdate, nil).Once()
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbAfterUpdate, nil).Once()
	mockDal.On("SaveSetting", "test-setting1", testTenantID, true, newValueBytes).Return(nil).Once()

	before, err := manager.GetAll(ctx, testTenantID, "test")
	assert.NoError(t, err)
	assert.Equal(t, "old", before[0].(*TestSetting1).Name)

	err = manager.Update(ctx, testTenantID, "test-setting1", true, newValue)
	assert.NoError(t, err)

	after, err := manager.GetAll(ctx, testTenantID, "test")
	assert.NoError(t, err)
	assert.Equal(t, "new", after[0].(*TestSetting1).Name)
}

func TestManager_GetAll_CachesOnlyRequestedCategory(t *testing.T) {
	cachetesting.EnableCache()
	defer cachetesting.DisableCache()

	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, &CacheConfig{Enabled: true})
	manager.Register(&TestSetting1{})
	manager.Register(&TestSetting2{})
	ctx := context.Background()

	setting1Bytes, _ := json.Marshal(TestSetting1{Name: "a", Count: 1})
	setting2Bytes, _ := json.Marshal(TestSetting2{Enabled: true, Message: "b"})
	dbSettings := &[]dal.Setting{
		{Key: "test-setting1", Value: setting1Bytes},
		{Key: "test-setting2", Value: setting2Bytes},
	}
	mockDal.On("GetSettingsByTenantId", testTenantID).Return(dbSettings, nil).Once()

	settings, err := manager.GetAll(ctx, testTenantID, "test")
	assert.NoError(t, err)
	assert.Len(t, settings, 1)
	assert.Equal(t, "test-setting1", settings[0].(*TestSetting1).Key())

	cached := getCachedSettingsList(ctx, testTenantID, "test")
	assert.NotNil(t, cached)
	assert.Len(t, *cached, 1)
	assert.Equal(t, "test-setting1", (*cached)[0].Key)
}

func TestManager_Update_Success_UserUpdate(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting := &TestSetting1{}
	manager.Register(setting)

	value := TestSetting1{Name: "test", Count: 42}
	valueBytes, _ := json.Marshal(value)

	mockDal.On("SaveSetting", "test-setting1", testTenantID, true, valueBytes).Return(nil).Once()

	err := manager.Update(ctx, testTenantID, "test-setting1", true, value)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_Update_Success_SystemUpdate(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	ctx := context.Background()

	setting := &TestSetting2{}
	manager.Register(setting)

	value := TestSetting2{Enabled: false, Message: "system update"}
	valueBytes, _ := json.Marshal(value)

	mockDal.On("SaveSetting", "test-setting2", testTenantID, false, valueBytes).Return(nil).Once()

	err := manager.Update(ctx, testTenantID, "test-setting2", false, value)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_createNewInstance(t *testing.T) {
	manager := NewManager(nil, nil)
	setting := &TestSetting1{Name: "original", Count: 100}

	newInstance := manager.createNewInstance(setting)

	assert.NotNil(t, newInstance)
	assert.NotEqual(t, setting, newInstance)
	assert.IsType(t, &TestSetting1{}, newInstance)
	testResult := newInstance.(*TestSetting1)
	assert.Equal(t, "", testResult.Name)
	assert.Equal(t, 0, testResult.Count)
}

func TestManager_convertSettingsSliceToMap_EmptySlice(t *testing.T) {
	manager := NewManager(nil, nil)

	result := manager.convertSettingsSliceToMap(nil)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)

	emptySlice := &[]dal.Setting{}
	result = manager.convertSettingsSliceToMap(emptySlice)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestManager_convertSettingsSliceToMap_WithSettings(t *testing.T) {
	manager := NewManager(nil, nil)

	settings := &[]dal.Setting{
		{Key: "key1", Value: []byte("value1")},
		{Key: "key2", Value: []byte("value2")},
	}

	result := manager.convertSettingsSliceToMap(settings)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "key1")
	assert.Contains(t, result, "key2")
	assert.Equal(t, []byte("value1"), result["key1"].Value)
	assert.Equal(t, []byte("value2"), result["key2"].Value)
}

func TestManager_ShouldMigrateSetting_ReturnsTrue(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	mockDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()

	result, err := manager.ShouldMigrateSetting("test-setting1")

	assert.NoError(t, err)
	assert.True(t, result)
	mockDal.AssertExpectations(t)
}

func TestManager_ShouldMigrateSetting_ReturnsFalse(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	mockDal.On("ShouldMigrateSetting", "test-setting1").Return(false, nil).Once()

	result, err := manager.ShouldMigrateSetting("test-setting1")

	assert.NoError(t, err)
	assert.False(t, result)
	mockDal.AssertExpectations(t)
}

func TestManager_ShouldMigrateSetting_DifferentKeys(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	mockDal.On("ShouldMigrateSetting", "test-setting1").Return(true, nil).Once()
	mockDal.On("ShouldMigrateSetting", "test-setting2").Return(false, nil).Once()

	result1, err1 := manager.ShouldMigrateSetting("test-setting1")
	result2, err2 := manager.ShouldMigrateSetting("test-setting2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.True(t, result1)
	assert.False(t, result2)
	mockDal.AssertExpectations(t)
}

func TestManager_ShouldMigrateSetting_ReturnsError(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	mockDal.On("ShouldMigrateSetting", "test-setting1").Return(false, errors.New("database error")).Once()

	result, err := manager.ShouldMigrateSetting("test-setting1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.False(t, result)
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_UnregisteredSetting(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "test", Count: 1}},
	}

	err := manager.UpdateAllTenantsByKey("nonexistent-key", tenantValues)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting is not registered")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for unregistered setting")
}

func TestManager_UpdateAllTenantsByKey_NoMigrationTransformation(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)
	manager.Register(&TestSetting2{})
	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting2{Enabled: false, Message: "test"}},
	}

	err := manager.UpdateAllTenantsByKey("test-setting2", tenantValues)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting does not support migration transformation")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for unregistered setting")
}

func TestManager_UpdateAllTenantsByKey_ValidationFailure(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "", Count: -1}},
	}

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for validation failure")
}

func TestManager_UpdateAllTenantsByKey_DatabaseSaveError(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "test", Count: 1}},
	}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.AnythingOfType("[]database.TenantSettingInput")).Return(errors.New("database batch save failed")).Once()

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to batch update setting")
	assert.Contains(t, err.Error(), "database batch save failed")
	assert.True(t, isInternalServerError(err), "expected ErrInternal for database error")
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_Success_SingleTenant(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "test", Count: 42}},
	}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		if len(inputs) != 1 {
			return false
		}
		return inputs[0].TenantID == "1"
	})).Return(nil).Once()

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_Success_MultipleTenants(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "test1", Count: 1}},
		{TenantId: 2, Value: TestSetting1{Name: "test2", Count: 2}},
		{TenantId: 3, Value: TestSetting1{Name: "test3", Count: 3}},
	}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		return len(inputs) == 3
	})).Return(nil).Once()

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_Success_VerifyTenantInputs(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	value1 := TestSetting1{Name: "test1", Count: 10}
	value2 := TestSetting1{Name: "test2", Count: 20}
	expectedBytes1, _ := json.Marshal(value1)
	expectedBytes2, _ := json.Marshal(value2)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: value1},
		{TenantId: 2, Value: value2},
	}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		if len(inputs) != 2 {
			return false
		}

		foundTenant1 := false
		foundTenant2 := false
		for _, input := range inputs {
			if input.TenantID == "1" && string(input.Value) == string(expectedBytes1) {
				foundTenant1 = true
			}
			if input.TenantID == "2" && string(input.Value) == string(expectedBytes2) {
				foundTenant2 = true
			}
		}
		return foundTenant1 && foundTenant2
	})).Return(nil).Once()

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_EmptyTenantValues(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting1", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		return len(inputs) == 0
	})).Return(nil).Once()

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}

func TestManager_UpdateAllTenantsByKey_PartialValidationFailure(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSetting1{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSetting1{Name: "valid", Count: 1}},
		{TenantId: 2, Value: TestSetting1{Name: "", Count: -1}},
	}

	err := manager.UpdateAllTenantsByKey("test-setting1", tenantValues)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid")
	assert.True(t, isBadRequest(err), "expected ErrBadRequest for validation failure")
	mockDal.AssertNotCalled(t, "SaveSettingByKeyForTenantsList")
}

func TestManager_UpdateAllTenantsByKey_WithTransformation(t *testing.T) {
	mockDal := &mockManager.SettingsDal{}
	var dalInterface SettingsDal = mockDal
	manager := NewManager(dalInterface, nil)

	setting := &TestSettingWithTransform{}
	manager.Register(setting)

	tenantValues := []TenantSettingValueForBackendResponse{
		{TenantId: 1, Value: TestSettingWithTransform{Value: "original"}},
	}

	mockDal.On("SaveSettingByKeyForTenantsList", "test-setting-transform", mock.MatchedBy(func(inputs []dal.TenantSettingInput) bool {
		if len(inputs) != 1 {
			return false
		}
		var result TestSettingWithTransform
		if err := json.Unmarshal(inputs[0].Value, &result); err != nil {
			return false
		}
		return result.Value == "transformed-original"
	})).Return(nil).Once()

	err := manager.UpdateAllTenantsByKey("test-setting-transform", tenantValues)

	assert.NoError(t, err)
	mockDal.AssertExpectations(t)
}
