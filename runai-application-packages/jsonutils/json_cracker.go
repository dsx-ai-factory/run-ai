/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonutils

import (
	"reflect"
	"strings"
)

// JsonCracker can take a structure such as:
//
//	type Some struct {
//	    Foo bool `json:"foo"`
//	    Bar int  `json:"bar"`
//	}
//
// And returns information about the json tags it contains, so you do not have to define hard-coded constants
// for them. It is suggested to do this scanning once for all the fields you are interested in, and store the
// results for later use, rather than calling it again and again.
type JsonCracker[T any] struct {
	inst     *T
	allNames []string
	namesMap map[string]string
}

// NewJsonCracker gets an instance of a structure and parses its json tags
// For example:
//
//	jc := NewJsonCracker[Some](Some{})
func NewJsonCracker[T any](structPtr T) *JsonCracker[T] {
	result := JsonCracker[T]{
		inst:     &structPtr,              // an instance that will be used for finding the tags
		allNames: make([]string, 0),       // all the tags contained in the structure
		namesMap: make(map[string]string), // a mapping b/w field name and its json tag
	}

	// Get the type of the struct
	dataType := reflect.TypeOf(structPtr)

	// Iterate over the fields of the struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Get the JSON tag value of the field
		jsonTag := field.Tag.Get("json")

		// If the JSON tag is not empty and not "-" (indicating it should be omitted), load it
		if jsonTag != "" && jsonTag != "-" {
			jsonName := strings.Split(jsonTag, ",")[0]

			// take only the column name, ignore additional info provided following the comma
			result.allNames = append(result.allNames, jsonName)

			result.namesMap[field.Name] = jsonName
		}
	}

	return &result
}

// AllNames returns all the json names as an array, in the "Some" example it will return
// []string{ "foo", "bar" }
func (jc *JsonCracker[T]) AllNames() []string {
	return jc.allNames
}

// Struct returns a pointer to the struct. it should be used when calling to NameOf function because
// it relies to work on the same struct as was sent to the New function
func (jc *JsonCracker[T]) Struct() *T {
	return jc.inst
}

// NameOf returns the json tag name for a given field in the struct, or empty string if such tag
// was not found. Calling to this function must be done like this:
//
//	jc.NameOf(&jc.Struct().Foo) --> returns "foo"
func (jc *JsonCracker[T]) NameOf(ptr interface{}) string {
	val := reflect.ValueOf(jc.inst).Elem()
	val2 := reflect.ValueOf(ptr).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		if valueField.Addr().Interface() == val2.Addr().Interface() {
			fldName := val.Type().Field(i).Name
			return jc.namesMap[fldName]
		}
	}
	return ""
}
