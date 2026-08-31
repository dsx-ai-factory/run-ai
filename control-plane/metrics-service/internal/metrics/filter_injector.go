/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
)

const (
	generalInjectionPlaceholder = "FILTERS"
	specificInjectionPrefix     = "FILTER::"
	specificInjectionSuffix     = "::FILTER"
)

// validLabelNameRegexp matches valid PromQL label names: [a-zA-Z_][a-zA-Z0-9_]*
var validLabelNameRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type simpleFilterInjector struct {
	specificInjectionRegexp *regexp.Regexp
	next                    QueryInjector
}

func NewSimpleFilterInjector(next QueryInjector) *simpleFilterInjector {
	return &simpleFilterInjector{
		specificInjectionRegexp: regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, specificInjectionPrefix, specificInjectionSuffix)),
		next:                    next,
	}
}

func (s *simpleFilterInjector) Inject(query string, metricInformation *services.MetricInformation) (string, error) {
	specificInjectedQuery, err := s.injectSpecificFilters(query, metricInformation.Filters)
	if err != nil {
		return "", err
	}
	injectedQuery, err := s.injectGeneralFilters(specificInjectedQuery, metricInformation.Filters)
	if err != nil {
		return "", err
	}
	if s.next == nil {
		return injectedQuery, nil
	}
	return s.next.Inject(injectedQuery, metricInformation)
}

func (s *simpleFilterInjector) injectSpecificFilters(query string, filters map[string][]string) (string, error) {
	matches := s.specificInjectionRegexp.FindAllStringSubmatch(query, -1)
	var err error
	for _, match := range matches {
		query, err = injectSpecificFilterIntoQuery(query, match[1], match[0], filters)
		if err != nil {
			return "", fmt.Errorf("could not inject specific filters: %v", err)
		}
	}
	return query, nil
}

func injectSpecificFilterIntoQuery(query string, filter string, filterPlaceholder string, filters map[string][]string) (string, error) {
	if !validLabelNameRegexp.MatchString(filter) {
		return "", fmt.Errorf("invalid filter key %q: must match [a-zA-Z_][a-zA-Z0-9_]*", filter)
	}
	if values, ok := filters[filter]; ok && len(values) > 0 {
		escaped := escapePromQLValues(values)
		return strings.ReplaceAll(query, filterPlaceholder, fmt.Sprintf("%s=~'%s'", filter, strings.Join(escaped, "|"))), nil
	}
	return "", fmt.Errorf("filter %s not found", filter)
}

func (s *simpleFilterInjector) injectGeneralFilters(query string, filters map[string][]string) (string, error) {
	filterStr, err := buildFiltersStr(filters)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(query, generalInjectionPlaceholder, filterStr), nil
}

func buildFiltersStr(filters map[string][]string) (string, error) {
	stringFilters := []string{}
	for key, values := range filters {
		if !validLabelNameRegexp.MatchString(key) {
			return "", fmt.Errorf("invalid filter key %q: must match [a-zA-Z_][a-zA-Z0-9_]*", key)
		}
		if len(values) > 0 {
			escaped := escapePromQLValues(values)
			stringFilters = append(stringFilters, fmt.Sprintf("%s=~'%s'", key, strings.Join(escaped, "|")))
		}
	}
	return strings.Join(stringFilters, ","), nil
}

// escapePromQLValues escapes each value so it cannot break out of a PromQL single-quoted string literal.
func escapePromQLValues(values []string) []string {
	escaped := make([]string, len(values))
	for i, v := range values {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `'`, `\'`)
		escaped[i] = v
	}
	return escaped
}
