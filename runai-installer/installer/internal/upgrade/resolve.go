/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

func ResolveUpgradeParams(cfg *config.UpgradeConfig, readValues HelmValuesReader) error {
	if cfg.FQDN == "" {
		fqdn, err := readFQDNFromHelm(readValues)
		if err != nil {
			return fmt.Errorf("--fqdn not provided and could not read from existing installation: %w", err)
		}
		logging.Infof("[Upgrade] Auto-detected FQDN: %s", fqdn)
		cfg.FQDN = fqdn
	} else {
		logging.Infof("[Upgrade] Using provided FQDN: %s", cfg.FQDN)
	}

	if cfg.UpgradeCluster && cfg.ClusterUUID == "" {
		uuid, err := readClusterUUIDFromHelm(readValues)
		if err != nil {
			return fmt.Errorf("--cluster-uuid not provided and could not read from existing installation: %w", err)
		}
		logging.Infof("[Upgrade] Auto-detected cluster UUID: %s", uuid)
		cfg.ClusterUUID = uuid
	} else if cfg.ClusterUUID != "" {
		logging.Infof("[Upgrade] Using provided cluster UUID: %s", cfg.ClusterUUID)
	}

	if cfg.UpgradeControlPlane {
		vals, err := readCarryForwardValues(readValues, step.CPReleaseName, step.CPNamespace)
		if err != nil {
			return fmt.Errorf("could not read carry-forward values from CP release: %w", err)
		} else if len(vals) > 0 {
			logging.Infof("[Upgrade] Carrying forward %d value(s) from CP release", len(vals))
			cfg.CPCarryForwardValues = vals
		}
	}

	if cfg.UpgradeCluster {
		vals, err := readCarryForwardValues(readValues, step.ClusterReleaseName, step.ClusterNamespace)
		if err != nil {
			return fmt.Errorf("could not read carry-forward values from cluster release: %w", err)
		} else if len(vals) > 0 {
			logging.Infof("[Upgrade] Carrying forward %d value(s) from cluster release", len(vals))
			cfg.ClusterCarryForwardValues = vals
		}
	}

	return nil
}

func readCarryForwardValues(readValues HelmValuesReader, releaseName, namespace string) ([]string, error) {
	vals, err := readValues(releaseName, namespace)
	if err != nil {
		return nil, err
	}
	var result []string
	result = append(result, extractCustomCAValues(vals)...)
	result = append(result, extractKubernetesDistribution(vals)...)
	return result, nil
}

func extractCustomCAValues(vals map[string]interface{}) []string {
	var result []string
	global, ok := vals["global"].(map[string]interface{})
	if !ok {
		return result
	}
	customCA, ok := global["customCA"].(map[string]interface{})
	if !ok {
		return result
	}
	enabled, ok := customCA["enabled"].(bool)
	if !ok || !enabled {
		return result
	}
	result = append(result, "global.customCA.enabled=true")
	if secret, ok := customCA["secret"].(map[string]interface{}); ok {
		if name, ok := secret["name"].(string); ok && name != "" {
			result = append(result, "global.customCA.secret.name="+name)
		}
		if key, ok := secret["key"].(string); ok && key != "" {
			result = append(result, "global.customCA.secret.key="+key)
		}
		if optional, ok := secret["optional"].(bool); ok {
			result = append(result, fmt.Sprintf("global.customCA.secret.optional=%t", optional))
		}
	}
	return result
}

func extractKubernetesDistribution(vals map[string]interface{}) []string {
	global, ok := vals["global"].(map[string]interface{})
	if !ok {
		return nil
	}
	cfg, ok := global["config"].(map[string]interface{})
	if !ok {
		return nil
	}
	dist, ok := cfg["kubernetesDistribution"].(string)
	if !ok || dist == "" || dist == "kubernetes" {
		return nil
	}
	return []string{"global.config.kubernetesDistribution=" + dist}
}

func readFQDNFromHelm(readValues HelmValuesReader) (string, error) {
	vals, err := readValues(step.CPReleaseName, step.CPNamespace)
	if err != nil {
		return "", err
	}
	global, ok := vals["global"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("'global' key not found in %s values", step.CPReleaseName)
	}
	domain, ok := global["domain"].(string)
	if !ok || domain == "" {
		return "", fmt.Errorf("'global.domain' not found in %s values", step.CPReleaseName)
	}
	return domain, nil
}

func readClusterUUIDFromHelm(readValues HelmValuesReader) (string, error) {
	vals, err := readValues(step.ClusterReleaseName, step.ClusterNamespace)
	if err != nil {
		return "", err
	}
	clusterCfg, ok := vals["cluster"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("'cluster' key not found in %s values", step.ClusterReleaseName)
	}
	uid, ok := clusterCfg["uid"].(string)
	if !ok || uid == "" {
		return "", fmt.Errorf("'cluster.uid' not found in %s values", step.ClusterReleaseName)
	}
	return uid, nil
}
