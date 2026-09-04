/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubeconfig

import (
	"fmt"
	"path/filepath"

	"github.com/siderolabs/go-kubeconfig"
)

// GetConfigRealPath if the config is just a symbolic link, we want to have the real path when we write the changes
func GetConfigRealPath() (string, error) {
	var realPath string
	path, err := kubeconfig.SinglePath()
	if err != nil {
		return "", err
	}
	realPath, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return realPath, err
}

func GetConfigMerger(path string) (*kubeconfig.Merger, error) {
	merger, err := kubeconfig.Load(path)
	if err != nil {
		return nil, err
	}
	if merger == nil {
		return nil, fmt.Errorf("failed to load the kubeconfig file: %s", path)
	}
	return merger, err
}
