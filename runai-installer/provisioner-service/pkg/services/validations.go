package services

import (
	"fmt"
	"regexp"

	"github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
)

var (
	// Ensures the domain is a valid FQDN, for example: example.com
	fqdnRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	// Ensures the wildcard domain is a valid wildcard FQDN, for example: *.example.com
	wildcardDomainRegex = regexp.MustCompile(`^\*\.([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	// Ensures the email is a valid email address
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	validProvisioningModes = map[string]bool{
		"control-plane-only": true,
		"full":               true,
	}
)

func validateCreateEnvironmentRequest(req api.CreateEnvironmentRequest) error {
	if !isValidProvisioningMode(req.ProvisioningMode) {
		return fmt.Errorf("invalid provisioning mode: %s must be either 'control-plane-only' or 'full'", req.ProvisioningMode)
	}

	if !isValidFQDN(req.Domain) {
		return fmt.Errorf("invalid domain: %s is not a valid FQDN", req.Domain)
	}

	if !isValidWildcardDomain(req.Wildcarddomain) {
		return fmt.Errorf("invalid wildcard domain: %s is not a valid wildcard domain (must start with *.)", req.Wildcarddomain)
	}

	if !validation.IsValidKubernetesDistribution(req.KubernetesDistribution) {
		return fmt.Errorf("invalid kubernetes distribution: %s must be one of: %v, or empty", req.KubernetesDistribution, validation.ValidKubernetesDistributions)
	}

	// userEmail is required when provisioningMode is "full"
	if req.ProvisioningMode == "full" {
		if req.UserEmail == "" {
			return fmt.Errorf("userEmail is required when provisioningMode is 'full'")
		}
		if !isValidEmail(req.UserEmail) {
			return fmt.Errorf("invalid userEmail: %s is not a valid email address", req.UserEmail)
		}
	}

	return nil
}

func isValidFQDN(domain string) bool {
	return fqdnRegex.MatchString(domain)
}

func isValidWildcardDomain(domain string) bool {
	return wildcardDomainRegex.MatchString(domain)
}

func isValidProvisioningMode(mode string) bool {
	return validProvisioningModes[mode]
}

func isValidKubernetesDistribution(distribution string) bool {
	return distribution == "openshift"
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
