/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package endpoints

import (
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validConfig() *Config {
	return &Config{
		SchemaVersion: SchemaVersionV1,
		Endpoints: []Endpoint{{
			Services: []Service{{
				Create: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "svc"},
					Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 80}}},
				},
			}},
		}},
	}
}

func TestGet_RejectsEmptyEndpoints(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{
		Annotation: `{"schemaVersion":"v1","endpoints":[]}`,
	}}
	_, err := Get(obj)
	if !errors.Is(err, ErrNoEndpoints) {
		t.Fatalf("expected ErrNoEndpoints, got %v", err)
	}
}

func TestGet_RejectsMissingEndpoints(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{
		Annotation: `{"schemaVersion":"v1"}`,
	}}
	_, err := Get(obj)
	if !errors.Is(err, ErrNoEndpoints) {
		t.Fatalf("expected ErrNoEndpoints, got %v", err)
	}
}

func TestGet_RejectsBadSchemaVersion(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{
		Annotation: `{"schemaVersion":"v9","endpoints":[{"services":[{"ref":{"name":"x","port":80}}]}]}`,
	}}
	_, err := Get(obj)
	if !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf("expected ErrUnsupportedSchemaVersion, got %v", err)
	}
}

func TestGet_AbsentAnnotationReturnsNilNil(t *testing.T) {
	obj := &metav1.ObjectMeta{}
	cfg, err := Get(obj)
	if err != nil || cfg != nil {
		t.Fatalf("expected (nil, nil) for absent annotation, got (%v, %v)", cfg, err)
	}
}

func TestSet_RejectsEmptyEndpoints(t *testing.T) {
	obj := &metav1.ObjectMeta{}
	err := Set(obj, &Config{SchemaVersion: SchemaVersionV1})
	if !errors.Is(err, ErrNoEndpoints) {
		t.Fatalf("expected ErrNoEndpoints, got %v", err)
	}
}

func TestSet_RejectsNilConfig(t *testing.T) {
	obj := &metav1.ObjectMeta{}
	err := Set(obj, nil)
	if err == nil || !strings.Contains(err.Error(), "nil Config") {
		t.Fatalf("expected nil-Config error, got %v", err)
	}
}

func TestSetGetRoundTrip(t *testing.T) {
	obj := &metav1.ObjectMeta{}
	if err := Set(obj, validConfig()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	cfg, err := Get(obj)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cfg.SchemaVersion != SchemaVersionV1 {
		t.Fatalf("schemaVersion: got %q, want %q", cfg.SchemaVersion, SchemaVersionV1)
	}
	if len(cfg.Endpoints) != 1 || len(cfg.Endpoints[0].Services) != 1 {
		t.Fatalf("unexpected shape after round-trip: %+v", cfg)
	}
}
