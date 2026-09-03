/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkClientVersionLessThan(clusterVersion, version2 string) (bool, error) {
	return compareVersions(clusterVersion, version2, LESS)
}

func TestCantParseKubeMajorVersion(t *testing.T) {
	_, err := checkClientVersionLessThan("vbla.21.0", "1.21")
	assert.ErrorContains(t, err, "invalid version1 Malformed version")
}

func TestCantParseKubeMinorVersion(t *testing.T) {
	_, err := checkClientVersionLessThan("v1.bla.0", "1.21")
	assert.ErrorContains(t, err, "invalid version1 Malformed version")
}

func TestKubeMinor(t *testing.T) {
	isLess, err := checkClientVersionLessThan("v1.20.0", "1.21")
	assert.Nil(t, err)
	assert.True(t, isLess)
}

func TestKubeVersionExact(t *testing.T) {
	isLess, err := checkClientVersionLessThan("v1.21.0", "1.21")
	assert.Nil(t, err)
	assert.False(t, isLess)
}

func TestKubeVersionBigger(t *testing.T) {
	isLess, err := checkClientVersionLessThan("v1.24.0", "1.21")
	assert.Nil(t, err)
	assert.False(t, isLess)
}

func TestSpecialKubeVersion(t *testing.T) {
	isLess, err := checkClientVersionLessThan("v1.23.14-eks-ffeb93d", "1.21")
	assert.Nil(t, err)
	assert.False(t, isLess)
}

func TestSpecialVersionLessEqual(t *testing.T) {
	isLessEqual, err := VersionLessThan("1.21", "v1.21.14-eks-ffeb93d")
	assert.Nil(t, err)
	assert.True(t, isLessEqual)
}

func TestVersionGreaterEqualThan(t *testing.T) {
	isLess, err := VersionGreaterEqualThan("1.21", "v1.21.0")
	assert.Nil(t, err)
	assert.True(t, isLess)
}
