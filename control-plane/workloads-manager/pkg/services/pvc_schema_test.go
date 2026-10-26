// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPVCConfigurationIsPublic(t *testing.T) {
	type schema struct {
		Internal   bool `json:"x-internal"`
		Type       string
		Nullable   bool
		Properties map[string]schema
	}
	var spec struct {
		Components struct{ Schemas map[string]schema }
	}
	data, err := os.ReadFile("../../swagger/_build/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"pvcs", "pvcsConfig"} {
		t.Run(name, func(t *testing.T) {
			field, exists := spec.Components.Schemas["WorkloadV2Configuration"].Properties[name]
			if !exists || field.Internal {
				t.Fatalf("WorkloadV2Configuration.%s must be public", name)
			}
		})
	}
	volumeMode := spec.Components.Schemas["WorkloadV2UpdateConfigurationPvcSpec"].Properties["volumeMode"]
	if volumeMode.Type != "string" || !volumeMode.Nullable {
		t.Fatal("volumeMode must remain a nullable string")
	}
}
