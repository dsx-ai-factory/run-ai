/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package personalcreds is the shared library for the run.ai/personal-credentials
// annotation contract. The CP composer writes the annotation, the
// agent forwards it, and the anyworkload-controller plugin acts on
// it — all using the helpers here so the contract can't drift across components.
//
// See designs/personal-credentials-plugin/.
package personalcreds

import (
	"errors"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const Annotation = "run.ai/personal-credentials"

var (
	ErrEmptyAnnotation   = errors.New("annotation is empty")
	ErrNoEntries         = errors.New("annotation must list at least one entry")
	ErrSourceSecretEmpty = errors.New("sourceSecret is required")
)

type Cred struct {
	SourceSecret string `json:"sourceSecret"`
}

// Parse reads a raw annotation value into entries. Rejects malformed YAML,
// empty lists, and entries with empty SourceSecret.
func Parse(raw string) ([]Cred, error) {
	if raw == "" {
		return nil, ErrEmptyAnnotation
	}
	var entries []Cred
	if err := yaml.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}
	if len(entries) == 0 {
		return nil, ErrNoEntries
	}
	for i, e := range entries {
		if e.SourceSecret == "" {
			return nil, fmt.Errorf("entry %d: %w", i, ErrSourceSecretEmpty)
		}
	}
	return entries, nil
}

// Marshal renders entries to a stable YAML string suitable for use as an
// annotation value. Entries are sorted by SourceSecret so updates are
// diff-friendly.
func Marshal(entries []Cred) (string, error) {
	if err := validate(entries); err != nil {
		return "", err
	}
	sorted := append([]Cred(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SourceSecret < sorted[j].SourceSecret
	})
	out, err := yaml.Marshal(sorted)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return string(out), nil
}

// Has reports whether obj carries the personal-credentials annotation.
func Has(obj metav1.Object) bool {
	_, ok := obj.GetAnnotations()[Annotation]
	return ok
}

// Get returns the parsed entries from obj's annotation, or (nil, nil) if the
// annotation is absent. Returns an error only on parse failure.
func Get(obj metav1.Object) ([]Cred, error) {
	raw, ok := obj.GetAnnotations()[Annotation]
	if !ok {
		return nil, nil
	}
	return Parse(raw)
}

// Set replaces obj's personal-credentials annotation. Removes the annotation
// entirely when entries is empty. Validates each entry.
func Set(obj metav1.Object, entries []Cred) error {
	if len(entries) == 0 {
		ann := obj.GetAnnotations()
		if ann == nil {
			return nil
		}
		delete(ann, Annotation)
		obj.SetAnnotations(ann)
		return nil
	}
	raw, err := Marshal(entries)
	if err != nil {
		return err
	}
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[Annotation] = raw
	obj.SetAnnotations(ann)
	return nil
}

// Add appends entry to obj's existing entries. Idempotent — a duplicate
// SourceSecret is skipped silently.
func Add(obj metav1.Object, cred Cred) error {
	if cred.SourceSecret == "" {
		return ErrSourceSecretEmpty
	}
	current, err := Get(obj)
	if err != nil {
		return err
	}
	for _, e := range current {
		if e.SourceSecret == cred.SourceSecret {
			return nil
		}
	}
	return Set(obj, append(current, cred))
}

// Remove drops any entry matching sourceSecret. Returns whether anything was
// removed. Removes the annotation entirely when no entries remain.
func Remove(obj metav1.Object, sourceSecret string) (bool, error) {
	current, err := Get(obj)
	if err != nil {
		return false, err
	}
	filtered := current[:0]
	for _, e := range current {
		if e.SourceSecret != sourceSecret {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == len(current) {
		return false, nil
	}
	return true, Set(obj, filtered)
}

func validate(entries []Cred) error {
	if len(entries) == 0 {
		return ErrNoEntries
	}
	for i, e := range entries {
		if e.SourceSecret == "" {
			return fmt.Errorf("entry %d: %w", i, ErrSourceSecretEmpty)
		}
	}
	return nil
}
