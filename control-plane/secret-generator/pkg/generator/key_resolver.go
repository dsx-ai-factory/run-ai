package generator

import "fmt"

// KeyResolver resolves a secret key value from its spec.
type KeyResolver interface {
	CanResolve(spec SecretKeySpec, secretData map[string][]byte) bool
	Resolve(spec SecretKeySpec, secretData map[string][]byte) ([]byte, error)
}

// ChartValueResolver uses the chart-supplied value when set.
// When AllowEmpty is false (default), empty strings are ignored so
// downstream resolvers (legacy, generate, copyFrom) get a chance to run.
type ChartValueResolver struct {
	AllowEmpty bool
}

func (r ChartValueResolver) CanResolve(spec SecretKeySpec, _ map[string][]byte) bool {
	if spec.ChartValue == nil {
		return false
	}
	return r.AllowEmpty || *spec.ChartValue != ""
}

func (ChartValueResolver) Resolve(spec SecretKeySpec, _ map[string][]byte) ([]byte, error) {
	return []byte(*spec.ChartValue), nil
}

// CopyFromOnlyResolver copies from the source key.
type CopyFromOnlyResolver struct{}

func (CopyFromOnlyResolver) CanResolve(spec SecretKeySpec, secretData map[string][]byte) bool {
	if spec.CopyFrom == "" || spec.CopyFrom == spec.Name {
		return false
	}
	_, ok := secretData[spec.CopyFrom]
	return ok
}

func (CopyFromOnlyResolver) Resolve(spec SecretKeySpec, secretData map[string][]byte) ([]byte, error) {
	value := secretData[spec.CopyFrom]
	if value == nil {
		return nil, fmt.Errorf("key %s: copyFrom source %s is missing", spec.Name, spec.CopyFrom)
	}
	return value, nil
}

// generateOnlyResolver generates a value from the spec.
type generateOnlyResolver struct{}

func (generateOnlyResolver) CanResolve(spec SecretKeySpec, _ map[string][]byte) bool {
	return spec.CopyFrom == ""
}

func (generateOnlyResolver) Resolve(spec SecretKeySpec, _ map[string][]byte) ([]byte, error) {
	if spec.ExternalOnly {
		return nil, fmt.Errorf("key %q is marked as externalOnly but no explicit value was provided (via chartValue, existingSecret, or legacy secret); random generation is not allowed for externalOnly keys", spec.Name)
	}
	value, err := GenerateValue(spec)
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// ChainResolver tries each resolver in order.
type ChainResolver struct {
	Chain []KeyResolver
}

func (c ChainResolver) CanResolve(spec SecretKeySpec, secretData map[string][]byte) bool {
	for _, r := range c.Chain {
		if r.CanResolve(spec, secretData) {
			return true
		}
	}
	return false
}

func (c ChainResolver) Resolve(spec SecretKeySpec, secretData map[string][]byte) ([]byte, error) {
	for _, r := range c.Chain {
		if r.CanResolve(spec, secretData) {
			return r.Resolve(spec, secretData)
		}
	}
	return nil, fmt.Errorf("no resolver in chain could resolve key %s", spec.Name)
}

// Resolver type names.
const (
	ResolverChartValue = "chartValue"
	ResolverStatic     = "static"
	ResolverCopyFrom   = "copyFrom"
	ResolverGenerate   = "generate"
)

var defaultResolver map[string]KeyResolver = map[string]KeyResolver{
	ResolverStatic:   &ChainResolver{Chain: []KeyResolver{ChartValueResolver{AllowEmpty: true}}},
	ResolverCopyFrom: &ChainResolver{Chain: []KeyResolver{ChartValueResolver{}, CopyFromOnlyResolver{}}},
	ResolverGenerate: &ChainResolver{Chain: []KeyResolver{ChartValueResolver{}, generateOnlyResolver{}}},
}

// ResolveKeyData resolves all keys; supports CopyFrom chains via multiple passes.
func ResolveKeyData(keys []SecretKeySpec, legacySecretData map[string][]byte) (map[string][]byte, error) {
	data := make(map[string][]byte)
	for k, v := range legacySecretData {
		data[k] = v
	}

	maxPasses := len(keys) + 1
	for pass := 0; pass < maxPasses; pass++ {
		resolvedAny := false
		for _, spec := range keys {
			if _, ok := data[spec.Name]; ok {
				continue
			}
			value, err := ResolveKey(spec, data)
			if err != nil {
				return nil, fmt.Errorf("key %s: %w", spec.Name, err)
			}
			data[spec.Name] = value
			resolvedAny = true
			break
		}
		if !resolvedAny {
			break
		}
	}

	var unresolved []string
	for _, spec := range keys {
		if _, ok := data[spec.Name]; !ok {
			if spec.CopyFrom != "" {
				unresolved = append(unresolved, fmt.Sprintf("%s (from %s)", spec.Name, spec.CopyFrom))
			} else {
				unresolved = append(unresolved, spec.Name)
			}
		}
	}
	if len(unresolved) > 0 {
		return nil, fmt.Errorf("unresolvable keys (possible cycle or missing source): %v", unresolved)
	}

	return data, nil
}

func ResolveKey(spec SecretKeySpec, data map[string][]byte) ([]byte, error) {
	if spec.Type == "static" {
		return defaultResolver[ResolverStatic].Resolve(spec, data)
	}
	if spec.CopyFrom != "" {
		return defaultResolver[ResolverCopyFrom].Resolve(spec, data)
	}
	return defaultResolver[ResolverGenerate].Resolve(spec, data)
}
