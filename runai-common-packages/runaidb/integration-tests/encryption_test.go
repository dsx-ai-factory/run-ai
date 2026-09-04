/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
)

type TestSecret struct {
	ID     int    `gorm:"primaryKey;column:id"`
	Name   string `gorm:"column:name"`
	Secret string `gorm:"column:secret;serializer:encrypted"`
}

func (TestSecret) TableName() string {
	return "test_secrets"
}

func TestEncryption_ValueIsEncryptedInDB(t *testing.T) {
	plaintext := "my-super-secret-value"

	record := TestSecret{Name: "encryption-test", Secret: plaintext}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)
	require.NotZero(t, record.ID)

	var rawSecret string
	err = runaidb.DB().Raw(
		"SELECT secret FROM test_schema.test_secrets WHERE id = ?", record.ID,
	).Row().Scan(&rawSecret)
	require.NoError(t, err)
	require.NotEmpty(t, rawSecret)
	require.NotEqual(t, plaintext, rawSecret)
}

func TestEncryption_ValueIsDecryptedOnRead(t *testing.T) {
	plaintext := "another-secret-value"

	record := TestSecret{Name: "decryption-test", Secret: plaintext}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)
	require.NotZero(t, record.ID)

	var result TestSecret
	err = runaidb.DB().First(&result, record.ID)
	require.NoError(t, err)
	require.Equal(t, plaintext, result.Secret, "GORM should decrypt the value on read")
}
