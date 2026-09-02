/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
)

type fakeHelmInstaller struct {
	values []string
}

func (f *fakeHelmInstaller) InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string) error {
	f.values = append([]string{}, values...)
	return nil
}

func (f *fakeHelmInstaller) InstallOrUpgradeFromLocalChart(ctx context.Context, chartPath, releaseName, targetNamespace string, values []string, valueFiles []string) error {
	f.values = append([]string{}, values...)
	return nil
}

func (f *fakeHelmInstaller) ReleaseStatus(releaseName, targetNamespace string) (bool, string, error) {
	return false, "", nil
}

func contains(slice []string, needle string) bool {
	for _, v := range slice {
		if v == needle {
			return true
		}
	}
	return false
}

func TestInstallCPChart_AddsCustomCAValue_WhenCAPathProvided(t *testing.T) {
	fi := &fakeHelmInstaller{}
	step := NewInstallCPChart(fi, nil).(*InstallCPChart)

	fqdn := "example.run.ai"
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN:        fqdn,
				configinputs.InstallInputCAChainPath: "/some/path/ca.pem",
			},
		},
		Config: &config.InstallConfig{
			CpVersion: "1.2.3",
		},
	}

	if err := step.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !contains(fi.values, "global.customCA.enabled=true") {
		t.Fatalf("expected values to include global.customCA.enabled=true, got: %v", fi.values)
	}
	wantDomain := "global.domain=" + fqdn
	if !contains(fi.values, wantDomain) {
		t.Fatalf("expected values to include %q, got: %v", wantDomain, fi.values)
	}
}

func TestInstallCPChart_EmptyFQDN_ReturnsErrorOrTimesOut(t *testing.T) {
	fi := &fakeHelmInstaller{}
	step := NewInstallCPChart(fi, nil).(*InstallCPChart)

	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				// FQDN intentionally omitted to trigger validation
			},
		},
		Config: &config.InstallConfig{
			CpVersion: "1.2.3",
		},
	}

	err := step.Execute(context.Background(), runaiData)
	if err == nil {
		t.Fatalf("expected error for empty FQDN, got nil")
	}
	if err.Error() != "FQDN is required but not provided" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstallCPChart_DoesNotAddCustomCAValue_WhenCAPathMissing(t *testing.T) {
	fi := &fakeHelmInstaller{}
	step := NewInstallCPChart(fi, nil).(*InstallCPChart)

	fqdn := "example.run.ai"
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN: fqdn,
				// CAChainPath intentionally absent
			},
		},
		Config: &config.InstallConfig{
			CpVersion: "1.2.3",
		},
	}

	if err := step.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	for _, v := range fi.values {
		if strings.HasPrefix(v, "global.customCA.enabled=") {
			t.Fatalf("did not expect customCA flag, got: %v", fi.values)
		}
	}
	wantDomain := "global.domain=" + fqdn
	if !contains(fi.values, wantDomain) {
		t.Fatalf("expected values to include %q, got: %v", wantDomain, fi.values)
	}
}
