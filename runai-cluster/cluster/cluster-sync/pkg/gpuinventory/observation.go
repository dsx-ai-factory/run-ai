/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package gpuinventory collects physical GPU identity independently of inventory consumers.
package gpuinventory

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const (
	Annotation         = "run.ai/gpu-pci-devices"
	CountLabel         = "nvidia.com/gpu.count"
	ProductLabel       = "nvidia.com/gpu.product"
	MemoryLabel        = "nvidia.com/gpu.memory"
	RefreshInterval    = time.Minute
	MaxAge             = 5 * time.Minute
	ClockSkew          = time.Minute
	observationVersion = 2
)

var (
	busPattern = regexp.MustCompile(`^[0-9a-f]{4}:[0-9a-f]{2}:[01][0-9a-f]\.[0-7]$`)
	pciPattern = regexp.MustCompile(`^[0-9a-f]{4}:10de$`)
)

type Device struct {
	UUID        string `json:"uuid"`
	BusID       string `json:"busId"`
	PCIDeviceID string `json:"pciDeviceId"`
	Product     string `json:"product"`
}

// Observation binds hardware identity to the existing vendor-label inventory
// snapshot. PCI sysfs does not report GPU memory capacity.
type Observation struct {
	Version    int       `json:"version"`
	NodeUID    string    `json:"nodeUID"`
	BootID     string    `json:"bootID"`
	ObservedAt time.Time `json:"observedAt"`
	GPUProduct string    `json:"gpuProduct"`
	GPUMemory  string    `json:"gpuMemory"`
	GPUCount   string    `json:"gpuCount"`
	Devices    []Device  `json:"devices"`
}

func NewObservation(node *corev1.Node, devices []Device, now time.Time) Observation {
	return Observation{
		Version: observationVersion, NodeUID: string(node.UID), BootID: node.Status.NodeInfo.BootID,
		ObservedAt: now.UTC(), GPUProduct: node.Labels[ProductLabel], GPUMemory: node.Labels[MemoryLabel],
		GPUCount: node.Labels[CountLabel], Devices: devices,
	}
}

func ReportedIdentity(node *corev1.Node, now time.Time) string {
	observation := ReportedObservation(node, now)
	if observation == nil {
		return ""
	}
	return observation.Devices[0].PCIDeviceID
}

func ReportedObservation(node *corev1.Node, now time.Time) *Observation {
	var observation Observation
	if json.Unmarshal([]byte(node.Annotations[Annotation]), &observation) != nil || observation.Identity(node, now) == "" {
		return nil
	}
	return &observation
}

func (o Observation) Identity(node *corev1.Node, now time.Time) string {
	if o.Version != observationVersion || o.NodeUID == "" || o.NodeUID != string(node.UID) ||
		o.BootID == "" || o.BootID != node.Status.NodeInfo.BootID || o.ObservedAt.IsZero() ||
		!now.Before(o.ObservedAt.Add(MaxAge)) || o.ObservedAt.After(now.Add(ClockSkew)) {
		return ""
	}
	if o.GPUProduct == "" || o.GPUProduct != node.Labels[ProductLabel] || o.GPUMemory != node.Labels[MemoryLabel] || o.GPUCount != node.Labels[CountLabel] {
		return ""
	}
	count, err := strconv.ParseInt(o.GPUCount, 10, 32)
	if err != nil || count <= 0 || count != int64(len(o.Devices)) {
		return ""
	}
	memory, err := strconv.ParseUint(o.GPUMemory, 10, 63)
	if err != nil || memory == 0 {
		return ""
	}
	pci := o.Devices[0].PCIDeviceID
	if !pciPattern.MatchString(pci) || pci == "0000:10de" {
		return ""
	}
	uuids, buses := map[string]bool{}, map[string]bool{}
	for _, d := range o.Devices {
		uuid := strings.ToLower(strings.TrimSpace(d.UUID))
		if !strings.HasPrefix(uuid, "gpu-") || len(uuid) <= 4 || uuids[uuid] || !busPattern.MatchString(d.BusID) || buses[d.BusID] || d.Product != o.GPUProduct || d.PCIDeviceID != pci {
			return ""
		}
		uuids[uuid] = true
		buses[d.BusID] = true
	}
	return pci
}
