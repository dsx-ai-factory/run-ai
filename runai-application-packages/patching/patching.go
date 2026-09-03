/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patching

import (
	"fmt"
	"reflect"
)

// patching related functions, worth moving to runai-common-applications when we're done with them

// CheckIfAllMembersAreNil takes a pointer to a structure and checks if
//
//	(1) all its members are pointers
//	(2) all pointers are nil
//
// note that the function is NOT intended to be recursive and traverse inner structures, it only
// checks the given struct itself
func CheckIfAllMembersAreNil(input interface{}) bool {
	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return false
	}

	// Get the underlying struct
	val = val.Elem()

	// Iterate over the fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Check if the field is a pointer, array, etc...
		kind := field.Kind()

		if kind != reflect.Ptr && kind != reflect.Struct && kind != reflect.Array && kind != reflect.Slice && kind != reflect.Map {
			return false
		}

		if !field.IsNil() {
			return false
		}
	}

	return true
}

// ApplyPatch gets two structures of identical type, one is existing data the other
// is a patch request. it applies the patch request on the existing data
// NOTE!
//   - the function only works on pointer fields, it ignores any field or struct
//     that is not a pointer
//   - the data parametrer must be a valid pointer to an initialized struct
func ApplyPatch(data interface{}, patch interface{}) {
	dataVal := reflect.ValueOf(data).Elem()
	patchVal := reflect.ValueOf(patch).Elem()

	// Ensure both are structs
	if dataVal.Kind() != reflect.Struct || patchVal.Kind() != reflect.Struct {
		panic(fmt.Sprintf("internal: both data and patch must be pointers to structs. but they are %s, %s",
			dataVal.Kind().String(), patchVal.Kind().String()))
	}

	for i := 0; i < dataVal.NumField(); i++ {
		dataField := dataVal.Field(i)
		patchField := patchVal.Field(i)

		if !dataField.CanSet() {
			panic("internal: structure contains non-exportable field")
		}

		// If the patch field is nil, skip -> nil field does not change the original data
		if patchField.IsNil() {
			continue
		}

		// For non-struct fields, directly set the pointer
		if patchField.Kind() != reflect.Ptr || patchField.Elem().Kind() != reflect.Struct {
			dataField.Set(patchField)
			continue
		}

		// Allocate the structure if it is nil, we have to put stuff into it
		if dataField.IsNil() {
			dataField.Set(reflect.New(dataField.Type().Elem()))
		}

		ApplyPatch(dataField.Interface(), patchField.Interface())

		if CheckIfAllMembersAreNil(dataField.Interface()) {
			dataField.Set(reflect.Zero(dataField.Type()))
		}
	}
}

// PatchField receives a pointer to a field and patch it with a new value, if given.
// it allows implementing patch API easily
func PatchField[T comparable](updatedValue **T, requestedValue *T) {
	if requestedValue != nil {
		*updatedValue = requestedValue
	}
}
