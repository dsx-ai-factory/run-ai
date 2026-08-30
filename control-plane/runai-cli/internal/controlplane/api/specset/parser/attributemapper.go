package parser

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

type ParseAlgo interface {
	parse(value string) error
	GetMap() map[string]string
	Get(key string) *string
}

type AttributeMapper struct {
	value string
	err   error
	algo  ParseAlgo
}

func NewAttributeMapper(value string, algo ParseAlgo) *AttributeMapper {
	return &AttributeMapper{value: value, algo: algo}
}

func (am *AttributeMapper) Parse() error {
	am.err = am.parse()
	return am.err
}

func (am *AttributeMapper) parse() error {
	return am.algo.parse(am.value)
}

func (am *AttributeMapper) GetMap() map[string]string {
	return am.algo.GetMap()
}

func (am *AttributeMapper) Get(key string) *string {
	return am.algo.Get(key)
}

func CheckMandatoryFields(am *AttributeMapper, attributes ...string) error {
	var missing []string
	for _, attribute := range attributes {
		if am.Get(attribute) == nil {
			missing = append(missing, attribute)
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf(consts.ErrorAttributeMissingFields, missing, am.value)
	}
	return nil
}

func CheckRequiredNonEmpty(am *AttributeMapper, attributes ...string) error {
	var missing []string
	for _, attr := range attributes {
		p := am.Get(attr)
		if p == nil || *p == "" {
			missing = append(missing, attr)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf(consts.ErrorAttributeMissingFields, missing, am.value)
	}
	return nil
}

func CheckRequiredNonEmptyMap(parsed map[string]string, value string, keys ...string) error {
	var missing []string
	for _, k := range keys {
		if parsed[k] == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf(consts.ErrorAttributeMissingFields, missing, value)
	}
	return nil
}
