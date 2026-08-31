/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package branding

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TRANSPARENT_PNG_BASE64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+V4nEAAAAASUVORK5CYII="

// getLogoIfValid

func TestGetLogoIfValid_HappyFlow(t *testing.T) {
	pngBytes, _ := base64.StdEncoding.DecodeString(TRANSPARENT_PNG_BASE64)

	result, err := getLogoIfValid(newLogoValidate(), TRANSPARENT_PNG_BASE64)

	require.Nil(t, err)
	assert.Equal(t, pngBytes, *result)
}

func TestGetLogoIfValid_FailBigSize(t *testing.T) {
	oversized := make([]byte, 128*1024+1)
	encoded := base64.StdEncoding.EncodeToString(oversized)

	result, err := getLogoIfValid(newLogoValidate(), encoded)

	assert.Nil(t, result)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusRequestEntityTooLarge, err.Code)
}

func TestGetLogoIfValid_FailMimeType(t *testing.T) {
	// JPEG bytes — valid base64 but wrong MIME type
	jpeg := append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, make([]byte, 50)...)
	encoded := base64.StdEncoding.EncodeToString(jpeg)

	result, err := getLogoIfValid(newLogoValidate(), encoded)

	assert.Nil(t, result)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, err.Code)
}

// DecodeLogo

func TestDecodeLogo_ValidRawBase64(t *testing.T) {
	lv := newLogoValidate()
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))

	result, err := lv.DecodeLogo(encoded)

	require.Nil(t, err)
	assert.Equal(t, []byte("hello"), *result)
}

func TestDecodeLogo_ValidDataURL(t *testing.T) {
	lv := newLogoValidate()
	raw := []byte("hello")
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)

	result, err := lv.DecodeLogo(dataURL)

	require.Nil(t, err)
	assert.Equal(t, raw, *result)
}

func TestDecodeLogo_InvalidDataURL_MissingComma(t *testing.T) {
	lv := newLogoValidate()

	result, err := lv.DecodeLogo("data:image/png;base64-no-comma")

	assert.Nil(t, result)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
}

func TestDecodeLogo_InvalidBase64(t *testing.T) {
	lv := newLogoValidate()

	result, err := lv.DecodeLogo("not-valid-base64!!!")

	assert.Nil(t, result)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
}

// ValidateLogoSize

func TestValidateLogoSize_WithinLimit(t *testing.T) {
	lv := newLogoValidate()
	logo := make([]byte, 128*1024) // exactly at limit

	err := lv.ValidateLogoSize(&logo)

	assert.Nil(t, err)
}

func TestValidateLogoSize_ExceedsLimit(t *testing.T) {
	lv := newLogoValidate()
	logo := make([]byte, 128*1024+1)

	err := lv.ValidateLogoSize(&logo)

	require.NotNil(t, err)
	assert.Equal(t, http.StatusRequestEntityTooLarge, err.Code)
}

// ValidateMimeType

func TestValidateMimeType_PNG(t *testing.T) {
	lv := newLogoValidate()
	pngBytes, decErr := base64.StdEncoding.DecodeString(TRANSPARENT_PNG_BASE64)
	require.NoError(t, decErr)

	err := lv.ValidateMimeType(&pngBytes)

	assert.Nil(t, err)
}

func TestValidateMimeType_SVG(t *testing.T) {
	lv := newLogoValidate()
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"></svg>`)

	err := lv.ValidateMimeType(&svg)

	assert.Nil(t, err)
}

func TestValidateMimeType_UnsupportedType(t *testing.T) {
	lv := newLogoValidate()
	// JPEG magic bytes
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	jpeg = append(jpeg, make([]byte, 50)...)

	err := lv.ValidateMimeType(&jpeg)

	require.NotNil(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, err.Code)
}

// isSVG

func TestIsSVG_ValidSVG(t *testing.T) {
	lv := &logoValidate{}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)

	assert.True(t, lv.isSVG(svg))
}

func TestIsSVG_NotSVG(t *testing.T) {
	lv := &logoValidate{}

	assert.False(t, lv.isSVG([]byte("just some text")))
}

func TestIsSVG_TruncatesTo512Bytes(t *testing.T) {
	lv := &logoValidate{}
	// SVG markers placed beyond 512 bytes — should not be detected
	padding := strings.Repeat("x", 520)
	notSVG := []byte(padding + `<svg xmlns="http://www.w3.org/2000/svg">`)

	assert.False(t, lv.isSVG(notSVG))
}
