package inputcheck

import "github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"

func init() { checks = append(checks, selfSignedTLSCheck) }

func selfSignedTLSCheck(s *spec.Spec) []Finding {
	if !s.SelfSignedTLSEnabled() {
		return nil
	}
	var findings []Finding
	for _, field := range s.SelfSignedTLS.ConflictingFields {
		findings = append(findings, Finding{
			Check:   "self-signed-tls",
			Field:   field,
			Message: "remove this certificate configuration when tls.generateSelfSigned is true",
		})
	}
	if reason := s.SelfSignedTLSUnsupportedReason(); reason != "" {
		findings = append(findings, Finding{
			Check:   "self-signed-tls",
			Field:   "tls.generateSelfSigned",
			Message: reason,
		})
	}
	if len(s.SelfSignedTLSDNSNames()) == 0 {
		findings = append(findings, Finding{
			Check:   "self-signed-tls",
			Field:   "tls.generateSelfSigned",
			Message: "a control-plane or cluster domain is required to generate a certificate",
		})
	}
	return findings
}
