package spec

import (
	"sort"
	"strings"
)

func (s *Spec) SelfSignedTLSEnabled() bool {
	return s.SelfSignedTLS != nil && s.SelfSignedTLS.Enabled
}

func (s *Spec) SelfSignedTLSDNSNames() []string {
	names := map[string]struct{}{}
	addDomain := func(domain string) {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			return
		}
		names[domain] = struct{}{}
		names["*."+domain] = struct{}{}
	}
	if s.ControlPlane != nil {
		addDomain(s.ControlPlane.Domain)
	}
	if s.Cluster != nil {
		domain := s.Cluster.Domain
		if domain == "" && s.ControlPlane != nil {
			domain = s.ControlPlane.Domain
		}
		addDomain(domain)
		if s.Cluster.Inference != nil {
			if domain := strings.TrimSpace(s.Cluster.Inference.WildcardDomain); domain != "" {
				names[domain] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (s *Spec) SelfSignedTLSUnsupportedReason() string {
	if s.IsOpenShift() {
		return "self-signed certificate generation is not supported on OpenShift because Routes own TLS termination"
	}
	if s.ExternalConnection != nil && s.ExternalConnection.Type == ExternalConnectionGatewayAPI {
		return "self-signed certificate generation is not supported with an existing Gateway API resource because the Gateway owns TLS termination"
	}
	return ""
}
