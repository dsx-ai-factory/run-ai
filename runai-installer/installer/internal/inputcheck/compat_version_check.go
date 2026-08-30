package inputcheck

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/compat"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, compatVersionCheck) }

// compatVersionCheck rejects any Run:ai version the compatibility matrix doesn't
// recognize. The control-plane and cluster versions are each checked on their
// own, so a bad version on either is caught (in a full install they are
// normally the same release). Such a version has no known-good prerequisite
// locks, so installing it would silently fall back to a different release's
// versions. Dev mode bypasses the check: it falls back to the latest matrix and
// warns instead (see buildPrereqFlows), for testing as-yet-unreleased versions.
// An unset version is allowed (it naturally defaults to the latest).
func compatVersionCheck(s *spec.Spec) []Finding {
	if s.DevMode || s.UsesAirgapLoad() {
		return nil
	}
	var findings []Finding
	check := func(version, field string) {
		if version == "" {
			return
		}
		if _, fellBack := compat.Resolve(version); fellBack {
			findings = append(findings, Finding{
				Check:   "compat-version",
				Field:   field,
				Message: fmt.Sprintf("Run:ai version %q is not in the compatibility matrix (supported: %s); pass --dev-mode to install an unsupported version", version, strings.Join(compat.SupportedVersions(), ", ")),
			})
		}
	}
	if s.ControlPlane != nil {
		check(s.ControlPlane.Version, "controlPlane.version")
	}
	if s.Cluster != nil {
		check(s.Cluster.Version, "cluster.version")
	}
	return findings
}
