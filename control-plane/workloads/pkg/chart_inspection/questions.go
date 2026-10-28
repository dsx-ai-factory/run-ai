// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const (
	maxChildQuestionDepth    = 4
	literalTrue              = "true"
	literalFalse             = "false"
	questionFloat            = "float"
	questionInteger          = "int"
	questionEnum             = "enum"
	propertyDescription      = "description"
	propertyLabel            = "label"
	propertyMin              = "min"
	propertyMax              = "max"
	propertyMinLength        = "min_length"
	propertyMaxLength        = "max_length"
	warningUnusableQuestions = "UNUSABLE_QUESTIONS"
)

var questionTypes = datastructs.NewSet(strings.Fields("string multiline int float boolean enum radio hostname ipaddr cidr cron array[string] map[string] password storageclass pvc configmap")...)
var questionProperties = datastructs.NewSet(
	strings.Fields("variable label description type default required group min max min_length max_length options multiple show_if show_subquestion_if show_subquestions_if subquestions")...,
)

type declarationValidator struct {
	defaults     map[string]any
	warnings     []api.AiApplicationChartInspectionWarning
	groups       []map[string]any
	groupNames   map[string]bool
	questions    []declarationQuestion
	destinations pathTree
}

// Build the form separately and publish it only after all fields and visibility
// dependencies validate. A partial form could omit required configuration.
func parseDeclaration(declaration, defaults map[string]any, out *api.AiApplicationChartInspectionResponse) {
	v := &declarationValidator{
		defaults:   defaults,
		groupNames: map[string]bool{},
		warnings:   []api.AiApplicationChartInspectionWarning{},
		groups:     []map[string]any{},
	}
	v.ignoreUnknown(declaration, datastructs.NewSet(strings.Fields("questions groups")...), "$")
	if groups, exists := declaration["groups"]; exists {
		if err := v.parseGroups(groups); err != nil {
			v.unusable(out, err)
			return
		}
	}
	list, ok := declaration["questions"].([]any)
	if !ok {
		v.unusable(out, api.NewValidationError("questions must be a list"))
		return
	}
	questions := make([]map[string]any, 0, len(list))
	for index, item := range list {
		question, err := v.question(item, 0, "", "", fmt.Sprintf("questions[%d]", index))
		if err != nil {
			v.unusable(out, err)
			return
		}
		questions = append(questions, question)
	}
	if err := v.checkConditions(); err != nil {
		v.unusable(out, err)
		return
	}
	out.QuestionsStatus = api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE
	out.Questions = questions
	out.Groups = v.groups
	out.Warnings = append(out.Warnings, v.warnings...)
}

func (v *declarationValidator) unusable(out *api.AiApplicationChartInspectionResponse, err *api.Error) {
	out.QuestionsStatus = api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE
	out.Questions = []map[string]any{}
	out.Groups = []map[string]any{}
	out.Warnings = append(out.Warnings, v.warnings...)
	out.Warnings = append(out.Warnings, api.AiApplicationChartInspectionWarning{Code: warningUnusableQuestions, Message: err.Message, Path: "$"})
}

func (v *declarationValidator) ignoreUnknown(raw map[string]any, known datastructs.Set[string], where string) {
	unknown := []string{}
	for key := range raw {
		if !known.Has(key) {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	for _, key := range unknown {
		v.warnings = append(v.warnings, api.AiApplicationChartInspectionWarning{Code: "IGNORED_PROPERTY", Message: "Unsupported optional property was ignored", Path: where + "." + key})
	}
}

func (v *declarationValidator) parseGroups(value any) *api.Error {
	list, ok := value.([]any)
	if !ok {
		return api.NewValidationError("groups must be a list")
	}
	for index, item := range list {
		where := fmt.Sprintf("groups[%d]", index)
		raw, ok := item.(map[string]any)
		if !ok {
			return api.NewValidationError(fmt.Sprintf("%s must be an object", where))
		}
		name, ok := raw["name"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return api.NewValidationError(fmt.Sprintf("%s requires a name", where))
		}
		key := strings.ToLower(strings.TrimSpace(name))
		if v.groupNames[key] {
			return api.NewValidationError(fmt.Sprintf("duplicate group name %s", name))
		}
		label := strings.TrimSpace(name)
		if text, exists := raw[propertyLabel]; exists {
			supplied, ok := text.(string)
			if !ok || strings.TrimSpace(supplied) == "" {
				return api.NewValidationError(fmt.Sprintf("%s label must be text", where))
			}
			label = strings.TrimSpace(supplied)
		}
		description := ""
		if text, exists := raw[propertyDescription]; exists {
			supplied, ok := text.(string)
			if !ok {
				return api.NewValidationError(fmt.Sprintf("%s description must be text", where))
			}
			description = strings.TrimSpace(supplied)
		}
		v.ignoreUnknown(raw, datastructs.NewSet(strings.Fields("name label description")...), where)
		v.groups = append(v.groups, map[string]any{"name": key, propertyLabel: label, propertyDescription: description})
		v.groupNames[key] = true
	}
	return nil
}

// Root fields start at depth zero; four child levels allow five levels total.
// Child paths are absolute chart values paths, not relative to their parent.
func (v *declarationValidator) question(value any, depth int, parent, parentGroup, where string) (map[string]any, *api.Error) {
	if depth > maxChildQuestionDepth {
		return nil, api.NewValidationError(fmt.Sprintf("%s exceeds four child levels", where))
	}
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, api.NewValidationError(fmt.Sprintf("%s must be an object", where))
	}
	variable, ok := raw["variable"].(string)
	if !ok {
		return nil, api.NewValidationError(fmt.Sprintf("%s requires a variable", where))
	}
	segments, err := parsePath(variable)
	if err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("%s variable: %s", where, err.Message))
	}
	if err := v.destinations.insert(segments); err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("%s variable: %s", where, err.Message))
	}
	// Keep stage order: warnings and the first validation error are part of the response.
	question, err := v.normalizeMetadata(raw, variable, where)
	if err != nil {
		return nil, err
	}
	group, err := v.normalizeGroup(raw, question, parent, parentGroup, where)
	if err != nil {
		return nil, err
	}
	if err := v.normalizeConstraints(raw, question, where); err != nil {
		return nil, err
	}
	if err := normalizeDefault(raw, question, where); err != nil {
		return nil, err
	}
	childList, hasChildren, err := normalizeChildGate(raw, question, where)
	if err != nil {
		return nil, err
	}
	key, err := v.trackQuestion(raw, question, segments, parent, where)
	if err != nil {
		return nil, err
	}
	if hasChildren {
		normalized := make([]any, 0, len(childList))
		for index, child := range childList {
			parsed, err := v.question(child, depth+1, key, group, fmt.Sprintf("%s.subquestions[%d]", where, index))
			if err != nil {
				return nil, err
			}
			normalized = append(normalized, parsed)
		}
		question["subquestions"] = normalized
	}
	return question, nil
}

func (v *declarationValidator) normalizeMetadata(raw map[string]any, variable, where string) (map[string]any, *api.Error) {
	kind, ok := raw["type"].(string)
	if !ok || strings.TrimSpace(kind) == "" {
		return nil, api.NewValidationError(fmt.Sprintf("%s requires a type", where))
	}
	kind = strings.TrimSpace(kind)
	if strings.EqualFold(kind, "secret") || strings.Contains(strings.ToLower(kind), "credential") {
		return nil, api.NewValidationError(fmt.Sprintf("%s uses an excluded secret or credential field", where))
	}
	if kind == "number" {
		kind = questionFloat
	} else if !questionTypes.Has(kind) {
		v.warnings = append(v.warnings, api.AiApplicationChartInspectionWarning{Code: "UNKNOWN_TYPE", Message: "Unknown field type uses a text field", Path: where + ".type"})
		kind = "string"
	}
	question := map[string]any{"variable": variable, "type": kind, propertyLabel: variable, propertyDescription: "", "required": false}
	labelDefaulted := true
	if label, exists := raw[propertyLabel]; exists {
		text, ok := label.(string)
		if !ok {
			return nil, api.NewValidationError(fmt.Sprintf("%s label must be text", where))
		}
		if strings.TrimSpace(text) != "" {
			question[propertyLabel] = text
			labelDefaulted = false
		}
	}
	if labelDefaulted {
		v.warnings = append(v.warnings, api.AiApplicationChartInspectionWarning{Code: "LABEL_DEFAULTED", Message: "Missing field label uses its values path", Path: where + ".label"})
	}
	for _, key := range []string{propertyDescription, "group", "show_if"} {
		if item, exists := raw[key]; exists {
			if _, ok := item.(string); !ok {
				return nil, api.NewValidationError(fmt.Sprintf("%s %s must be text", where, key))
			}
			question[key] = item
		}
	}
	for _, key := range []string{"required", "multiple"} {
		if item, exists := raw[key]; exists {
			if _, ok := item.(bool); !ok {
				return nil, api.NewValidationError(fmt.Sprintf("%s %s must be a boolean", where, key))
			}
			question[key] = item
		}
	}
	if _, exists := question["multiple"]; exists && kind != questionEnum {
		return nil, api.NewValidationError(fmt.Sprintf("%s multiple is supported only for enum", where))
	}
	return question, nil
}

func (v *declarationValidator) normalizeGroup(raw, question map[string]any, parent, parentGroup, where string) (string, *api.Error) {
	group := parentGroup
	if text, ok := raw["group"].(string); ok && strings.TrimSpace(text) != "" {
		group = strings.ToLower(strings.TrimSpace(text))
	}
	if parent != "" && group != parentGroup {
		return "", api.NewValidationError(fmt.Sprintf("%s must inherit its parent group", where))
	}
	question["group"] = group
	if !v.groupNames[group] {
		label := "Configuration"
		if text, ok := raw["group"].(string); ok && group != "" {
			label = strings.TrimSpace(text)
		}
		v.groups = append(v.groups, map[string]any{"name": group, propertyLabel: label, propertyDescription: ""})
		v.groupNames[group] = true
	}
	return group, nil
}

func (v *declarationValidator) normalizeConstraints(raw, question map[string]any, where string) *api.Error {
	kind := question["type"].(string)
	for _, key := range []string{propertyMin, propertyMax, propertyMinLength, propertyMaxLength} {
		item, exists := raw[key]
		if !exists {
			continue
		}
		number, ok := numeric(item)
		if !ok {
			return api.NewValidationError(fmt.Sprintf("%s %s must be numeric", where, key))
		}
		if key == propertyMin || key == propertyMax {
			if kind != questionInteger && kind != questionFloat {
				return api.NewValidationError(fmt.Sprintf("%s %s requires a numeric field", where, key))
			}
		} else if !textTypes.Has(kind) || number.Sign() < 0 || !number.IsInt() {
			return api.NewValidationError(fmt.Sprintf("%s %s requires text and a nonnegative integer", where, key))
		}
		question[key] = item
	}
	for _, pair := range [][2]string{{propertyMin, propertyMax}, {propertyMinLength, propertyMaxLength}} {
		low, lowOK := numeric(question[pair[0]])
		high, highOK := numeric(question[pair[1]])
		if lowOK && highOK && low.Cmp(high) > 0 {
			return api.NewValidationError(fmt.Sprintf("%s minimum exceeds maximum", where))
		}
	}
	if kind == questionEnum || kind == "radio" {
		options, err := v.normalizeOptions(raw["options"], where+".options")
		if err != nil {
			return api.NewValidationError(fmt.Sprintf("%s options: %s", where, err.Message))
		}
		question["options"] = options
	} else if _, exists := raw["options"]; exists {
		return api.NewValidationError(fmt.Sprintf("%s options require enum or radio", where))
	}
	return nil
}

func normalizeDefault(raw, question map[string]any, where string) *api.Error {
	kind := question["type"].(string)
	if item, exists := raw["default"]; exists {
		if kind == "boolean" {
			if text, ok := item.(string); ok && (text == literalTrue || text == literalFalse) {
				item = text == literalTrue
			}
		}
		if err := validateAnswer(question, item); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s default: %s", where, err.Message))
		}
		question["default"] = item
	}
	return nil
}

func normalizeChildGate(raw, question map[string]any, where string) ([]any, bool, *api.Error) {
	kind := question["type"].(string)
	// Accept both spellings, normalize to singular, and reject conflicting values.
	gate, hasGate := raw["show_subquestion_if"]
	if plural, exists := raw["show_subquestions_if"]; exists {
		if hasGate && !equalScalar(gate, plural) {
			return nil, false, api.NewValidationError(fmt.Sprintf("%s contains conflicting child gates", where))
		}
		gate = plural
		hasGate = true
	}
	children, hasChildren := raw["subquestions"]
	var childList []any
	if hasChildren {
		var ok bool
		childList, ok = children.([]any)
		if !ok {
			return nil, false, api.NewValidationError(fmt.Sprintf("%s subquestions must be a list", where))
		}
	}
	if hasGate {
		if !isScalar(gate) {
			return nil, false, api.NewValidationError(fmt.Sprintf("%s child gate must be a scalar", where))
		}
		if len(childList) > 0 && (kind == "array[string]" || kind == "map[string]" || question["multiple"] == true) {
			return nil, false, api.NewValidationError(fmt.Sprintf("%s child gate requires a scalar parent", where))
		}
		question["show_subquestion_if"] = gate
	}
	return childList, hasChildren, nil
}

func (v *declarationValidator) trackQuestion(raw, question map[string]any, segments []any, parent, where string) (string, *api.Error) {
	variable := question["variable"].(string)
	v.ignoreUnknown(raw, questionProperties, where)
	if _, exists := readPath(v.defaults, segments); v.defaults != nil && !exists {
		v.warnings = append(v.warnings, api.AiApplicationChartInspectionWarning{Code: "MISSING_DEFAULT_PATH", Message: "Field path is absent from the chart defaults", Path: variable})
	}
	references := []string{}
	var err *api.Error
	if text, ok := question["show_if"].(string); ok {
		references, err = parseCondition(text)
		if err != nil {
			return "", api.NewValidationError(fmt.Sprintf("%s show_if: %s", where, err.Message))
		}
	}
	key := canonicalPath(segments)
	v.questions = append(v.questions, declarationQuestion{key: key, parent: parent, references: references})
	return key, nil
}

func (v *declarationValidator) normalizeOptions(value any, where string) ([]any, *api.Error) {
	list, ok := value.([]any)
	if !ok || len(list) == 0 {
		return nil, api.NewValidationError("nonempty options are required")
	}
	out := make([]any, 0, len(list))
	values := map[string]bool{}
	for index, item := range list {
		label := ""
		value := item
		if object, ok := item.(map[string]any); ok {
			var valid bool
			label, valid = object[propertyLabel].(string)
			if !valid {
				return nil, api.NewValidationError("option labels must be text")
			}
			var exists bool
			value, exists = object["value"]
			if !exists {
				return nil, api.NewValidationError("options require a label and value")
			}
			v.ignoreUnknown(object, datastructs.NewSet(strings.Fields("label value")...), fmt.Sprintf("%s[%d]", where, index))
		} else {
			label = fmt.Sprint(value)
			if value == nil {
				label = "null"
			}
		}
		if !isScalar(value) {
			return nil, api.NewValidationError("options must have scalar values")
		}
		key := scalarKey(value)
		if values[key] {
			return nil, api.NewValidationError("option values must be distinct")
		}
		values[key] = true
		out = append(out, map[string]any{propertyLabel: label, "value": value})
	}
	return out, nil
}
