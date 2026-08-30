package setting

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	errHandler "github.com/run-ai/runai/runai-common-packages/setting/error"
	manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type SettingsManager interface {
	Register(setting manager.Settinger) error
	Get(ctx context.Context, tenantId string, settingKey string) (manager.Settinger, error)
	GetAll(ctx context.Context, tenantId, category string) ([]manager.Settinger, error)
	Update(ctx context.Context, tenantId string, settingKey string, updateByCustomer bool, value interface{}) error
	ShouldMigrateSetting(key string) (bool, error)
	UpdateAllTenantsByKey(settingKey string, tenantIdToSettingValue []manager.TenantSettingValueForBackendResponse) error
}

type BackendClient interface {
	GetSettingByKeyForAllTenants(ctx context.Context, key string) (*manager.SettingValuesForAllTenantsBackendResponse, error)
}

var (
	once            sync.Once
	settingsManager SettingsManager
	backendClient   BackendClient
)

type SettingConfig struct {
	CacheEnabled           bool
	CacheExpirationMinutes *int
	BackendClient          BackendClient
}

func Init(settingConfig *SettingConfig, settingsDal manager.SettingsDal) {
	once.Do(func() {
		var cacheConfig *manager.CacheConfig
		if settingConfig != nil {
			cacheConfig = &manager.CacheConfig{
				Enabled:           settingConfig.CacheEnabled,
				ExpirationMinutes: settingConfig.CacheExpirationMinutes,
			}
			if settingConfig.BackendClient != nil {
				backendClient = settingConfig.BackendClient
			}
		}
		settingsManager = manager.NewManager(settingsDal, cacheConfig)
	})
}

func initializSetting[T manager.Settinger]() T {
	var c T
	cType := reflect.TypeOf(c)
	if cType.Kind() != reflect.Ptr {
		return c
	}
	elemType := cType.Elem()
	newValue := reflect.New(elemType).Interface().(T)
	return newValue
}

func Register[T manager.Settinger]() error {
	if err := checkInitialized(); err != nil {
		return err
	}
	setting := initializSetting[T]()
	return settingsManager.Register(setting)
}

func Validate[T manager.Settinger](settingValue manager.Settinger) bool {
	setting := initializSetting[T]()
	return setting.Validate(settingValue)
}

func Get[T manager.Settinger](tenantId string) (T, error) {
	var zeroValue T
	if err := checkInitialized(); err != nil {
		return zeroValue, err
	}
	ctx := context.Background()
	setting := initializSetting[T]()
	settingKey := setting.Key()

	s, err := settingsManager.Get(ctx, tenantId, settingKey)
	if err != nil {
		return zeroValue, err
	}
	typedSetting, ok := s.(T)
	if !ok {
		return zeroValue, errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to get setting by key: '%s'. type error", settingKey),
		}
	}
	return typedSetting, nil
}

func GetAll(tenantId, category string) ([]manager.Settinger, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	return settingsManager.GetAll(ctx, tenantId, category)
}

func Update[T manager.Settinger](tenantId string, value interface{}) *updateBuilder[T] {
	return &updateBuilder[T]{
		tenantId:     tenantId,
		value:        value,
		updateByUser: true,
	}
}

func Migrate[T manager.Settinger]() error {
	setting := initializSetting[T]()
	settingKey := setting.Key()

	shouldMigrate, err := settingsManager.ShouldMigrateSetting(settingKey)
	if err != nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to migrate setting by key '%s'. err: %v", settingKey, err),
		}
	}
	if !shouldMigrate {
		return nil
	}

	if backendClient == nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to migrate setting by key: '%s'. backend client not configured", settingKey),
		}
	}

	tenantIdToSettingValue, err := backendClient.GetSettingByKeyForAllTenants(context.Background(), settingKey)
	if err != nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to migrate setting by key: '%s'. err: %v", settingKey, err),
		}
	}

	if tenantIdToSettingValue == nil || len(tenantIdToSettingValue.Values) == 0 {
		return nil
	}

	return settingsManager.UpdateAllTenantsByKey(settingKey, tenantIdToSettingValue.Values)
}

func SetMock(mockManager SettingsManager) {
	settingsManager = mockManager
}

func checkInitialized() error {
	if settingsManager == nil {
		return errHandler.ErrorInternal{
			Message: "setting package is not initialized",
		}
	}
	return nil
}
