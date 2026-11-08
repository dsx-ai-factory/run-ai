/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpuinventory

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadDevices associates kernel PCI identities with the NVIDIA driver's physical
// GPUs. Reading files works without allocating GPUs or loading host driver libraries.
func ReadDevices(sysRoot, procRoot string) ([]Device, error) {
	entries, err := os.ReadDir(filepath.Join(sysRoot, "bus/pci/devices"))
	if err != nil {
		return nil, fmt.Errorf("list PCI devices: %w", err)
	}
	devices := []Device{}
	for _, entry := range entries {
		bus := entry.Name()
		if !busPattern.MatchString(bus) {
			return nil, fmt.Errorf("invalid PCI address %q", bus)
		}
		path := filepath.Join(sysRoot, "bus/pci/devices", bus)
		vendor, err := readHex(filepath.Join(path, "vendor"), 16)
		if err != nil {
			return nil, err
		}
		if vendor != 0x10de {
			continue
		}
		class, err := readHex(filepath.Join(path, "class"), 24)
		if err != nil {
			return nil, err
		}
		if class>>16 != 3 {
			continue
		}
		id, err := readHex(filepath.Join(path, "device"), 16)
		if err != nil {
			return nil, err
		}
		if id == 0 {
			return nil, fmt.Errorf("GPU %s has zero PCI device ID", bus)
		}
		info, err := os.ReadFile(filepath.Join(procRoot, bus, "information")) // #nosec G304 -- Fixed driver mount, validated PCI address and fixed attribute.
		if err != nil {
			return nil, fmt.Errorf("read NVIDIA GPU %s: %w", bus, err)
		}
		fields := map[string]string{}
		for _, line := range strings.Split(string(info), "\n") {
			key, value, ok := strings.Cut(line, ":")
			if ok {
				fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
		product := strings.Join(strings.Fields(fields["Model"]), "-")
		uuid := fields["GPU UUID"]
		if fields["Bus Location"] != bus || product == "" || !strings.HasPrefix(uuid, "GPU-") || len(uuid) <= 4 || strings.EqualFold(fields["GPU Excluded"], "Yes") {
			return nil, fmt.Errorf("incomplete or inconsistent NVIDIA GPU %s identity", bus)
		}
		devices = append(devices, Device{UUID: uuid, BusID: bus, PCIDeviceID: fmt.Sprintf("%04x:%04x", id, vendor), Product: product})
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no physical NVIDIA GPUs discovered")
	}
	return devices, nil
}

func readHex(path string, bits int) (uint64, error) {
	b, err := os.ReadFile(path) // #nosec G304 -- Caller constructs only vendor/device/class paths below the fixed PCI mount.
	if err != nil {
		return 0, fmt.Errorf("read PCI attribute %s: %w", path, err)
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(b)), 0, bits)
	if err != nil {
		return 0, fmt.Errorf("parse PCI attribute %s: %w", path, err)
	}
	return value, nil
}
