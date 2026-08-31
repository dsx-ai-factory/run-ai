/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

type ClusterPlatform struct {
	Type        string `json:"type,omitempty"`
	KubeVersion string `json:"kubeVersion,omitempty"`
}

type ClusterStatus struct {
	State    openapi.ClusterState `json:"state,omitempty"`
	Platform ClusterPlatform      `json:"platform,omitempty"`
}

type BriefCluster struct {
	Uuid    string        `json:"uuid,omitempty"`
	Name    string        `json:"name,omitempty"`
	Domain  string        `json:"domain,omitempty"`
	Version string        `json:"version,omitempty"`
	Status  ClusterStatus `json:"status,omitempty"`
}

type Data struct {
	Clusters  []openapi.ClustersMinimalData `json:"clusters,omitempty" yaml:"clusters"`
	NextToken *string                       `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type BriefData struct {
	Clusters  []BriefCluster `json:"clusters,omitempty" yaml:"clusters"`
	NextToken *string        `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

// ClustersAlignedRow represents a data row with proper fields for aligned output
type ClustersAlignedRow struct {
	Name                string `json:"name,omitempty"`
	URL                 string `json:"url,omitempty"`
	Version             string `json:"version,omitempty"`
	Status              string `json:"status,omitempty"`
	KubeDistribution    string `json:"kubeDistribution,omitempty"`
	DistributionVersion string `json:"distributionVersion,omitempty"`
}

// ClustersAlignedOutput represents clusters output with column titles as keys
type ClustersAlignedOutput struct {
	Clusters  []ClustersAlignedRow `json:"clusters" yaml:"clusters"`
	NextToken *string              `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

func newBriefFromCluster(cluster openapi.ClustersMinimalData) BriefCluster {
	brief := BriefCluster{
		Uuid: cluster.Uuid,
		Name: cluster.Name,
	}

	// Handle optional Domain field
	if cluster.Domain.IsSet() {
		brief.Domain = cluster.GetDomain()
	}

	// Handle optional Version field
	if cluster.Version.IsSet() {
		brief.Version = cluster.GetVersion()
	}

	// Handle optional Status field
	if cluster.State != nil {
		brief.Status.State = *cluster.State

		// Handle nested Platform within Status
		if cluster.Platform.IsSet() {
			platform := cluster.GetPlatform()
			brief.Status.Platform.Type = platform.GetType()
			brief.Status.Platform.KubeVersion = platform.GetKubeVersion()
		}
	}

	return brief
}

func BriefListFromResponse(clusterList []openapi.ClustersMinimalData) ([]BriefCluster, error) {
	var briefList = make([]BriefCluster, 0)

	for _, cluster := range clusterList {
		briefList = append(briefList, newBriefFromCluster(cluster))
	}

	return briefList, nil
}

func ReduceStruct(cluster openapi.ClustersMinimalData) BriefCluster {
	return newBriefFromCluster(cluster)
}
