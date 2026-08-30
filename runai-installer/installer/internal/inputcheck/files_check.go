package inputcheck

import (
	"path/filepath"

	"github.com/run-ai/runai/runai-installer/installer/internal/airgapvalues"
	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, filesCheck) }

// filesCheck verifies that file paths referenced by the spec actually exist on
// disk. Certificate and key existence is covered by the certificate and
// key-pair checks; this covers the remaining paths: CA chains and the air-gap
// bundle / working directory. Token and password files are not checked here
// because the registry resolver reads them into the spec before validation.
func filesCheck(s *spec.Spec) []Finding {
	var findings []Finding

	if s.Registry != nil && s.Registry.CAChainPath != "" {
		findings = append(findings, fromResult(
			validation.ValidateSSLPaths(s.Registry.CAChainPath),
			"file-exists", "registry.caChainPath",
		)...)
	}

	for _, src := range tlsSources(s) {
		if src.tls.CAChainPath != "" {
			findings = append(findings, fromResult(
				validation.ValidateSSLPaths(src.tls.CAChainPath),
				"file-exists", src.field+".caChainPath",
			)...)
		}
	}

	findings = append(findings, airgapFileFindings(s)...)

	return findings
}

// airgapFileFindings checks the air-gap inputs that must exist before install.
// Without SkipPrepare the bundle is extracted, so the bundle must exist; with
// SkipPrepare the bundle is untouched and the already-populated working
// directory must exist instead.
func airgapFileFindings(s *spec.Spec) []Finding {
	if !s.IsAirGapped() || s.Airgap == nil {
		return nil
	}
	if s.Airgap.SkipPrepare {
		if s.UsesAirgapLoad() {
			return loadedAirgapFindings(s)
		}
		return fromResult(
			validation.ValidateSSLPaths(s.Airgap.WorkingDir),
			"file-exists", "airgap.workingDir",
		)
	}
	return fromResult(
		validation.ValidateSSLPaths(s.Airgap.BundlePath),
		"file-exists", "airgap.bundlePath",
	)
}

func loadedAirgapFindings(s *spec.Spec) []Finding {
	paths := []struct {
		field string
		path  string
	}{
		{"airgap.workingDir", s.Airgap.WorkingDir},
		{"airgap.valuesDir", filepath.Join(s.Airgap.ValuesDir, airgapvalues.ControlPlaneValuesFile)},
		{"airgap.valuesDir", filepath.Join(s.Airgap.ValuesDir, airgapvalues.ClusterValuesFile)},
	}
	var findings []Finding
	for _, item := range paths {
		findings = append(findings, fromResult(
			validation.ValidateSSLPaths(item.path),
			"file-exists", item.field,
		)...)
	}
	for _, chart := range requiredAirgapCharts(s) {
		if _, err := artifacts.LockedChartPath(s.Airgap.WorkingDir, chart); err != nil {
			findings = append(findings, Finding{
				Check:   "airgap-loaded-result",
				Field:   "airgap.workingDir",
				Message: err.Error(),
			})
		}
	}
	return findings
}

func requiredAirgapCharts(s *spec.Spec) []string {
	var charts []string
	if s.ControlPlane != nil {
		charts = append(charts, "control-plane")
	}
	if s.Cluster != nil {
		charts = append(charts, "runai-cluster")
	}
	return charts
}
