package inputcheck

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, systemNodesCheck) }

// systemNodesCheck validates the system-node placement config offline: the mode
// must be a known value, and when the installer manages labels exactly one of
// nodes/selector must be set (otherwise there is nothing to label).
func systemNodesCheck(s *spec.Spec) []Finding {
	sn := s.SystemNodes
	if sn == nil {
		return nil
	}
	var findings []Finding

	switch sn.Mode {
	case spec.SystemNodeModePreferred, spec.SystemNodeModeRequired:
	default:
		findings = append(findings, Finding{
			Check:   "system-nodes-mode",
			Field:   "systemNodes.mode",
			Message: fmt.Sprintf("invalid mode %q: must be %q or %q", sn.Mode, spec.SystemNodeModePreferred, spec.SystemNodeModeRequired),
		})
	}

	if sn.ManageLabels {
		hasNodes := len(sn.Nodes) > 0
		hasSelector := len(sn.Selector) > 0
		if hasNodes == hasSelector {
			findings = append(findings, Finding{
				Check:   "system-nodes-selection",
				Field:   "systemNodes",
				Message: "exactly one of systemNodes.nodes or systemNodes.selector must be set when manageLabels is true",
			})
		}
	}

	return findings
}
