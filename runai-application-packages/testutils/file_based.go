/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testutils

import (
	"errors"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
)

const (
	NA = ""

	DefaultInputDir                 = "testing/inputs"
	DefaultExpectedDir              = "testing/expected"
	DefaultClusterVersionForTesting = "9.9.9"
)

func LoadInputFile(testName string, basePath string) ([]byte, error) {
	return LoadInputFileFrom(testName, basePath, DefaultInputDir)
}

func LoadInputFileFrom(testName string, basePath string, dir string) ([]byte, error) {

	inputPath := fmt.Sprintf("%s/%s/%s.yaml", basePath, dir, testName)

	//
	//   read the Workload yaml from the inputs directory
	//
	var err error
	var inputYaml []byte

	if _, err = os.Stat(inputPath); err != nil {
		gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("expecting to find input file %s", inputPath))
		return nil, err
	}

	if inputYaml, err = os.ReadFile(inputPath); err != nil { //nolint:gosec // G304: test helper reads a caller-provided path
		gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("failed to read input file %s", inputPath))
		return nil, err
	}

	return inputYaml, nil
}

func LoadExpectedFile(testName string, actualYaml []byte, basePath string) []byte {
	return LoadExpectedFileFrom(testName, actualYaml, basePath, DefaultExpectedDir)
}

func LoadExpectedFileFrom(testName string, actualYaml []byte, basePath string, dir string) []byte {
	expectedPath := fmt.Sprintf("%s/%s/%s.yaml", basePath, dir, testName)

	if _, err := os.Stat(expectedPath); errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(expectedPath, actualYaml, 0o600)
		if err != nil {
			assert.Fail(ginkgo.GinkgoT(), fmt.Sprintf("did not find expected file %s and failed to create it: %v", expectedPath, err))
		} else {
			assert.Fail(ginkgo.GinkgoT(), fmt.Sprintf("expected file %s not found, created it. make sure to review it!", expectedPath))
		}
	}

	expectedYaml, err := os.ReadFile(expectedPath) //nolint:gosec // G304: test helper reads a caller-provided path
	gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("could not read expected file %s: %v", expectedPath, err))

	return expectedYaml
}
