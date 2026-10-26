/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_actions

import (
	"context"
	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestBackupRunaiConfigWithoutBackup(t *testing.T) {
	runaiConfig := DefaultRunaiConfig()
	runaiConfig.Annotations = map[string]string{
		"test": "test",
	}

	k8s := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(runaiConfig).Build()

	err := BackupRunaiConfig(k8s)
	assert.Equal(t, err, nil)

	cm := runaiConfigBackupConfigMap()
	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(cm), cm)
	assert.Equal(t, err, nil)

	backupRunaiConfig := &runaiv1.RunaiConfig{}

	err = yaml.Unmarshal([]byte(cm.Data[runaiConfigKey]), backupRunaiConfig)
	assert.Equal(t, err, nil)

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(runaiConfig), runaiConfig)
	assert.Equal(t, err, nil)

	assert.EqualValues(t, reflect.DeepEqual(runaiConfig, backupRunaiConfig), true)
}

func TestBackupRunaiConfigWithBackup(t *testing.T) {
	runaiConfig := DefaultRunaiConfig()
	runaiConfig.Annotations = map[string]string{
		"test": "test",
	}

	cm := runaiConfigBackupConfigMap()

	k8s := fake.NewClientBuilder().
		WithScheme(schemeinit.Scheme()).
		WithObjects(runaiConfig).
		WithObjects(cm).
		Build()

	err := BackupRunaiConfig(k8s)
	assert.Equal(t, err, nil)

	cm = runaiConfigBackupConfigMap()

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(cm), cm)
	assert.Equal(t, err, nil)

	backupRunaiConfig := &runaiv1.RunaiConfig{}

	err = yaml.Unmarshal([]byte(cm.Data[runaiConfigKey]), backupRunaiConfig)
	assert.Equal(t, err, nil)

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(runaiConfig), runaiConfig)
	assert.Equal(t, err, nil)

	assert.EqualValues(t, reflect.DeepEqual(runaiConfig, backupRunaiConfig), true)
}
