# Settings Package
A flexible, type-safe settings management system for multi-tenant applications with database persistence and caching support.

## Overview
The settings package provides a generic, type-safe way to manage application settings across multiple tenants. It supports:

- **Type-safe setting definitions** using Go generics
- **Multi-tenant isolation** with tenant-specific configurations
- **Database persistence** with pluggable DAL interface
- **Caching support** for improved performance
- **Builder pattern for updates** with clear user/system distinction
- **Validation and defaults** per setting type
- **Lifecycle management** with staging (Alpha, Beta, Legacy)


#### Interfaces
- **`Settinger`**: Core interface that all settings must implement
- **`SettingsDal`**: Database access layer interface for pluggable persistence

#### Types
- **`Setting`**: Database entity representing a persisted setting
- **`SettingConfig`**: Configuration structure for initialization
- **`UpdateBuilder[T]`**: Builder pattern for setting updates
- **`Stage`**: Enumeration for setting lifecycle stages
- **`SettingMetadata`**: Version compatibility information

## Quick Start

### 1. Define a Setting

```go
package main

import "github.com/run-ai/runai/runai-common-packages/setting"

type MyAppSetting struct {
    MaxConnections int    `json:"max_connections"`
    EnableFeatureX bool   `json:"enable_feature_x"`
    ApiTimeout     int    `json:"api_timeout"`
}

// Implement the Settinger interface
func (s MyAppSetting) Key() string {
    return "my_app_config"
}

func (s MyAppSetting) SetDefault() interface{} {
    return MyAppSetting{
        MaxConnections: 100,
        EnableFeatureX: false,
        ApiTimeout:     30,
    }
}

func (s MyAppSetting) DisplayName() string {
    return "My Application Settings"
}

func (s MyAppSetting) Description() string {
    return "Core configuration settings for My Application"
}

func (s MyAppSetting) Category() string {
    return "application"
}

func (s MyAppSetting) Stage() setting.Stage {
    return setting.Beta
}

func (s MyAppSetting) ValidateValue(value interface{}) bool {
    if cfg, ok := value.(MyAppSetting); ok {
        return cfg.MaxConnections > 0 && cfg.ApiTimeout > 0
    }
    return false
}
```

### 2. Implement the DAL Interface

```go
type MySettingsDal struct {
    db *sql.DB
}

func (dal *MySettingsDal) GetSettingByKey(key, tenantId string) (*setting.SettingsDao, error) {
    // Implement database retrieval logic
    var settingsDao setting.SettingsDao
    err := dal.db.QueryRow(
        "SELECT key, value, tenant_id, updated_at, updated_by_customer FROM settings WHERE key = ? AND tenant_id = ?",
        key, tenantId,
    ).Scan(&settingsDao.Key, &settingsDao.Value, &settingsDao.TenantID, &settingsDao.UpdatedAt, &settingsDao.UpdatedByCustomer)

    if err == sql.ErrNoRows {
        return nil, nil // Setting not found
    }
    return &settingsDao, err
}

func (dal *MySettingsDal) GetSettingsByTenantId(tenantId string) (map[string]*setting.SettingsDao, error) {
    // Implement retrieval of all settings for a tenant
    rows, err := dal.db.Query(
        "SELECT key, value, tenant_id, updated_at, updated_by_customer FROM settings WHERE tenant_id = ?",
        tenantId,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[string]*setting.SettingsDao)
    for rows.Next() {
        var settingsDao setting.SettingsDao
        err := rows.Scan(&settingsDao.Key, &settingsDao.Value, &settingsDao.TenantID, &settingsDao.UpdatedAt, &settingsDao.UpdatedByCustomer)
        if err != nil {
            return nil, err
        }
        result[settingsDao.Key] = &settingsDao
    }
    return result, nil
}

func (dal *MySettingsDal) SaveSetting(key, tenantId string, updateByUser bool, value []byte) error {
    // Implement database save logic
    _, err := dal.db.Exec(
        "INSERT OR REPLACE INTO settings (key, tenant_id, value, updated_by_customer, updated_at) VALUES (?, ?, ?, ?, ?)",
        key, tenantId, value, updateByUser, time.Now(),
    )
    return err
}
```

### 3. Initialize and Register Settings

> **Prerequisite:** the `runaidb` package must be initialized before using this
> package, since the DAL runs its queries against `runaidb.DB()`. Call
> `runaidb.Init(nil)` (nil uses the default/env-based database config) once at
> startup, before `setting.Init`.

```go
func main() {
    // Initialize the database layer first (required by the DAL).
    if err := runaidb.Init(nil); err != nil {
        log.Fatal("Failed to init runaidb:", err)
    }

    // Create your DAL implementation
    dal := &MySettingsDal{}

    // Configure the settings package
    cacheExpirationMinutes := 10
    config := &setting.SettingConfig{
        CacheEnabled:           true,
        CacheExpirationMinutes: &cacheExpirationMinutes,
        BackendClient:          myBackendClient, // optional
    }

    // Initialize the package (only called once); the DAL is passed separately.
    // The database connection is provided via the runaidb package.
    setting.Init(config, dal)

    // Register your settings
    err := setting.Register[MyAppSetting]()
    if err != nil {
        log.Fatal("Failed to register setting:", err)
    }
}
```

### 4. Use the Settings

```go
// Get a setting for a specific tenant
tenantId := "tenant-123"
appConfig, err := setting.Get[MyAppSetting](tenantId)
if err != nil {
    log.Fatal("Failed to get setting:", err)
}

fmt.Printf("Max connections: %d\n", appConfig.MaxConnections)
fmt.Printf("Feature X enabled: %t\n", appConfig.EnableFeatureX)

// Update a setting
newConfig := MyAppSetting{
    MaxConnections: 200,
    EnableFeatureX: true,
    ApiTimeout:     45,
}

// Update by user (default behavior):
err = setting.Update[MyAppSetting](tenantId, newConfig).Execute()
if err != nil {
    log.Fatal("Failed to update setting:", err)
}

//Or explicitly specify update type:
err = setting.Update[MyAppSetting](tenantId, newConfig).UpdatedByUser().Execute()
err = setting.Update[MyAppSetting](tenantId, newConfig).UpdatedBySystem().Execute()

// Get all settings in a category
allAppSettings, err := setting.GetAll[Settinger](tenantId, "application")
if err != nil {
    log.Fatal("Failed to get settings:", err)
}
```

## API Reference

### Core Functions

#### `Init(settingConfig *SettingConfig)`
Initializes the settings package with database, cache, and DAL configuration. Must be called once before using any other functions.

#### `Register[T Settinger]() error`
Registers a setting type with the package. Must be called for each setting type you want to use.

#### `Get[T Settinger](tenantId string) (*T, error)`
Retrieves a setting for a specific tenant. Returns default values if not set in database.

#### `GetAll[T Settinger](tenantId, category string) ([]T, error)`
Retrieves all settings of a specific type for a tenant, optionally filtered by category.

#### `Update[T Settinger](tenantId string, value interface{}) *UpdateBuilder[T]`
Creates an update builder for modifying a setting value. **Defaults to user update** (UpdatedByCustomer = true).

### Update Builder Methods

#### `UpdatedByUser() *UpdateBuilder[T]`
Explicitly marks the update as performed by a user (sets `UpdatedByCustomer` to true). This is the default behavior.

#### `UpdatedBySystem() *UpdateBuilder[T]`
Marks the update as performed by the system (sets `UpdatedByCustomer` to false).

#### `Execute() error`
Validates and persists the setting update to the database.

### Update Behavior

The UpdateBuilder pattern provides a fluent interface for setting updates with clear semantics:

```go
// Default behavior - User update (UpdatedByCustomer = true)
setting.Update[MySetting](tenantId, newValue).Execute()

// Explicit user update (same as default)
setting.Update[MySetting](tenantId, newValue).UpdatedByUser().Execute()

// System update (UpdatedByCustomer = false)
setting.Update[MySetting](tenantId, newValue).UpdatedBySystem().Execute()
```

**Important**: The default behavior is **user update** (`UpdatedByCustomer = true`), meaning updates are attributed to user actions unless explicitly marked as system updates.

## Settinger Interface Requirements
All settings must implement the `Settinger` interface:
```go
type Settinger interface {
    Key() string                    // Unique identifier for the setting
    SetDefault() interface{}        // Default values when not set
    DisplayName() string           // Human-readable name
    Description() string           // Description of the setting's purpose
    Category() string              // Category for organization
    Stage() Stage                  // Lifecycle stage (Alpha, Beta, Legacy)
    ValidateValue(interface{}) bool // Validation logic
}
```


## Example Test
```go
func TestMyAppSetting(t *testing.T) {
    // Create a mock DAL
    mockDal := &MockSettingsDal{}

    config := &setting.SettingConfig{
        CacheEnabled: false,
        // ... other config
    }

    setting.Init(config, mockDal)
    err := setting.Register[MyAppSetting]()
    assert.NoError(t, err)

    // Test getting default values
    config, err := setting.Get[MyAppSetting]("test-tenant")
    assert.NoError(t, err)
    assert.Equal(t, 100, config.MaxConnections)

    // Test updating (defaults to user update)
    newConfig := MyAppSetting{MaxConnections: 200}
    err = setting.Update[MyAppSetting]("test-tenant", newConfig).Execute()
    assert.NoError(t, err)

    // Test system update
    err = setting.Update[MyAppSetting]("test-tenant", newConfig).
        UpdatedBySystem().
        Execute()
    assert.NoError(t, err)
}
```

## Dependencies
- `github.com/run-ai/runai/runai-common-packages/cache/v2` - Caching support
- `github.com/run-ai/runai/runai-common-packages/runaidb` - Database access layer


## License
This package is part of the RunAI common packages and follows the same licensing terms.
