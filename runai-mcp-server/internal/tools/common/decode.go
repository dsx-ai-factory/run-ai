package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// DecodeTypedRequest converts a polymorphic map field from an already-decoded
// tool input into its resolved API request type and rejects unknown fields that
// the generated API models would otherwise silently ignore.
func DecodeTypedRequest(value, out any) error {
	if err := rejectUnknownJSONFields(value, reflect.TypeOf(out)); err != nil {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func rejectUnknownJSONFields(value any, target reflect.Type) error {
	target = unwrapJSONTarget(target)
	if target == nil || value == nil {
		return nil
	}

	switch target.Kind() {
	case reflect.Struct:
		object, ok := stringMapValue(value)
		if !ok {
			return nil // Let encoding/json report the type mismatch.
		}
		fields := jsonStructFields(target)
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fieldValue := object[key]
			fieldType, exists := fields[key]
			if !exists {
				return fmt.Errorf("json: unknown field %q", key)
			}
			if err := rejectUnknownJSONFields(fieldValue, fieldType); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		items := reflect.ValueOf(value)
		if items.Kind() != reflect.Slice && items.Kind() != reflect.Array {
			return nil
		}
		for i := 0; i < items.Len(); i++ {
			if err := rejectUnknownJSONFields(items.Index(i).Interface(), target.Elem()); err != nil {
				return err
			}
		}
	case reflect.Map:
		object, ok := stringMapValue(value)
		if !ok {
			return nil
		}
		for _, fieldValue := range object {
			if err := rejectUnknownJSONFields(fieldValue, target.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}

func unwrapJSONTarget(target reflect.Type) reflect.Type {
	for target != nil && target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	if target != nil && target.Kind() == reflect.Struct && strings.HasPrefix(target.Name(), "Nullable") {
		if valueField, ok := target.FieldByName("value"); ok {
			return unwrapJSONTarget(valueField.Type)
		}
	}
	return target
}

func jsonStructFields(target reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type)
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			if field.Anonymous {
				embedded := unwrapJSONTarget(field.Type)
				if embedded != nil && embedded.Kind() == reflect.Struct {
					for embeddedName, embeddedType := range jsonStructFields(embedded) {
						fields[embeddedName] = embeddedType
					}
				}
				continue
			}
			name = field.Name
		}
		fields[name] = field.Type
	}
	return fields
}

func stringMapValue(value any) (map[string]any, bool) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		out[iter.Key().String()] = iter.Value().Interface()
	}
	return out, true
}
