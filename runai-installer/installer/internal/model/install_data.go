/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
)

type InstallValues struct {
	Prerequisites []string                                 `json:"prerequisites"`
	Inputs        map[configinputs.InstallInputType]string `json:"inputs"`
}

type RunAIInstallData struct {
	InstallValues
	UserConfirmation bool
	SessionID        string
	Config           *config.InstallConfig
}

func NewRunAIInstallData(opts *config.InstallConfig) (*RunAIInstallData, error) {
	runaiData := &RunAIInstallData{
		Config: opts,
	}
	runaiData.Inputs = make(map[configinputs.InstallInputType]string)
	err := initConfigInputsMap(opts, runaiData)
	if err != nil {
		return nil, err
	}
	return runaiData, nil
}

func initConfigInputsMap(opts *config.InstallConfig, runaiData *RunAIInstallData) error {
	if len(opts.ConfigValues) > 0 {
		for key, value := range opts.ConfigValues {
			if _, exists := configinputs.FlagToInstallInput[key]; !exists {
				return fmt.Errorf("unknown config value key: %s", key)
			} else {
				runaiData.Inputs[configinputs.FlagToInstallInput[key]] = value
			}
		}
	}
	return nil
}

func (r *RunAIInstallData) GetInputValue(inputType configinputs.InstallInputType) string {
	return r.Inputs[inputType]
}

func (r *RunAIInstallData) GetGPUType() config.GPUType {
	return r.Config.GPUType
}
