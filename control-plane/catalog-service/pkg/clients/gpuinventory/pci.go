// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package gpuinventory

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/nvidia-gpus.json data/LICENSE.nvidia-driver-assistant
var gpuSnapshot embed.FS

type gpuPCISnapshot struct {
	Devices map[string][]string `json:"devices"`
}

var gpuPCIDevices = func() map[string][]string {
	var snapshot gpuPCISnapshot
	data, err := gpuSnapshot.ReadFile("data/nvidia-gpus.json")
	if err != nil {
		panic(fmt.Sprintf("read embedded NVIDIA GPU snapshot: %v", err))
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		panic(fmt.Sprintf("invalid embedded NVIDIA GPU snapshot: %v", err))
	}
	return snapshot.Devices
}()

// ResolvePCIDevice matches the exact GPU Feature Discovery product label.
// Names with multiple device IDs cannot establish PCI identity. In particular,
// memory capacity or a matching NGC profile must not break an identity tie.
func ResolvePCIDevice(gpuType string) (string, error) {
	ids := gpuPCIDevices[gpuType]
	if len(ids) == 0 {
		return "", fmt.Errorf("GPU type %q has no PCI identity in the NVIDIA snapshot", gpuType)
	}
	if len(ids) != 1 {
		return "", fmt.Errorf("GPU type %q maps to multiple NVIDIA PCI identities", gpuType)
	}
	return ids[0], nil
}
