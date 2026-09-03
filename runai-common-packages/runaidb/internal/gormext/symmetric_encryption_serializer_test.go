/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gormext

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

type testModel struct {
	Secret string
}

func newTestSerializer(t *testing.T) *SymmetricEncryptionSerializer {
	key := "abcdefghijklmnopqrstuvwxyz012345" // 32 bytes
	s, err := NewSymmetricEncryptionSerializer(key)
	require.NoError(t, err)
	return s
}

func parseTestField(t *testing.T) *schema.Field {
	s, err := schema.Parse(&testModel{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field := s.LookUpField("Secret")
	require.NotNil(t, field)
	return field
}

func TestScan_NilDbValue_DoesNotModifyDst(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{Secret: "unchanged"}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, nil)
	assert.NoError(t, err)
	assert.Equal(t, "unchanged", model.Secret)
}

func TestScan_EmptyString_DoesNotModifyDst(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{Secret: "unchanged"}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, "")
	assert.NoError(t, err)
	assert.Equal(t, "unchanged", model.Secret)
}

func TestScan_EmptyByteSlice_DoesNotModifyDst(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{Secret: "unchanged"}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, []byte(""))
	assert.NoError(t, err)
	assert.Equal(t, "unchanged", model.Secret)
}

func TestScan_StringDbValue_Decrypts(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	plaintext := "my-secret-data"
	encrypted, err := serializer.encrypt(plaintext)
	require.NoError(t, err)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()

	err = serializer.Scan(context.Background(), field, dst, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, model.Secret)
}

func TestScan_ByteSliceDbValue_Decrypts(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	plaintext := "byte-secret"
	encrypted, err := serializer.encrypt(plaintext)
	require.NoError(t, err)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()

	err = serializer.Scan(context.Background(), field, dst, []byte(encrypted))
	assert.NoError(t, err)
	assert.Equal(t, plaintext, model.Secret)
}

func TestScan_UnsupportedType_ReturnsError(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, 12345)
	assert.Error(t, err)
	assert.EqualError(t, err, "unsupported data type for symmetric decryption: 12345")
}

func TestScan_InvalidBase64_ReturnsError(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, "not-valid-base64!!!")
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to decrypt field Secret: failed to decode base64: illegal base64 data at input byte 3")
}

func TestScan_CorruptedCiphertext_ReturnsError(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()

	err := serializer.Scan(context.Background(), field, dst, "dGFtcGVyZWQ=")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
}

func TestValue_NilFieldValue_ReturnsNil(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)
	dst := reflect.ValueOf(testModel{})

	result, err := serializer.Value(context.Background(), field, dst, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestValue_EmptyString_ReturnsEmptyString(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)
	dst := reflect.ValueOf(testModel{})

	result, err := serializer.Value(context.Background(), field, dst, "")
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestValue_NonEmptyString_Encrypts(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)
	dst := reflect.ValueOf(testModel{})

	plaintext := "my-secret-data"
	result, err := serializer.Value(context.Background(), field, dst, plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, result)

	decrypted, err := serializer.decrypt(result.(string))
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestValue_UnsupportedType_ReturnsError(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)
	dst := reflect.ValueOf(testModel{})

	result, err := serializer.Value(context.Background(), field, dst, 12345)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported data type")
}

func TestValue_SameInputTwice_ProducesDifferentCiphertext(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)
	dst := reflect.ValueOf(testModel{})

	plaintext := "determinism-check"
	result1, err := serializer.Value(context.Background(), field, dst, plaintext)
	assert.NoError(t, err)

	result2, err := serializer.Value(context.Background(), field, dst, plaintext)
	assert.NoError(t, err)

	assert.NotEqual(t, result1, result2, "same plaintext must produce different ciphertext (unique nonce)")
}

func TestValue_ThenScan_Roundtrip(t *testing.T) {
	serializer := newTestSerializer(t)
	field := parseTestField(t)

	plaintext := "roundtrip-test"
	encrypted, err := serializer.Value(context.Background(), field, reflect.ValueOf(testModel{}), plaintext)
	assert.NoError(t, err)

	model := testModel{}
	dst := reflect.ValueOf(&model).Elem()
	err = serializer.Scan(context.Background(), field, dst, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, model.Secret)
}

func TestEncrypt_ThenDecrypt_Roundtrip(t *testing.T) {
	serializer := newTestSerializer(t)

	plaintext := "my-secret-data"
	encrypted, err := serializer.encrypt(plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, encrypted)

	decrypted, err := serializer.decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}
