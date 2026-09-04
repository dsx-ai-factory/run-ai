/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestAddSingleEnvVar(t *testing.T) {
	container := &v1.Container{}
	AddEnvVarToContainer(container, v1.EnvVar{
		Name:  "myVar",
		Value: "myValue",
	})

	assert.Equal(t, len(container.Env), 1)
	env := container.Env[0]
	assert.Equal(t, env.Name, "myVar")
	assert.Equal(t, env.Value, "myValue")
}

func TestAddMultipleEnvVars(t *testing.T) {
	container := &v1.Container{}
	AddEnvVarToContainer(container, v1.EnvVar{
		Name:  "myVar1",
		Value: "myValue1",
	})

	AddEnvVarToContainer(container, v1.EnvVar{
		Name:  "myVar2",
		Value: "myValue2",
	})

	assert.Equal(t, len(container.Env), 2)
	assert.Equal(t, container.Env[0].Name, "myVar1")
	assert.Equal(t, container.Env[0].Value, "myValue1")
	assert.Equal(t, container.Env[1].Name, "myVar2")
	assert.Equal(t, container.Env[1].Value, "myValue2")
}

func TestOverrideVar(t *testing.T) {
	container := &v1.Container{
		Env: []v1.EnvVar{
			{
				Name:  "myVar",
				Value: "myValue",
			},
		},
	}
	AddEnvVarToContainer(container, v1.EnvVar{
		Name:  "myVar",
		Value: "myValue1",
	})

	assert.Equal(t, len(container.Env), 1)
	assert.Equal(t, container.Env[0].Name, "myVar")
	assert.Equal(t, container.Env[0].Value, "myValue1")
}
