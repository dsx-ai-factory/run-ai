package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// JsonBMap allows storing/loading a JSONB column to/from a generic map and thus avoid strict
// structure of the record.
type JsonBMap map[string]interface{}

func (m JsonBMap) AsMAp() map[string]interface{} {
	return m
}

func (m JsonBMap) Value() (driver.Value, error) {
	if m == nil {
		return "null", nil
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

func (m *JsonBMap) Scan(value interface{}) error {
	var b []byte

	switch realValue := value.(type) {
	case []byte:
		b = realValue
	case string:
		b = []byte(realValue)
	default:
		return fmt.Errorf("invalid scan type for JsonBMap")
	}

	tmpMap, err := gojsonq.LoadJsonBytes(b)
	if err != nil {
		return err
	}
	*m = tmpMap
	return nil
}

func (m *JsonBMap) Unmarshal(value any) {
	bytes, err := json.Marshal(m)
	if err != nil {
		logging.Errorf("failed to convert json to bytes: %v", err)
		return
	}
	err = json.Unmarshal(bytes, value)
	if err != nil {
		logging.Errorf("failed to convert json to struct: %v", err)
	}
}

func (orig JsonBMap) Clone() JsonBMap {
	return cloneJsonMap(orig)
}

func cloneJsonMap(orig map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range orig {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Map:
			nestedCopy := cloneJsonMap(value.(map[string]interface{}))
			result[key] = nestedCopy
		case reflect.Slice:
			srcSlice := value.([]interface{})
			copiedSlice := make([]interface{}, len(srcSlice))
			copy(copiedSlice, srcSlice)
			result[key] = copiedSlice
		default:
			result[key] = value
		}
	}

	return result
}
