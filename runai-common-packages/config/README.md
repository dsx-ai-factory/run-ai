# `config` Package Documentation

The `config` package provides a streamlined way to manage application configurations using a generic, type-safe, and centralized registry. It builds upon the capabilities of the `viper` library, offering automatic registration, retrieval, and updates of configuration objects.

---

## Features

- **Generics-Based API**: Simplified and type-safe configuration retrieval and registration using Go generics.
- **Automatic Defaults**: Supports setting and managing default values for configurations.
- **Centralized Registry**: Manages all configurations in a single thread-safe registry.
- **File and Environment Support**: Load configurations from files or environment variables.
- **Secret Fields Support**: Load secrets from a dedicated config file.
- **Dynamic Registration**: Automatically registers and retrieves configurations if not already registered.

---

## Environment variables

The `config` package reads several environment variables to control where configuration and secret data are loaded from. If an environment variable is not set, the package will use the default value shown below.

- **`CONFIG_KEY_DELIMITER`**
	- Purpose: Delimiter used to convert nested config keys into environment variable names. Example default delimiter: `.` — a key `db.connection.host` maps to env `DB_CONNECTION_HOST`.
	- Default: `.`
	- Required: No
	- Used by: [Setting a Custom Key Delimiter for Config Files](#setting-a-custom-key-delimiter-for-config-files---not-required)

- **`CONFIG_FILE_PATH`**
	- Purpose: Path to a single configuration file to load (appended to the list of config files). Useful to override or add a single config file path.
	- Default: (none)
	- Required: No
	- Used by: [Setting Configuration Files](#setting-configuration-files---not-required)

- **`CONFIG_SECRETS_FILE_PATH`**
	- Purpose: Path to a single secrets configuration file. When set, the file is used to load secret values and is watched for changes.
	- Default: (none)
	- Required: No
	- Used by: [Setting a Secret Configuration File](#setting-a-secret-configuration-file---not-required)

- **`SECRETS_MOUNT_PATH`**
	- Purpose: Filesystem path containing secret files (e.g., Kubernetes-mounted secrets). Files inside this folder are read and exposed as secret keys.
	- Default: `/etc/secrets`
	- Required: No
	- Used by: [Setting a K8S Mounted Secrets Folder Path](#setting-a-k8s-mounted-secrets-folder-path---not-required)


## Usage

### Initializing the Registry with Init - (Not Required)

The `config.Init()` function allows you to initialize the configuration manually before using the package. You can pass a *Config struct with custom values.

```go
import "config"

conf := &config.Config{
    KeyDelimiter: ";",
}
err := config.Init(conf)
if err != nil {
	fmt.Printf("Error initializing config: %v\n", err)
}
```

If the passed config is nil, no changes are applied and the function exits early.

### Setting a Custom Key Delimiter for Config Files - (Not Required)

By default, the delimiter for configuration keys is "."
To use a different delimiter, set the environment variable `CONFIG_KEY_DELIMITER` to your preferred delimiter before starting the application.

### Setting Configuration Files - (Not required)

Before accessing or registering configurations, specify one or more configuration files:

```go
import "config"

err := config.SetConfigFiles("path/to/config1.yaml", "path/to/config2.yaml")
if err != nil {
    fmt.Printf("Error loading config files: %v\n", err)
}

err := config.SetConfigFiles() // Load the path from an environment variable (CONFIG_FILE_PATH)
if err != nil {
    fmt.Printf("Error loading config file: %v\n", err)
}
```

This reads the given configuration files and makes them available through the registry.
If no params are given, ensure that the corresponding environment variable `CONFIG_FILE_PATH` is set correctly in your system to allow automatic loading of the path.

The system also watches for changes in the configuration files and automatically reloads them.

If no configuration file is provided all values will load from ENV vars

### Setting a Secret Configuration File - (Not required)

Before accessing or registering configurations, specify a secret configuration file:

```go
import "config"

err := config.SetSecretConfigFile("path/to/secret_config.yaml")
if err != nil {
    fmt.Printf("Error loading config file: %v\n", err)
}

err := config.SetSecretConfigFile("") // Load the path from the environment variable (CONFIG_SECRETS_FILE_PATH)
if err != nil {
    fmt.Printf("Error loading config file: %v\n", err)
}
```

This reads the configuration file and makes its values available through the registry for fields tagged with `config:"secret"`.
If the string is empty, ensure that the corresponding environment variable `CONFIG_SECRETS_FILE_PATH` is set correctly in your system to allow automatic loading of the path.

See [Secret Field](#secret-fields)

### Setting a K8S Mounted Secrets Folder Path - (Not required)

The `config` package supports loading secrets from files mounted as volumes in Kubernetes containers. To specify the folder containing these secret files, use the `SetKbSecretsFolder` method:

```go
import "config"

err := config.SetKbSecretsFolder("path/to/secrets/folder")
if err != nil {
    fmt.Printf("Error loading secret files: %v\n", err)
}

err := config.SetKbSecretsFolder("") // Load the path from an environment variable (SECRETS_MOUNT_PATH)
if err != nil {
    fmt.Printf("Error loading secret files: %v\n", err)
}
```

This method reads the content of each file in the specified folder and associates the file name with its corresponding configuration key. For example, if a file named `database.host` contains the value `1.1.1.1`, it will be mapped to the configuration key `database.host`.
The folder is watched for direct file changes and Kubernetes atomic secret-volume updates. When a secret changes, registered configurations are reloaded and the `OnConfigChange` callback is invoked.

If the string is empty, ensure that the corresponding environment variable `SECRETS_MOUNT_PATH` is set correctly in your system to allow automatic loading of the path.
If none is provided the fallback path is `/etc/secrets`

> [!NOTE]
> Fields tagged with `config:"secret"` will exclusively use these values for their configuration.

#### Example

Given the following file structure:

```
path/to/secrets/folder/
├── database.host
└── database.password
```

Contents of `database.host`:

```
1.1.1.1
```

Contents of `database.password`:

```
securepassword123
```

And the following configuration struct:

```go
type Database struct {
    Host     string `mapstructure:"host" config:"secret"`
    Password string `mapstructure:"password" config:"secret"`
}
```

The secret values will be automatically loaded into the fields tagged with `config:"secret"`.

### Environment Variable Overrides

Environment variables can override configuration values specified in the file. The `config` package automatically loads environment variables whose names match the `mapstructure` tags in your configuration struct. For example:

```go
type AppConfig struct {
    Name string      `mapstructure:"name"`
    Port int         `mapstructure:"port"`
}
```

If the following environment variables are set:

- `NAME=EnvApp`
- `PORT=8081`

These values will override the corresponding fields in the configuration file.

> [!IMPORTANT] 
> Secret fields are set only using the secret config file or kb secrets folder

### Secret Fields

When tagging a field with the `config:"secret"` tag, the value will be loaded from the **secrets config file only**. For example:

```go
type AppConfig struct {
    Name     string      `mapstructure:"name"`
    Port     int         `mapstructure:"port"`
    Password string      `mapstructure:"password" config:"secret"`
}
```

If the secrets config file contains:
```yaml
password: 123
```
The filed will be set.

### Defining a Configuration

Define a configuration struct that implements the `Configer` interface by providing a `SetDefaults` method:

```go
type AppConfig struct {
    Name     string      `mapstructure:"name"`
    Port     int         `mapstructure:"port"`
    Password string      `mapstructure:"password" config:"secret"`
}

func (a *AppConfig) SetDefaults() {
    a.Name = "DefaultApp"
    a.Port = 8080
}
```

> **Unexported fields are not supported.**
> Every field on a `Configer` — including nested structs — must be exported. `Register` and `RegisterRoot` do not support unexported fields: they are never loaded from files or environment variables and are lost on the registered config. Export anything the service needs to read or keep.

### Registering a Configuration

Register a configuration using generics:

```go
err := config.Register[*AppConfig]()
if err != nil {
    fmt.Printf("Error registering config: %v\n", err)
}
```

### Retrieving a Configuration

Retrieve a registered configuration:

```go
appConfig := config.Get[*AppConfig]()
fmt.Printf("App Name: %s, Port: %d\n", appConfig.Name, appConfig.Port)
```

### Automatically Registering and Retrieving

Use `GetR` to retrieve a configuration and register it if not already registered:

```go
appConfig, err := config.GetR[*AppConfig]()
if err != nil {
    fmt.Printf("Error registering config: %v\n", err)
}
fmt.Printf("App Name: %s, Port: %d\n", appConfig.Name, appConfig.Port)
```

### Setting a Configuration

The `Set` method allows you to manually set an already initialized configuration struct in the registry. This can be useful for testing purposes or explicitly overriding configurations.

```go
appConfig := AppConfig{
    Name: "TestApp",
    Port: 3000,
}

err := config.Set(&appConfig)
if err != nil {
    fmt.Printf("Error setting config: %v\n", err)
}
```

### Setting Overrides

`SetOverrides` pins specific field values so they take priority over config-file and environment-variable values. Overrides survive automatic config-file reloads but are cleared when `Set` or `SetWithDefaults` is called for the same type.

```go
err := config.SetOverrides(&AppConfig{
    Name: "ForcedApp",
})
if err != nil {
    fmt.Printf("Error setting overrides: %v\n", err)
}
```

After this call, `config.Get[*AppConfig]().Name` returns `"ForcedApp"` even if the config file or an env var sets a different value. If the config file is then hot-reloaded, the override is re-applied automatically.

Calling `config.Set(...)` or `config.SetWithDefaults(...)` replaces the item entirely and clears any stored overrides.

> **Limitation — zero values cannot be overridden.**
> `SetOverrides` only applies fields whose value is non-zero in the struct you pass in. The zero value for each Go type (`false` for `bool`, `0` for numeric types, `""` for `string`, `nil` for pointers/maps/slices) is treated as "not set" and is silently ignored. This means you **cannot** use `SetOverrides` to force a field to `false` or `0`:
>
> ```go
> // This does NOT override Enabled to false — false is the zero value and is skipped.
> config.SetOverrides(&AppConfig{Enabled: false})
>
> // This does NOT override Port to 0 — same reason.
> config.SetOverrides(&AppConfig{Port: 0})
> ```
>
> If you need to force a field to its zero value, use `config.Set(...)` or `config.SetWithDefaults(...)` to replace the entire config instead.

### Flushing All Configurations

Clear all registered configurations:

```go
config.FlushAll()
```

### On Config Change

Use to register a callback function when a configuration change event is triggered

```go
config.OnConfigChange(func(){
	fmt.Print("configuration changed")
})
```

## Registering a Root Configuration

In cases where your configuration file uses a nested structure with the name of the struct as the root key, you can use `RegisterRoot` to register and load the configuration. For instance, given the following struct:

```go
type Database struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}
```

The corresponding configuration file might look like this:

```yaml
database:
  host: my-host
  port: 999
```

To register this structure, use the `RegisterRoot` method:

```go
err := config.RegisterRoot[*Database]()
if err != nil {
    fmt.Printf("Error registering root config: %v\n", err)
}
```

This approach ensures that the configuration values under the `database` root key are correctly mapped to the `Database` struct.

#### Important Note about RegisterRoot
It will automatically trim the struct name suffix in case it ends in: Config|Cfg|Conf|Configuration.
For example:
```go
type DatabaseConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}
```
Will be handled exactly as the Database example shown above.

### Environment Variable Overrides Root
Environment variables can override configuration values specified in the file. The config package automatically loads environment variables whose names match the mapstructure tags in your configuration struct and prefixed by the struct name. For example:

```go
type App struct {
    Name string      `mapstructure:"name"`
    Port int         `mapstructure:"port"`
}
```
If the following environment variables are set:

```yaml
APP_NAME=EnvApp
APP_PORT=8081
These values will override the corresponding fields in the configuration file.
```

### Validation Support
Both `Register` and `RegisterRoot` now validate your configuration structs using `github.com/go-playground/validator/v10`. If your struct includes validation tags (e.g., validate:"required"), they will be enforced during registration. If validation fails, the `Register`/ `RegisterRoot` methods will return an error.

### Retrieving a Root Configuration

Once registered, you can retrieve the configuration as usual using the `Get` method:

```go
databaseConfig := config.Get[*Database]()
fmt.Printf("Host: %s, Port: %d\n", databaseConfig.Host, databaseConfig.Port)
```

### Automatically Registering and Retrieving a Root Configuration

For convenience, the `GetRootR` method can be used to automatically register and retrieve a root configuration if it has not already been registered:

```go
databaseConfig, err := config.GetRootR[*Database]()
if err != nil {
    fmt.Printf("Error registering root config: %v\n", err)
}
fmt.Printf("Host: %s, Port: %d\n", databaseConfig.Host, databaseConfig.Port)
```

This simplifies cases where the root structure of the configuration file aligns with the struct's definition.


## Example

Here’s a complete example of using the `config` package:

```go
package main

import (
	"fmt"
	"config"
)

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

func (a *AppConfig) SetDefaults() {
	a.Name = "DefaultApp"
	a.Port = 8080
}

func main() {
	// Set configuration files
	err := config.SetConfigFiles("config.yaml")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Automatically register and retrieve the configuration
	appConfig, err := config.GetR[*AppConfig]()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Access the configuration
	fmt.Printf("App Name: %s, Port: %d\n", appConfig.Name, appConfig.Port)

	// Set a custom configuration (useful for tests)
	testConfig := AppConfig{
		Name: "TestApp",
		Port: 3000,
	}
	err = config.Set(&testConfig)
	if err != nil {
		fmt.Printf("Error setting config: %v\n", err)
	}

	// Retrieve and verify the updated configuration
	appConfig = config.Get[*AppConfig]()
	fmt.Printf("Updated App Name: %s, Port: %d\n", appConfig.Name, appConfig.Port)
}
```
