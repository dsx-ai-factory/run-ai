/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package capability

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	"github.com/hashicorp/go-version"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	clusterservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"k8s.io/utils/ptr"
)

// --- Actions and Constants ---

type Actions int

const (
	Logs Actions = iota
	Exec
	PortForward
)

const (
	MinimalClusterVersionForInfwClusterApiActions = "2.22.0"
	testEnvironmentVersion                        = "master"
	ClusterPrToMasterVersionPrefix                = "0.0.0"
	ClusterVersionError                           = "cluster version is not supported; please upgrade your cluster to at least version %s"
)

type Checker struct {
	ControlPlaneURL string
	Token           string
	ClusterUUID     string

	// allow dependency injection for testability
	getClusterByUuidFunc func(txn *transaction.Txn, url, token, clusterUUID string) (*manipulate.BriefCluster, error)
}

func NewChecker(cpURL, token, clusterUUID string) *Checker {
	return &Checker{
		ControlPlaneURL:      cpURL,
		Token:                token,
		ClusterUUID:          clusterUUID,
		getClusterByUuidFunc: clusterservice.GetClusterByUuid,
	}
}

// IsSupported checks if a given action is supported for the workload type in the specified cluster
func (m *Checker) IsSupported(txn *transaction.Txn, workloadType api.WorkloadType, action Actions) (bool, error) {
	if workloadType != api.Inference {
		return true, nil // No restriction on other workload types
	}

	if action != Logs && action != Exec && action != PortForward {
		return true, nil // Only specific actions are restricted
	}

	cluster, err := m.getClusterByUuidFunc(txn, m.ControlPlaneURL, m.Token, m.ClusterUUID)
	if err != nil {
		return false, fmt.Errorf("failed to get cluster info: %w", err)
	}

	isSupported, err := isClusterVersionSupported(cluster, MinimalClusterVersionForInfwClusterApiActions)
	if err != nil {
		return false, err
	}

	return ptr.Deref(isSupported, false), nil
}

func isClusterVersionSupported(cluster *manipulate.BriefCluster, requiredVersion string) (*bool, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster data is nil")
	}

	if isTestEnvironmentVersion(cluster.Version) {
		return ptr.To(true), nil
	}

	clusterVer, err := version.NewVersion(cluster.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster version %q: %w", cluster.Version, err)
	}

	requiredVer, err := version.NewVersion(requiredVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid required version %q: %w", requiredVersion, err)
	}

	supported := !clusterVer.LessThan(requiredVer)
	return ptr.To(supported), nil
}

func isTestEnvironmentVersion(clusterVersion string) bool {
	if strings.Contains(clusterVersion, testEnvironmentVersion) || strings.Contains(clusterVersion, ClusterPrToMasterVersionPrefix) {
		return true
	}
	return false
}
