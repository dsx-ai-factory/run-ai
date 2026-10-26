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
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"k8s.io/apimachinery/pkg/util/validation"
)

const numCrontabFields = 5

var textTypes = datastructs.NewSet(strings.Fields("string multiline password storageclass pvc configmap hostname ipaddr cidr cron")...)
var hostnamePart = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var cronPart = regexp.MustCompile(`^(\*|\d+(?:-\d+)?)(?:/(\d+))?$`)

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

func validateAnswer(question map[string]any, value any) *api.Error {
	if value == nil || value == "" {
		return nil
	}
	kind := question["type"].(string)
	if textTypes.Has(kind) {
		text, ok := value.(string)
		if !ok {
			return api.NewValidationError("default must be text")
		}
		length := new(big.Rat).SetInt64(int64(utf8.RuneCountInString(text)))
		if low, ok := numeric(question[propertyMinLength]); ok && length.Cmp(low) < 0 {
			return api.NewValidationError("default is shorter than min_length")
		}
		if high, ok := numeric(question[propertyMaxLength]); ok && length.Cmp(high) > 0 {
			return api.NewValidationError("default is longer than max_length")
		}
		if kind == "hostname" {
			if len(text) > validation.DNS1123SubdomainMaxLength {
				return api.NewValidationError(fmt.Sprintf("hostname exceeds %d bytes", validation.DNS1123SubdomainMaxLength))
			}
			for _, part := range strings.Split(text, ".") {
				if !hostnamePart.MatchString(part) {
					return api.NewValidationError("invalid hostname")
				}
			}
		}
		if kind == "ipaddr" && net.ParseIP(text) == nil {
			return api.NewValidationError("invalid IP address")
		}
		if kind == "cidr" {
			if _, _, err := net.ParseCIDR(text); err != nil {
				return api.NewValidationError(fmt.Sprintf("invalid CIDR: %v", err))
			}
		}
		if kind == "cron" && !validCron(text) {
			return api.NewValidationError("invalid five-field numeric cron")
		}
	}
	if kind == "boolean" {
		if _, ok := value.(bool); !ok {
			return api.NewValidationError("default must be a boolean")
		}
	}
	if kind == questionInteger || kind == questionFloat {
		number, ok := numeric(value)
		if !ok || (kind == questionInteger && !number.IsInt()) {
			return api.NewValidationError("default must match the numeric field type")
		}
		if low, ok := numeric(question[propertyMin]); ok && number.Cmp(low) < 0 {
			return api.NewValidationError(fmt.Sprintf("default is below min (%v)", question[propertyMin]))
		}
		if high, ok := numeric(question[propertyMax]); ok && number.Cmp(high) > 0 {
			return api.NewValidationError(fmt.Sprintf("default is above max (%v)", question[propertyMax]))
		}
	}
	if kind == questionEnum || kind == "radio" {
		answers := []any{value}
		if question["multiple"] == true {
			var ok bool
			answers, ok = value.([]any)
			if !ok {
				return api.NewValidationError("multiple default must be a list")
			}
		}
		available := datastructs.NewSet[string]()
		for _, option := range question["options"].([]any) {
			available.Add(scalarKey(option.(map[string]any)["value"]))
		}
		seen := datastructs.NewSet[string]()
		for _, answer := range answers {
			key := scalarKey(answer)
			if !isScalar(answer) || !available.Has(key) {
				return api.NewValidationError("default must be an available option")
			}
			if seen.Has(key) {
				return api.NewValidationError("multiple default must be distinct")
			}
			seen.Add(key)
		}
	}
	if kind == "array[string]" {
		list, ok := value.([]any)
		if !ok {
			return api.NewValidationError("default must be a string list")
		}
		for _, item := range list {
			if _, ok := item.(string); !ok {
				return api.NewValidationError("default list must contain strings")
			}
		}
	}
	if kind == "map[string]" {
		object, ok := value.(map[string]any)
		if !ok {
			return api.NewValidationError("default must be a string map")
		}
		for _, item := range object {
			if _, ok := item.(string); !ok {
				return api.NewValidationError("default map must contain string values")
			}
		}
	}
	return nil
}

func validCron(value string) bool {
	fields := strings.Fields(value)
	if len(fields) != numCrontabFields {
		return false
	}
	// Minute, hour, day of month, month, day of week; Sunday accepts both 0 and 7.
	limits := [numCrontabFields][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
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
