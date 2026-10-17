// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
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

var questionTypes = words("string multiline int float boolean enum radio hostname ipaddr cidr cron array[string] map[string] password storageclass pvc configmap")
var questionProperties = words("variable label description type default required group min max min_length max_length options multiple show_if show_subquestion_if show_subquestions_if subquestions")
var textTypes = words("string multiline password storageclass pvc configmap hostname ipaddr cidr cron")
var hostnamePart = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var cronPart = regexp.MustCompile(`^(\*|\d+(?:-\d+)?)(?:/(\d+))?$`)

type declarationQuestion struct {
	key, parent string
	references  []string
}
type pathTree struct {
	terminal  bool
	container byte
	children  map[any]*pathTree
}
type declarationValidator struct {
	defaults     map[string]any
	warnings     []Warning
	groups       []map[string]any
	groupNames   map[string]bool
	questions    []declarationQuestion
	destinations pathTree
}

func words(value string) map[string]bool {
	out := map[string]bool{}
	for _, word := range strings.Fields(value) {
		out[word] = true
	}
	return out
}

// Build the form separately and publish it only after all fields and visibility
// dependencies validate. A partial form could omit required configuration.
func parseDeclaration(declaration, defaults map[string]any, out *Result) {
	v := &declarationValidator{defaults: defaults, groupNames: map[string]bool{}, warnings: []Warning{}, groups: []map[string]any{}}
	v.ignoreUnknown(declaration, words("questions groups"), "$")
	if groups, exists := declaration["groups"]; exists {
		if err := v.parseGroups(groups); err != nil {
			v.unusable(out, err)
			return
		}
	}
	list, ok := declaration["questions"].([]any)
	if !ok {
		v.unusable(out, fmt.Errorf("questions must be a list"))
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
	out.QuestionsStatus = QuestionsUsable
	out.Questions = questions
	out.Groups = v.groups
	out.Warnings = append(out.Warnings, v.warnings...)
}

func (v *declarationValidator) unusable(out *Result, err error) {
	out.QuestionsStatus = QuestionsUnusable
	out.Questions = []map[string]any{}
	out.Groups = []map[string]any{}
	out.Warnings = append(out.Warnings, v.warnings...)
	out.Warnings = append(out.Warnings, Warning{Code: warningUnusableQuestions, Message: err.Error(), Path: "$"})
}

func (v *declarationValidator) ignoreUnknown(raw map[string]any, known map[string]bool, where string) {
	unknown := []string{}
	for key := range raw {
		if !known[key] {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	for _, key := range unknown {
		v.warnings = append(v.warnings, Warning{Code: "IGNORED_PROPERTY", Message: "Unsupported optional property was ignored", Path: where + "." + key})
	}
}

func (v *declarationValidator) parseGroups(value any) error {
	list, ok := value.([]any)
	if !ok {
		return fmt.Errorf("groups must be a list")
	}
	for index, item := range list {
		where := fmt.Sprintf("groups[%d]", index)
		raw, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be an object", where)
		}
		name, ok := raw["name"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return fmt.Errorf("%s requires a name", where)
		}
		key := strings.ToLower(strings.TrimSpace(name))
		if v.groupNames[key] {
			return fmt.Errorf("duplicate group name %s", name)
		}
		label := strings.TrimSpace(name)
		if text, exists := raw[propertyLabel]; exists {
			supplied, ok := text.(string)
			if !ok || strings.TrimSpace(supplied) == "" {
				return fmt.Errorf("%s label must be text", where)
			}
			label = strings.TrimSpace(supplied)
		}
		description := ""
		if text, exists := raw[propertyDescription]; exists {
			supplied, ok := text.(string)
			if !ok {
				return fmt.Errorf("%s description must be text", where)
			}
			description = strings.TrimSpace(supplied)
		}
		v.ignoreUnknown(raw, words("name label description"), where)
		v.groups = append(v.groups, map[string]any{"name": key, propertyLabel: label, propertyDescription: description})
		v.groupNames[key] = true
	}
	return nil
}

// Root fields start at depth zero; four child levels allow five levels total.
// Child paths are absolute chart values paths, not relative to their parent.
func (v *declarationValidator) question(value any, depth int, parent, parentGroup, where string) (map[string]any, error) {
	if depth > 4 {
		return nil, fmt.Errorf("%s exceeds four child levels", where)
	}
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", where)
	}
	variable, ok := raw["variable"].(string)
	if !ok {
		return nil, fmt.Errorf("%s requires a variable", where)
	}
	segments, err := parsePath(variable)
	if err != nil {
		return nil, fmt.Errorf("%s variable: %w", where, err)
	}
	if err := v.destinations.insert(segments); err != nil {
		return nil, fmt.Errorf("%s variable: %w", where, err)
	}
	kind, ok := raw["type"].(string)
	if !ok || strings.TrimSpace(kind) == "" {
		return nil, fmt.Errorf("%s requires a type", where)
	}
	kind = strings.TrimSpace(kind)
	if strings.EqualFold(kind, "secret") || strings.Contains(strings.ToLower(kind), "credential") {
		return nil, fmt.Errorf("%s uses an excluded secret or credential field", where)
	}
	if kind == "number" {
		kind = questionFloat
	} else if !questionTypes[kind] {
		v.warnings = append(v.warnings, Warning{Code: "UNKNOWN_TYPE", Message: "Unknown field type uses a text field", Path: where + ".type"})
		kind = "string"
	}
	question := map[string]any{"variable": variable, "type": kind, propertyLabel: variable, propertyDescription: "", "required": false}
	labelDefaulted := true
	if label, exists := raw[propertyLabel]; exists {
		text, ok := label.(string)
		if !ok {
			return nil, fmt.Errorf("%s label must be text", where)
		}
		if strings.TrimSpace(text) != "" {
			question[propertyLabel] = text
			labelDefaulted = false
		}
	}
	if labelDefaulted {
		v.warnings = append(v.warnings, Warning{Code: "LABEL_DEFAULTED", Message: "Missing field label uses its values path", Path: where + ".label"})
	}
	for _, key := range []string{propertyDescription, "group", "show_if"} {
		if item, exists := raw[key]; exists {
			if _, ok := item.(string); !ok {
				return nil, fmt.Errorf("%s %s must be text", where, key)
			}
			question[key] = item
		}
	}
	for _, key := range []string{"required", "multiple"} {
		if item, exists := raw[key]; exists {
			if _, ok := item.(bool); !ok {
				return nil, fmt.Errorf("%s %s must be a boolean", where, key)
			}
			question[key] = item
		}
	}
	if _, exists := question["multiple"]; exists && kind != questionEnum {
		return nil, fmt.Errorf("%s multiple is supported only for enum", where)
	}
	group := parentGroup
	if text, ok := raw["group"].(string); ok && strings.TrimSpace(text) != "" {
		group = strings.ToLower(strings.TrimSpace(text))
	}
	if parent != "" && group != parentGroup {
		return nil, fmt.Errorf("%s must inherit its parent group", where)
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
	for _, key := range []string{propertyMin, propertyMax, propertyMinLength, propertyMaxLength} {
		item, exists := raw[key]
		if !exists {
			continue
		}
		number, ok := numeric(item)
		if !ok {
			return nil, fmt.Errorf("%s %s must be numeric", where, key)
		}
		if key == propertyMin || key == propertyMax {
			if kind != questionInteger && kind != questionFloat {
				return nil, fmt.Errorf("%s %s requires a numeric field", where, key)
			}
		} else if !textTypes[kind] || number.Sign() < 0 || !number.IsInt() {
			return nil, fmt.Errorf("%s %s requires text and a nonnegative integer", where, key)
		}
		question[key] = item
	}
	for _, pair := range [][2]string{{propertyMin, propertyMax}, {propertyMinLength, propertyMaxLength}} {
		low, lowOK := numeric(question[pair[0]])
		high, highOK := numeric(question[pair[1]])
		if lowOK && highOK && low.Cmp(high) > 0 {
			return nil, fmt.Errorf("%s minimum exceeds maximum", where)
		}
	}
	if kind == questionEnum || kind == "radio" {
		options, err := v.normalizeOptions(raw["options"], where+".options")
		if err != nil {
			return nil, fmt.Errorf("%s options: %w", where, err)
		}
		question["options"] = options
	} else if _, exists := raw["options"]; exists {
		return nil, fmt.Errorf("%s options require enum or radio", where)
	}
	if item, exists := raw["default"]; exists {
		if kind == "boolean" {
			if text, ok := item.(string); ok && (text == literalTrue || text == literalFalse) {
				item = text == literalTrue
			}
		}
		if err := validateAnswer(question, item); err != nil {
			return nil, fmt.Errorf("%s default: %w", where, err)
		}
		question["default"] = item
	}
	gate, hasGate := raw["show_subquestion_if"]
	if plural, exists := raw["show_subquestions_if"]; exists {
		if hasGate && !equalScalar(gate, plural) {
			return nil, fmt.Errorf("%s contains conflicting child gates", where)
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
			return nil, fmt.Errorf("%s subquestions must be a list", where)
		}
	}
	if hasGate {
		if !isScalar(gate) {
			return nil, fmt.Errorf("%s child gate must be a scalar", where)
		}
		if len(childList) > 0 && (kind == "array[string]" || kind == "map[string]" || question["multiple"] == true) {
			return nil, fmt.Errorf("%s child gate requires a scalar parent", where)
		}
		question["show_subquestion_if"] = gate
	}
	v.ignoreUnknown(raw, questionProperties, where)
	if _, exists := readPath(v.defaults, segments); v.defaults != nil && !exists {
		v.warnings = append(v.warnings, Warning{Code: "MISSING_DEFAULT_PATH", Message: "Field path is absent from the chart defaults", Path: variable})
	}
	references := []string{}
	if text, ok := question["show_if"].(string); ok {
		references, err = parseCondition(text)
		if err != nil {
			return nil, fmt.Errorf("%s show_if: %w", where, err)
		}
	}
	key := canonicalPath(segments)
	v.questions = append(v.questions, declarationQuestion{key: key, parent: parent, references: references})
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

// Reject bindings that would overwrite one another or require the same values
// container to be both an object and an array.
func (tree *pathTree) insert(path []any) error {
	node := tree
	for _, segment := range path {
		if node.terminal {
			return fmt.Errorf("duplicate or overlapping values path")
		}
		if node.children == nil {
			node.children = map[any]*pathTree{}
		}
		container := byte(1)
		if _, index := segment.(int); index {
			container = 2
		}
		if node.container != 0 && node.container != container {
			return fmt.Errorf("values paths require incompatible object and array containers")
		}
		node.container = container
		if node.children[segment] == nil {
			node.children[segment] = &pathTree{}
		}
		node = node.children[segment]
	}
	if node.terminal || len(node.children) > 0 {
		return fmt.Errorf("duplicate or overlapping values path")
	}
	node.terminal = true
	return nil
}

// Resolve references after every field is known, allowing forward references.
// Parent visibility and show_if dependencies share one graph for cycle detection.
func (v *declarationValidator) checkConditions() error {
	known := map[string]bool{}
	for _, q := range v.questions {
		known[q.key] = true
	}
	edges := map[string][]string{}
	for _, q := range v.questions {
		if q.parent != "" {
			edges[q.key] = append(edges[q.key], q.parent)
		}
		for _, reference := range q.references {
			segments, err := parsePath(reference)
			if err != nil {
				return err
			}
			key := canonicalPath(segments)
			if known[key] {
				edges[q.key] = append(edges[q.key], key)
			} else if _, exists := readPath(v.defaults, segments); v.defaults != nil && !exists {
				v.warnings = append(v.warnings, Warning{Code: "UNKNOWN_CONDITION_REFERENCE", Message: "Condition path is absent from chart defaults and declared fields", Path: reference})
			}
		}
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(key string) error {
		if state[key] == 1 {
			return fmt.Errorf("question visibility conditions contain a cycle")
		}
		if state[key] == 2 {
			return nil
		}
		state[key] = 1
		for _, next := range edges[key] {
			if err := visit(next); err != nil {
				return err
			}
		}
		state[key] = 2
		return nil
	}
	for _, q := range v.questions {
		if err := visit(q.key); err != nil {
			return err
		}
	}
	return nil
}

func numeric(value any) (*big.Rat, bool) {
	var text string
	switch value := value.(type) {
	case int:
		text = strconv.Itoa(value)
	case int64:
		text = strconv.FormatInt(value, 10)
	case uint64:
		text = strconv.FormatUint(value, 10)
	case uint:
		text = strconv.FormatUint(uint64(value), 10)
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, false
		}
		text = strconv.FormatFloat(value, 'g', -1, 64)
	case json.Number:
		text = value.String()
	default:
		return nil, false
	}
	return new(big.Rat).SetString(text)
}
func isScalar(value any) bool {
	if value == nil {
		return true
	}
	switch value.(type) {
	case string, bool:
		return true
	}
	_, ok := numeric(value)
	return ok
}
func equalScalar(left, right any) bool {
	if a, ok := numeric(left); ok {
		b, ok := numeric(right)
		return ok && a.Cmp(b) == 0
	}
	switch left := left.(type) {
	case nil:
		return right == nil
	case string:
		value, ok := right.(string)
		return ok && left == value
	case bool:
		value, ok := right.(bool)
		return ok && left == value
	}
	return false
}

func scalarKey(value any) string {
	if number, ok := numeric(value); ok {
		return "number:" + number.RatString()
	}
	switch value := value.(type) {
	case nil:
		return "null"
	case string:
		return "string:" + value
	case bool:
		return "boolean:" + strconv.FormatBool(value)
	}
	return "unsupported"
}

func (v *declarationValidator) normalizeOptions(value any, where string) ([]any, error) {
	list, ok := value.([]any)
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("nonempty options are required")
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
				return nil, fmt.Errorf("option labels must be text")
			}
			var exists bool
			value, exists = object["value"]
			if !exists {
				return nil, fmt.Errorf("options require a label and value")
			}
			v.ignoreUnknown(object, words("label value"), fmt.Sprintf("%s[%d]", where, index))
		} else {
			label = fmt.Sprint(value)
			if value == nil {
				label = "null"
			}
		}
		if !isScalar(value) {
			return nil, fmt.Errorf("options must have scalar values")
		}
		key := scalarKey(value)
		if values[key] {
			return nil, fmt.Errorf("option values must be distinct")
		}
		values[key] = true
		out = append(out, map[string]any{propertyLabel: label, "value": value})
	}
	return out, nil
}

func validateAnswer(question map[string]any, value any) error {
	if value == nil || value == "" {
		return nil
	}
	kind := question["type"].(string)
	if textTypes[kind] {
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("default must be text")
		}
		length := new(big.Rat).SetInt64(int64(utf8.RuneCountInString(text)))
		if low, ok := numeric(question[propertyMinLength]); ok && length.Cmp(low) < 0 {
			return fmt.Errorf("default is shorter than min_length")
		}
		if high, ok := numeric(question[propertyMaxLength]); ok && length.Cmp(high) > 0 {
			return fmt.Errorf("default is longer than max_length")
		}
		if kind == "hostname" {
			if len(text) > 253 {
				return fmt.Errorf("invalid hostname")
			}
			for _, part := range strings.Split(text, ".") {
				if !hostnamePart.MatchString(part) {
					return fmt.Errorf("invalid hostname")
				}
			}
		}
		if kind == "ipaddr" && net.ParseIP(text) == nil {
			return fmt.Errorf("invalid IP address")
		}
		if kind == "cidr" {
			if _, _, err := net.ParseCIDR(text); err != nil {
				return fmt.Errorf("invalid CIDR: %w", err)
			}
		}
		if kind == "cron" && !validCron(text) {
			return fmt.Errorf("invalid five-field numeric cron")
		}
	}
	if kind == "boolean" {
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("default must be a boolean")
		}
	}
	if kind == questionInteger || kind == questionFloat {
		number, ok := numeric(value)
		if !ok || (kind == questionInteger && !number.IsInt()) {
			return fmt.Errorf("default must match the numeric field type")
		}
		if low, ok := numeric(question[propertyMin]); ok && number.Cmp(low) < 0 {
			return fmt.Errorf("default is below min")
		}
		if high, ok := numeric(question[propertyMax]); ok && number.Cmp(high) > 0 {
			return fmt.Errorf("default is above max")
		}
	}
	if kind == questionEnum || kind == "radio" {
		answers := []any{value}
		if question["multiple"] == true {
			var ok bool
			answers, ok = value.([]any)
			if !ok {
				return fmt.Errorf("multiple default must be a list")
			}
		}
		available := map[string]bool{}
		for _, option := range question["options"].([]any) {
			available[scalarKey(option.(map[string]any)["value"])] = true
		}
		seen := map[string]bool{}
		for _, answer := range answers {
			key := scalarKey(answer)
			if !isScalar(answer) || !available[key] {
				return fmt.Errorf("default must be an available option")
			}
			if seen[key] {
				return fmt.Errorf("multiple default must be distinct")
			}
			seen[key] = true
		}
	}
	if kind == "array[string]" {
		list, ok := value.([]any)
		if !ok {
			return fmt.Errorf("default must be a string list")
		}
		for _, item := range list {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("default list must contain strings")
			}
		}
	}
	if kind == "map[string]" {
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("default must be a string map")
		}
		for _, item := range object {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("default map must contain string values")
			}
		}
	}
	return nil
}

func validCron(value string) bool {
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return false
	}
	limits := [][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	for index, field := range fields {
		for _, piece := range strings.Split(field, ",") {
			parts := cronPart.FindStringSubmatch(piece)
			if parts == nil {
				return false
			}
			if parts[2] != "" {
				step, err := strconv.Atoi(parts[2])
				if err != nil || step < 1 {
					return false
				}
			}
			if parts[1] == "*" {
				continue
			}
			numbers := strings.Split(parts[1], "-")
			previous := -1
			for _, text := range numbers {
				number, err := strconv.Atoi(text)
				if err != nil || number < limits[index][0] || number > limits[index][1] || number < previous {
					return false
				}
				previous = number
			}
		}
	}
	return true
}
