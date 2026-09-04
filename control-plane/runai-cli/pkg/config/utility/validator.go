/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utility

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"

	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
)

const (
	AlphanumericOrHyphenRegex = "^[a-zA-Z0-9-]+$"
	MaxProjectLength          = 63
)

var regex = regexp.MustCompile(AlphanumericOrHyphenRegex)

func SanityCheck(cpUrl, authUrl string) (string, error) {
	var err error
	if cpUrl != "" {
		if err = ValidateUrl(cpUrl); err != nil {
			return "control plane URL", err
		}
	}
	if authUrl != "" {
		if err = ValidateUrl(authUrl); err != nil {
			return "authentication URL", err
		}
	}
	return "", nil
}

func ValidateUrl(urlStr string) error {
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return clierrors.UserEnrichment(err)
	}
	return nil
}

func ValidateProjectName(project string) error {
	if len(project) > MaxProjectLength {
		return fmt.Errorf("string contain more then 63 characters")
	}
	if !regex.MatchString(project) {
		return fmt.Errorf("string contain other characters other then alphanumeric or hyphen")
	}
	return nil
}

func ValidateDirectory(dirPath string) error {
	// Check if input is empty
	if dirPath == "" {
		return errors.New("directory path cannot be empty")
	}

	// Check if path exists
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("'%s' does not exist", dirPath)
		}
		return fmt.Errorf("error accessing '%s': %v", dirPath, err)
	}

	// Check if path is a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", dirPath)
	}

	// Check if directory is writable by creating and removing a temporary file
	testFile := dirPath + "/.write_test_" + fmt.Sprintf("%d", os.Getpid())
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory '%s' is not writable", dirPath)
	}
	// Close and remove the test file
	defer os.Remove(testFile)
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
