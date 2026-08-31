/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
)

// arrayKeyCtxKey is a private type used as the context key for arrayKeyCtx,
// preventing collisions with other keys in the conjungo options context.
type arrayKeyCtxKey struct{}

var arrayKey arrayKeyCtxKey

// arrayKeyCtx is the value passed via conjungo.Options.Context to
// arrayMergeByKey so it can identify array elements by a specific field.
type arrayKeyCtx struct {
	Arg       string
	SpecField string
}

// newSetMergeOptions builds the conjungo options used by the set enforcement:
// array merging by key and the overwrite flag. When ak is nil the standard
// conjungo slice behavior is used.
func newSetMergeOptions(overwrite bool, ak *RuntimeArrayKey) *conjungo.Options {
	options := conjungo.NewOptions()
	options.Overwrite = overwrite

	if ak != nil {
		options.SetKindMergeFunc(reflect.Slice, arrayMergeByKey)
		options.Context = context.WithValue(context.Background(), arrayKey, arrayKeyCtx{
			Arg:       ak.Arg,
			SpecField: ak.SpecField,
		})
	}

	return options
}

// arrayMergeByKey is a conjungo slice merge function that merges two arrays
// element-wise by matching elements that share the same value at a specified
// spec field. Elements present only in src are appended; when overwrite is true
// matching elements from src replace those in target.
func arrayMergeByKey(target, src reflect.Value, o *conjungo.Options) (reflect.Value, error) {
	arrayKeyInfo, ok := o.Context.Value(arrayKey).(arrayKeyCtx)
	if !ok {
		return target, fmt.Errorf("arrayMergeByKey: missing arrayKeyCtx in options context")
	}

	keySpecField := arrayKeyInfo.SpecField
	if keySpecField == "" {
		keySpecField = arrayKeyInfo.Arg
	}

	targetInstances, ok := target.Interface().([]any)
	if !ok {
		return target, fmt.Errorf("arrayMergeByKey: target is not []any (got %T)", target.Interface())
	}
	srcInstances, ok := src.Interface().([]any)
	if !ok {
		return target, fmt.Errorf("arrayMergeByKey: src is not []any (got %T)", src.Interface())
	}

	if len(srcInstances) == 0 {
		return target, nil
	}
	if len(targetInstances) == 0 {
		return src, nil
	}

	// When no arrayKey is set (e.g. envFrom entries), replace target with src.
	// A future "append" enforcement can instead concatenate the two arrays.
	if keySpecField == "" {
		return src, nil
	}

	// Record the position of each target element by key so we can overwrite
	// it in place (via target.Index(i).Set(...)) when a matching src element
	// is found and overwrite is enabled.
	indexByKey := make(map[string]int, len(targetInstances))
	for i, instance := range targetInstances {
		keyValue := getValueBySpecField(instance, keySpecField)
		if keyValue != nil {
			indexByKey[fmt.Sprintf("%v", keyValue)] = i
		}
	}

	elemType := target.Type().Elem()

	for _, instance := range srcInstances {
		keyValue := getValueBySpecField(instance, keySpecField)
		if keyValue == nil {
			target = reflect.Append(target, reflect.ValueOf(instance))
			continue
		}

		keyStr := fmt.Sprintf("%v", keyValue)
		idx, exists := indexByKey[keyStr]
		if !exists {
			target = reflect.Append(target, reflect.ValueOf(instance))
			continue
		}

		if !o.Overwrite {
			continue
		}

		srcValue := reflect.ValueOf(instance)
		if !srcValue.Type().AssignableTo(elemType) {
			if !srcValue.Type().ConvertibleTo(elemType) {
				return target, fmt.Errorf("arrayMergeByKey: cannot assign src element of type %s to target element type %s", srcValue.Type(), elemType)
			}
			srcValue = srcValue.Convert(elemType)
		}
		target.Index(idx).Set(srcValue)
	}

	return target, nil
}

// getValueBySpecField walks a dotted path inside a map[string]any instance and
// returns the value at that path, or nil if any segment is missing or the
// intermediate is not a map.
func getValueBySpecField(instance any, specField string) any {
	current := instance
	for _, part := range strings.Split(specField, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		v, ok := m[part]
		if !ok {
			return nil
		}
		current = v
	}
	return current
}
