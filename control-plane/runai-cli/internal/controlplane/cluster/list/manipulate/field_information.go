/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/utils/ptr"
)

const (
	DefaultWidthName             = 20
	DefaultWidthURL              = 30
	DefaultWidthVersion          = 20
	DefaultWidthStatus           = 10
	DefaultWidthKubeDistribution = 25
	DefaultWidthKubeVersion      = 20
)

var emptyDisplayPlatform = openapi.ClusterReportedStatusPlatform{Type: openapi.NullableString{}}

// Field definitions for cluster columns
var (
	NameField = FieldInformation{
		columnTitle: "Name",
		columnWidth: DefaultWidthName,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	URLField = FieldInformation{
		columnTitle: "URL",
		columnWidth: DefaultWidthURL,
		extractFunc: extractURLData,
		formatFunc:  formatURLValue,
	}
	VersionField = FieldInformation{
		columnTitle: "Run:AI Cluster Version",
		columnWidth: DefaultWidthVersion,
		extractFunc: extractVersionData,
		formatFunc:  formatVersionValue,
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidthStatus,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
	KubeDistributionField = FieldInformation{
		columnTitle: "Kubernetes Distribution",
		columnWidth: DefaultWidthKubeDistribution,
		extractFunc: extractKubeDistributionData,
		formatFunc:  formatKubeDistributionValue,
	}
	KubeVersionField = FieldInformation{
		columnTitle: "Distribution Version",
		columnWidth: DefaultWidthKubeVersion,
		extractFunc: extractKubeVersionData,
		formatFunc:  formatKubeVersionValue,
	}
)

// Extract functions for cluster fields
func extractNameData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	clusterRow.Name = cluster.Name
}

func extractURLData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	if cluster.Domain.IsSet() {
		clusterRow.URL = cluster.GetDomain()
	} else {
		clusterRow.URL = ""
	}
}

func extractVersionData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	if cluster.Version.IsSet() {
		clusterRow.Version = cluster.GetVersion()
	} else {
		clusterRow.Version = ""
	}
}

func extractStatusData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	clusterRow.Status = string(cluster.GetState())
}

func extractKubeDistributionData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	platform := ptr.Deref(cluster.Platform.Get(), emptyDisplayPlatform)
	clusterRow.KubeDistribution = platform.GetType()
}

func extractKubeVersionData(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow) {
	platform := ptr.Deref(cluster.Platform.Get(), emptyDisplayPlatform)
	clusterRow.DistributionVersion = platform.GetKubeVersion()
}

// Format functions for cluster fields
func formatNameValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.Name
}

func formatURLValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.URL
}

func formatVersionValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.Version
}

func formatStatusValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.Status
}

func formatKubeDistributionValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.KubeDistribution
}

func formatKubeVersionValue(clusterRow *ClustersAlignedRow) string {
	return clusterRow.DistributionVersion
}
