package steps

import (
	"context"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-installer/clients/runai/cluster"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
)

type fakeHelmInstallerCluster struct {
	values []string
}

func (f *fakeHelmInstallerCluster) InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string) error {
	f.values = append([]string{}, values...)
	return nil
}

func (f *fakeHelmInstallerCluster) InstallOrUpgradeFromLocalChart(ctx context.Context, chartPath, releaseName, targetNamespace string, values []string, valueFiles []string) error {
	f.values = append([]string{}, values...)
	return nil
}

func (f *fakeHelmInstallerCluster) ReleaseStatus(releaseName, targetNamespace string) (bool, string, error) {
	return false, "", nil
}

func TestInstallClusterChart_AddsCustomCAValue_WhenCAPathProvided(t *testing.T) {
	create := NewCreateCluster(nil)
	create.clusterUuid = "uuid-123"
	installInfoStep := NewGetClusterInstallationDetails(create, nil)
	installInfoStep.installInfo = cluster.NewClusterInstallationInfoResponse("", "", "", "s3cr3t")

	fi := &fakeHelmInstallerCluster{}
	step := NewInstallClusterChart(installInfoStep, fi, nil).(*InstallClusterChart)

	fqdn := "example.run.ai"
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN:        fqdn,
				configinputs.InstallInputCAChainPath: "/some/path/ca.pem",
			},
		},
		Config: &config.InstallConfig{
			ClusterVersion: "1.2.3",
		},
	}

	if err := step.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !contains(fi.values, "global.customCA.enabled=true") {
		t.Fatalf("expected values to include global.customCA.enabled=true, got: %v", fi.values)
	}
	wantDomain := "controlPlane.url=" + fqdn
	if !contains(fi.values, wantDomain) {
		t.Fatalf("expected values to include %q, got: %v", wantDomain, fi.values)
	}
}

func TestInstallClusterChart_UsesBootstrapToken(t *testing.T) {
	create := NewCreateCluster(nil)
	create.clusterUuid = "uuid-123"
	installInfoStep := NewGetClusterInstallationDetails(create, nil)
	installInfoStep.installInfo = &cluster.ClusterInstallationInfoResponse{
		BootstrapToken: "bootstrap-token",
		ClientSecret:   "client-secret",
	}

	fi := &fakeHelmInstallerCluster{}
	installStep := NewInstallClusterChart(installInfoStep, fi, nil).(*InstallClusterChart)
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN: "example.run.ai",
			},
		},
		Config: &config.InstallConfig{ClusterVersion: "1.2.3"},
	}

	if err := installStep.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !contains(fi.values, "controlPlane.bootstrapToken=bootstrap-token") {
		t.Fatalf("expected bootstrap token value, got: %v", fi.values)
	}
	if contains(fi.values, "controlPlane.clientSecret=client-secret") {
		t.Fatalf("did not expect client secret value, got: %v", fi.values)
	}
}

func TestInstallClusterChart_SubdomainSupport_HostRouting(t *testing.T) {
	create := NewCreateCluster(nil)
	create.clusterUuid = "uuid-123"
	installInfoStep := NewGetClusterInstallationDetails(create, nil)
	installInfoStep.installInfo = cluster.NewClusterInstallationInfoResponse("", "", "", "s3cr3t")

	fi := &fakeHelmInstallerCluster{}
	step := NewInstallClusterChart(installInfoStep, fi, nil).(*InstallClusterChart)

	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN:            "example.run.ai",
				configinputs.InstallInputWorkloadRouting: pkgvalidation.WorkloadRoutingHost,
			},
		},
		Config: &config.InstallConfig{ClusterVersion: "1.2.3"},
	}

	if err := step.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !contains(fi.values, "clusterConfig.global.subdomainSupport=true") {
		t.Fatalf("expected clusterConfig.global.subdomainSupport=true, got: %v", fi.values)
	}
}

func TestInstallClusterChart_SubdomainSupport_PathRouting(t *testing.T) {
	create := NewCreateCluster(nil)
	create.clusterUuid = "uuid-123"
	installInfoStep := NewGetClusterInstallationDetails(create, nil)
	installInfoStep.installInfo = cluster.NewClusterInstallationInfoResponse("", "", "", "s3cr3t")

	fi := &fakeHelmInstallerCluster{}
	step := NewInstallClusterChart(installInfoStep, fi, nil).(*InstallClusterChart)

	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN: "example.run.ai",
			},
		},
		Config: &config.InstallConfig{ClusterVersion: "1.2.3"},
	}

	if err := step.Execute(context.Background(), runaiData); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !contains(fi.values, "clusterConfig.global.subdomainSupport=false") {
		t.Fatalf("expected clusterConfig.global.subdomainSupport=false, got: %v", fi.values)
	}
	if contains(fi.values, "clusterConfig.global.subdomainSupport=true") {
		t.Fatalf("did not expect clusterConfig.global.subdomainSupport=true, got: %v", fi.values)
	}
}

func TestInstallClusterChart_DoesNotAddCustomCAValue_WhenCAPathMissing(t *testing.T) {
	create := NewCreateCluster(nil)
	create.clusterUuid = "uuid-123"
	installInfoStep := NewGetClusterInstallationDetails(create, nil)
	installInfoStep.installInfo = cluster.NewClusterInstallationInfoResponse("", "", "", "s3cr3t")

	fi := &fakeHelmInstallerCluster{}
	step := NewInstallClusterChart(installInfoStep, fi, nil).(*InstallClusterChart)

	fqdn := "example.run.ai"
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN: fqdn,
			},
		},
		Config: &config.InstallConfig{
			ClusterVersion: "1.2.3",
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
	wantDomain := "controlPlane.url=" + fqdn
	if !contains(fi.values, wantDomain) {
		t.Fatalf("expected values to include %q, got: %v", wantDomain, fi.values)
	}
}
