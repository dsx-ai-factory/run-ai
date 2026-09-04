/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

type Path struct {
	Parent string
	Field  string
	Type   reflect.Kind
}

func (p *Path) GetFullPath() string {
	if p.Parent == "" {
		return p.Field
	}

	return fmt.Sprintf("%s.%s", p.Parent, p.Field)
}

type FieldPredicate func(field reflect.StructField) bool

type TypePathExplorer struct {
	Paths []*Path
}

func NewPathExplorer(predicate FieldPredicate, ts ...interface{}) *TypePathExplorer {
	paths := []*Path{}
	for _, t := range ts {
		paths = append(paths, getPaths(reflect.TypeOf(t), &Path{Parent: "", Type: reflect.Struct}, predicate)...)
	}

	return &TypePathExplorer{
		Paths: paths,
	}
}

func getPaths(typ reflect.Type, parent *Path, predicate FieldPredicate) []*Path {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	paths := []*Path{}
	if typ.Kind() != reflect.Struct {
		return paths
	}

	for i := 0; i < typ.NumField(); i++ {
		paths = getPathsFromField(predicate, typ.Field(i), parent, paths)
	}

	return paths
}

func getPathsFromField(predicate FieldPredicate, field reflect.StructField, parent *Path, paths []*Path) []*Path {
	fieldName := getJsonFieldName(field)

	var current *Path
	if strings.Contains(fieldName, "squash") {
		current = parent
	} else {
		current = &Path{
			Parent: parent.GetFullPath(),
			Field:  fieldName,
			Type:   field.Type.Kind(),
		}
	}

	if predicate(field) {
		return append(paths, current)
	}

	subPaths := getPaths(field.Type, current, predicate)
	paths = append(paths, subPaths...)

	return paths
}

func getJsonFieldName(field reflect.StructField) string {
	if field.Tag.Get("mapstructure") != "" {
		return field.Tag.Get("mapstructure")
	}

	if field.Tag.Get("json") != "" {
		return field.Tag.Get("json")
	}

	return field.Name
}
