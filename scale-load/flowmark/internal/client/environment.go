/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

const (
	versionPath  = "/_version"
	fakeGPULabel = "run.ai/fake.gpu"
)

type Environment struct {
	ControlPlaneVersion string
	ClusterName         string
	ClusterVersion      string
	Platform            string
	KubernetesVersion   string
	FakeGPUs            bool
}

func (e Environment) ClusterType() string {
	platform := e.Platform
	if platform == "" {
		platform = "unknown"
	}
	if e.FakeGPUs {
		return platform + "-fake-gpu"
	}
	return platform + "-real-gpu"
}

func (e Environment) String() string {
	gpus := "real GPUs"
	if e.FakeGPUs {
		gpus = "fake GPUs"
	}
	return fmt.Sprintf("control plane %s, cluster %s %s (%s, kubernetes %s, %s)", e.ControlPlaneVersion, e.ClusterName, e.ClusterVersion, e.Platform, e.KubernetesVersion, gpus)
}

func (c *Clients) Environment(ctx context.Context, clusterName string) (Environment, error) {
	controlPlane, err := c.ControlPlaneVersion(ctx)
	if err != nil {
		return Environment{}, err
	}
	clusters, _, err := c.ClusterService.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return Environment{}, fmt.Errorf("listing clusters: %w", err)
	}
	index := 0
	if clusterName != "" {
		index = -1
		for i := range clusters {
			if clusters[i].Name == clusterName {
				index = i
			}
		}
		if index < 0 {
			return Environment{}, fmt.Errorf("cluster %q not found on the control plane", clusterName)
		}
	} else if len(clusters) != 1 {
		return Environment{}, fmt.Errorf("control plane has %d clusters, set the cluster knob to choose one", len(clusters))
	}
	cluster := clusters[index]
	status := cluster.GetStatus()
	platform := status.GetPlatform()

	fake, err := c.fakeGPUs(ctx)
	if err != nil {
		return Environment{}, err
	}
	return Environment{
		ControlPlaneVersion: controlPlane,
		ClusterName:         cluster.Name,
		ClusterVersion:      cluster.GetVersion(),
		Platform:            platform.GetType(),
		KubernetesVersion:   platform.GetKubeVersion(),
		FakeGPUs:            fake,
	}, nil
}

func (c *Clients) ControlPlaneVersion(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+versionPath, nil)
	if err != nil {
		return "", err
	}
	response, err := c.http.Do(request)
	if err != nil {
		return "", fmt.Errorf("reading the control plane version: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reading the control plane version: %s", response.Status)
	}
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decoding the control plane version: %w", err)
	}
	if body.Version == "" {
		return "", fmt.Errorf("the control plane reported an empty version")
	}
	return body.Version, nil
}

func (c *Clients) fakeGPUs(ctx context.Context) (bool, error) {
	var nodes corev1.NodeList
	if err := c.Kube.List(ctx, &nodes); err != nil {
		return false, fmt.Errorf("listing nodes: %w", err)
	}
	for _, node := range nodes.Items {
		if node.Labels[fakeGPULabel] == "true" {
			return true, nil
		}
	}
	return false, nil
}
