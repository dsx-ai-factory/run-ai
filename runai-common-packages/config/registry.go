package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	envKbSecretsFolderPath   = "SECRETS_MOUNT_PATH"
	envConfigFilePath        = "CONFIG_FILE_PATH"
	envSecretsConfigFilePath = "CONFIG_SECRETS_FILE_PATH" //nolint:gosec // G101: environment variable name, not a credential
	envConfigKeyDelimiter    = "CONFIG_KEY_DELIMITER"

	defaultKeyDelimiter        = "."
	defaultKbSecretsFolderPath = "/etc/secrets"
	kbSecretsReloadDebounce    = 50 * time.Millisecond
	kbSecretsReloadNever       = 365 * 24 * time.Hour
)

var validate = validator.New()

func newRegistry() *registry {
	return newRegistryWithConfig(getConfig())
}

func newRegistryWithConfig(conf *Config) *registry {
	r := &registry{
		mu:               sync.RWMutex{},
		items:            make(map[reflect.Type]*registryItem),
		v:                initViperWithOptions(conf.KeyDelimiter),
		secretV:          viper.New(),
		callbackOnChange: func() {},
		conf:             conf,
	}
	r.confPool = newConfigPool(r.onConfigPoolChange, conf.KeyDelimiter)

	return r
}

type registry struct {
	mu               sync.RWMutex
	items            map[reflect.Type]*registryItem
	confPool         *configPool
	v                *viper.Viper
	secretV          *viper.Viper
	kbSecretsWatcher *fsnotify.Watcher
	callbackOnChange func()
	conf             *Config
}

func NewRegistryItem(item Configer, root bool) *registryItem {
	return &registryItem{
		name:           reflect.TypeOf(item).Elem().Name(),
		item:           item,
		rType:          reflect.TypeOf(item).Elem(),
		registeredRoot: root,
	}
}

func NewRegistryWithOverrides(overrides Configer) *registryItem {
	return &registryItem{
		name:      reflect.TypeOf(overrides).Elem().Name(),
		rType:     reflect.TypeOf(overrides).Elem(),
		overrides: overrides,
	}
}

type registryItem struct {
	name           string
	item           Configer
	rType          reflect.Type
	registeredRoot bool
	overrides      Configer
}

func (ri *registryItem) validate() error {
	if err := validate.Struct(ri.item); err != nil {
		return fmt.Errorf("failed to validate config struct %s, error: %w", ri.name, err)
	}
	return nil
}

func (r *registry) SetKbSecretsFolder(folderPath string) error {
	if folderPath == "" {
		folderPath = getEnvOrDefault(envKbSecretsFolderPath, defaultKbSecretsFolderPath)
	}

	if err := r.loadKbSecretsFolder(folderPath); err != nil {
		return err
	}

	watcher, err := r.initK8SWatcher(folderPath)
	if err != nil {
		return nil
	}

	go r.watchKbSecretsFolder(watcher, folderPath)

	return nil
}

func (r *registry) SetConfigFiles(paths ...string) error {
	envPath := os.Getenv(envConfigFilePath)
	if envPath != "" {
		paths = append(paths, envPath)
	}
	if len(paths) == 0 {
		return fmt.Errorf("no config file paths provided")
	}

	v, err := r.confPool.AddConfigFiles(paths)
	if err != nil {
		return err
	}
	r.v = v
	return nil
}

func (r *registry) SetSecretConfigFile(path string) error {
	if path == "" {
		path = os.Getenv(envSecretsConfigFilePath)
	}

	r.secretV.SetConfigFile(path)
	r.secretV.WatchConfig()
	r.secretV.OnConfigChange(func(in fsnotify.Event) {
		r.reloadRegistry()
	})

	err := r.secretV.ReadInConfig()
	if err == nil {
		return nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return fmt.Errorf("config file %s was not found", path)
	}

	return fmt.Errorf("error while reading config file: %e", err)
}

// Method to register a config in the registry
func (r *registry) Register(c Configer) error {
	c.SetDefaults()
	err := r.v.Unmarshal(c)
	if err != nil {
		return err
	}

	removeSecretValuesRecursive(c)
	err = r.secretV.Unmarshal(c)
	if err != nil {
		return err
	}

	item := NewRegistryItem(c, false)

	r.restoreOverrides(item)

	if err = item.validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.rType] = item
	return nil
}

// Method to register a config in the registry but with the struct name as part of the path
func (r *registry) RegisterRoot(c Configer) error {
	rItem := NewRegistryItem(c, true)
	c.SetDefaults()

	cWithRoot, err := wrapStructWithRootName(c, rItem.name)
	if err != nil {
		return errors.Wrap(err, "failed to register struct as root")
	}

	err = r.v.Unmarshal(cWithRoot)
	if err != nil {
		return err
	}

	removeSecretValuesRecursive(c)
	err = r.secretV.Unmarshal(cWithRoot)
	if err != nil {
		return err
	}

	r.restoreOverrides(rItem)

	if err = rItem.validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[rItem.rType] = rItem
	return nil
}

func (r *registry) Get(config Configer) Configer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.items[reflect.TypeOf(config).Elem()]; ok {
		return c.item
	}
	return nil
}

func (r *registry) Set(c Configer) error {
	item := NewRegistryItem(c, false)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.rType] = item
	return nil
}

func (r *registry) SetWithDefaults(c Configer) error {
	cWithDefaults := applyDefaults(c)
	item := NewRegistryItem(cWithDefaults, false)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.rType] = item
	return nil
}

func (r *registry) SetOverrides(c Configer) error {
	rType := reflect.TypeOf(c).Elem()
	overrides := cloneConfiger(c)

	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[rType]
	if !ok || item.item == nil {
		r.items[rType] = NewRegistryWithOverrides(overrides)
		return nil
	}

	cloned := cloneConfiger(item.item)
	newItem := NewRegistryItem(cloned, item.registeredRoot)
	newItem.overrides = overrides
	mergeOverridesRecursive(newItem.item, overrides)
	if err := newItem.validate(); err != nil {
		return err
	}

	r.items[rType] = newItem
	return nil
}

func (r *registry) FlushAll() {
	r.mu.Lock()
	kbSecretsWatcher := r.kbSecretsWatcher
	r.items = make(map[reflect.Type]*registryItem)
	r.secretV = viper.New()
	r.kbSecretsWatcher = nil
	r.v = initViperWithOptions(r.conf.KeyDelimiter)
	r.mu.Unlock()

	if kbSecretsWatcher != nil {
		_ = kbSecretsWatcher.Close()
	}
	r.confPool.FlushAll()
}

func (r *registry) OnConfigChange(callback func()) {
	r.callbackOnChange = callback
}

func (r *registry) onConfigPoolChange(vip *viper.Viper) {
	r.v = vip
	r.reloadRegistry()
}

func (r *registry) loadKbSecretsFolder(folderPath string) error {
	files, allErrors := os.ReadDir(folderPath)
	if allErrors != nil {
		return errors.Wrap(allErrors, "kb secrets folder failed to load")
	}

	allErrors = errors.New("read k8s secrets error")
	hadError := false
	for _, f := range files {
		if isDirectory, _ := isDir(folderPath, f.Name()); isDirectory {
			continue
		}

		value, err := os.ReadFile(filepath.Join(folderPath, f.Name())) //nolint:gosec // G304: reading config/secret files from a trusted, operator-configured directory
		if err != nil {
			allErrors = errors.Wrapf(allErrors, "error reading secret file %s, error: %v", f.Name(), err)
			hadError = true
			continue
		}
		r.secretV.Set(strings.ToLower(f.Name()), string(value))
	}

	if hadError {
		return allErrors
	}
	return nil
}

func (r *registry) watchKbSecretsFolder(watcher *fsnotify.Watcher, folderPath string) {
	debounce := time.NewTicker(kbSecretsReloadNever)
	defer debounce.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return // channel closed
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			debounce.Reset(kbSecretsReloadDebounce)
		case <-debounce.C:
			debounce.Reset(kbSecretsReloadNever)
			r.mu.Lock()
			err := r.loadKbSecretsFolder(folderPath)
			r.mu.Unlock()
			if err != nil {
				continue
			}

			r.reloadRegistry()
		case _, ok := <-watcher.Errors:
			if !ok {
				return // channel closed
			}
		}
	}
}

func (r *registry) initK8SWatcher(folderPath string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := watcher.Add(folderPath); err != nil {
		_ = watcher.Close()
		return nil, err
	}

	r.mu.Lock()
	if r.kbSecretsWatcher != nil {
		_ = r.kbSecretsWatcher.Close()
	}
	r.kbSecretsWatcher = watcher
	r.mu.Unlock()

	return watcher, nil
}

func (r *registry) reloadRegistry() {
	fieldsToReload := []reflect.Type{}
	r.mu.RLock()
	for key := range r.items {
		fieldsToReload = append(fieldsToReload, key)
	}
	r.mu.RUnlock()

	for _, field := range fieldsToReload {
		r.reloadItem(field)
	}

	r.callbackOnChange()
}

func (r *registry) reloadItem(field reflect.Type) {
	r.mu.RLock()
	item := r.items[field]
	r.mu.RUnlock()
	if item == nil || item.item == nil {
		return
	}

	if item.registeredRoot {
		_ = r.RegisterRoot(reflect.New(item.rType).Interface().(Configer))
		return
	}

	_ = r.Register(reflect.New(item.rType).Interface().(Configer))
}

func (r *registry) restoreOverrides(item *registryItem) {
	r.mu.RLock()
	existing, ok := r.items[item.rType]
	r.mu.RUnlock()
	if !ok || existing.overrides == nil {
		return
	}

	item.overrides = existing.overrides
	mergeOverridesRecursive(item.item, item.overrides)
}

func mergeOverridesRecursive(dst, src any) {
	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)
	if !isPointerToStruct(dstVal) || !isPointerToStruct(srcVal) {
		return
	}

	dstVal = dstVal.Elem()
	srcVal = srcVal.Elem()
	dstType := dstVal.Type()

	for i := 0; i < dstType.NumField(); i++ {
		if !dstType.Field(i).IsExported() {
			continue
		}
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)
		if srcField.Kind() == reflect.Ptr && !srcField.IsNil() {
			if dstField.IsNil() {
				dstField.Set(srcField)
				continue
			}

			mergeOverridesRecursive(dstField.Interface(), srcField.Interface())
			continue
		}

		if srcField.Kind() == reflect.Struct && !srcField.IsZero() {
			mergeOverridesRecursive(dstField.Addr().Interface(), srcField.Addr().Interface())
			continue
		}

		if !srcField.IsZero() {
			dstField.Set(srcField)
		}
	}
}

func removeSecretValuesRecursive(conf any) {
	rv := reflect.ValueOf(conf)
	if !isPointerToStruct(rv) {
		return
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		if val, ok := field.Tag.Lookup("config"); ok && val == "secret" {
			rv.Field(i).SetZero()
		}

		if rv.Field(i).Kind() == reflect.Ptr && !rv.Field(i).IsNil() {
			removeSecretValuesRecursive(rv.Field(i).Interface())
		}

		if rv.Field(i).Kind() == reflect.Struct {
			removeSecretValuesRecursive(rv.Field(i).Addr().Interface())
		}
	}
}

func trimConfigSuffixFromPascal(input string) string {
	re := regexp.MustCompile(`(_config|_conf|_cfg|_configuration)$`)
	return re.ReplaceAllString(input, "")
}

func pascalToSnake(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	return strings.ToLower(snake)
}

func wrapStructWithRootName(input Configer, fieldName string) (any, error) {
	// Ensure input is a struct or a pointer to a struct
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or a pointer to a struct")
	}

	// Create a new struct type with the specified field name embedding the input struct
	field := reflect.StructField{
		Name: fieldName,
		Type: reflect.TypeOf(input),
		Tag:  reflect.StructTag(fmt.Sprintf(`mapstructure:"%s"`, trimConfigSuffixFromPascal(pascalToSnake(fieldName)))),
	}

	wrapperType := reflect.StructOf([]reflect.StructField{field})

	// Create a new instance of the wrapper struct and set the embedded field
	wrapper := reflect.New(wrapperType).Elem()
	wrapper.Field(0).Set(reflect.ValueOf(input))

	// Return a pointer to the new wrapper struct
	return wrapper.Addr().Interface(), nil
}

func getEnvOrDefault(envVar string, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

func isDir(basePath string, entryName string) (bool, error) {
	info, err := os.Stat(filepath.Join(basePath, entryName))
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func applyDefaults(c Configer) Configer {
	cType := reflect.TypeOf(c).Elem()
	cWithDefaults := reflect.New(cType).Interface().(Configer)
	cWithDefaults.SetDefaults()

	var cMap map[string]interface{}
	err := mapstructure.Decode(c, &cMap)
	if err != nil {
		return cWithDefaults // fallback: just return defaults if decode fails
	}

	cfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           cWithDefaults,
		ZeroFields:       false, // do not zero fields before decoding
		WeaklyTypedInput: true,
		DecodeHook:       onlySetNonZeroHook(),
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err == nil {
		_ = decoder.Decode(cMap)
	}
	return cWithDefaults
}

func cloneConfiger(c Configer) Configer {
	clone := reflect.New(reflect.TypeOf(c).Elem()).Interface().(Configer)
	var cMap map[string]interface{}
	err := mapstructure.Decode(c, &cMap)
	if err != nil {
		return clone
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "mapstructure",
		Result:  clone,
	})
	if err == nil {
		_ = decoder.Decode(cMap)
	}
	return clone
}

func onlySetNonZeroHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if data == nil {
			return data, nil
		}
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return data, nil
		}
		if isZeroValue(v) {
			// Returning mapstructure.ComposeDecodeHookFunc will skip setting this field
			return mapstructure.ComposeDecodeHookFunc(), nil
		}
		return data, nil
	}
}

func isPointerToStruct(rv reflect.Value) bool {
	return rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct
}

// isZeroValue checks if a reflect.Value is zero for its type.
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}
