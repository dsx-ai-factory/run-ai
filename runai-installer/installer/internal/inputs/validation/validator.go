package validation

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
)

// Shared format patterns, reused by both the network-aware validators (which
// additionally resolve DNS) and the format-only validators used by offline
// input validation.
var (
	fqdnFormatRegex   = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	wildcardFQDNRegex = regexp.MustCompile(`^\*\.[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,}$`)
)

// ValidationResult represents a validation result
type ValidationResult struct {
	Name    string
	Status  bool // true for success, false for failure
	Message string
}

// ValidateJfrogCredentials validates NGC credentials in base64 format
func ValidateJfrogCredentials(credentials string) ValidationResult {
	credentials = strings.TrimSpace(credentials)

	if credentials == "" {
		return ValidationResult{
			Name:    "Jfrog credentials",
			Status:  false,
			Message: "registry token is required: provide the base64-encoded docker config (.dockerconfigjson) from JFrog",
		}
	}

	// Decode with the same strict standard-base64 decoder used to build the
	// registry secret (registry_secret.go) so input validation accepts exactly
	// what the install will later decode — not a lenient regex that passes
	// tokens (wrong padding/length, URL-safe alphabet) the real decoder rejects.
	if _, err := base64.StdEncoding.DecodeString(credentials); err != nil {
		return ValidationResult{
			Name:    "Jfrog credentials",
			Status:  false,
			Message: "registry token must be a single-line standard base64-encoded docker config (.dockerconfigjson) with no spaces or line breaks.",
		}
	}

	return ValidationResult{
		Name:    "Jfrog credentials",
		Status:  true,
		Message: "Valid",
	}
}

// ValidateFQDN validates FQDN format and reachability
func ValidateFQDN(fqdn string) ValidationResult {
	if fqdn == "" {
		return ValidationResult{
			Name:    "FQDN",
			Status:  false,
			Message: "FQDN is required",
		}
	}

	// FQDN validation regex
	if !fqdnFormatRegex.MatchString(fqdn) {
		return ValidationResult{
			Name:    "FQDN",
			Status:  false,
			Message: "Invalid FQDN format",
		}
	}

	// Check for whitespace and special characters
	if strings.ContainsAny(fqdn, " \t\n\r") {
		return ValidationResult{
			Name:    "FQDN",
			Status:  false,
			Message: "FQDN contains whitespace",
		}
	}

	// DNS resolution test
	_, err := net.LookupHost(fqdn)
	reachable, err := checkReachability(fqdn)
	if err != nil || !reachable {
		return ValidationResult{
			Name:    "FQDN",
			Status:  false,
			Message: "FQDN not reachable",
		}
	}

	return ValidationResult{
		Name:    "FQDN",
		Status:  true,
		Message: "Valid and reachable",
	}
}

// ValidateWildcardFQDN validates wildcard FQDN format
func ValidateWildcardFQDN(wildcardFQDN string) ValidationResult {
	if wildcardFQDN == "" {
		return ValidationResult{
			Name:    "Wildcard FQDN",
			Status:  true,
			Message: "Not required",
		}
	}

	// Wildcard FQDN validation regex - must start with "*." and follow domain pattern
	if !wildcardFQDNRegex.MatchString(wildcardFQDN) {
		return ValidationResult{
			Name:    "Wildcard FQDN",
			Status:  false,
			Message: "Invalid wildcard FQDN format",
		}
	}

	fqdn := strings.ReplaceAll(wildcardFQDN, "*", "test")
	reachable, err := checkReachability(fqdn)
	if err != nil || !reachable {
		return ValidationResult{
			Name:    "Wildcard FQDN",
			Status:  false,
			Message: "Domain not reachable",
		}
	}

	return ValidationResult{
		Name:    "Wildcard FQDN",
		Status:  true,
		Message: "Valid and reachable",
	}
}

// ValidateFQDNFormat validates FQDN format only, without resolving DNS. Use
// this for offline input validation; ValidateFQDN additionally checks
// reachability.
func ValidateFQDNFormat(fqdn string) ValidationResult {
	if fqdn == "" {
		return ValidationResult{Name: "FQDN", Status: false, Message: "FQDN is required"}
	}
	if strings.ContainsAny(fqdn, " \t\n\r") {
		return ValidationResult{Name: "FQDN", Status: false, Message: "FQDN contains whitespace"}
	}
	if !fqdnFormatRegex.MatchString(fqdn) {
		return ValidationResult{Name: "FQDN", Status: false, Message: "Invalid FQDN format"}
	}
	return ValidationResult{Name: "FQDN", Status: true, Message: "Valid"}
}

// ValidateWildcardFQDNFormat validates wildcard FQDN format only, without
// resolving DNS. An empty value is treated as valid (wildcard is optional).
func ValidateWildcardFQDNFormat(wildcardFQDN string) ValidationResult {
	if wildcardFQDN == "" {
		return ValidationResult{Name: "Wildcard FQDN", Status: true, Message: "Not required"}
	}
	if !wildcardFQDNRegex.MatchString(wildcardFQDN) {
		return ValidationResult{Name: "Wildcard FQDN", Status: false, Message: "Invalid wildcard FQDN format"}
	}
	return ValidationResult{Name: "Wildcard FQDN", Status: true, Message: "Valid"}
}

// ValidateClusterName validates Kubernetes cluster name
func ValidateClusterName(clusterName string) ValidationResult {
	if clusterName == "" {
		return ValidationResult{
			Name:    "Cluster Name",
			Status:  false,
			Message: "Cluster name is required",
		}
	}

	// Kubernetes cluster name validation regex
	clusterNameRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !clusterNameRegex.MatchString(clusterName) {
		return ValidationResult{
			Name:    "Cluster Name",
			Status:  false,
			Message: "Invalid cluster name format",
		}
	}

	return ValidationResult{
		Name:    "Cluster Name",
		Status:  true,
		Message: "Valid",
	}
}

// ValidateIngressControllerHttpsPort validates ingress controller HTTPS port
func ValidateIngressControllerHttpsPort(port string) ValidationResult {
	if port == "" {
		return ValidationResult{
			Name:    "Ingress Controller HTTPS Port",
			Status:  true,
			Message: "Not required",
		}
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return ValidationResult{
			Name:    "Ingress Controller HTTPS Port",
			Status:  false,
			Message: "Invalid port number",
		}
	}

	if portNum < 30000 || portNum > 32767 {
		return ValidationResult{
			Name:    "Ingress Controller HTTPS Port",
			Status:  false,
			Message: "Port must be between 30000 and 32767",
		}
	}

	return ValidationResult{
		Name:    "Ingress Controller HTTPS Port",
		Status:  true,
		Message: "Valid",
	}
}

// ValidateKubernetesDistribution validates Kubernetes distribution type
func ValidateKubernetesDistribution(distribution string) ValidationResult {
	if distribution == "" {
		return ValidationResult{
			Name:    "Kubernetes Distribution",
			Status:  true,
			Message: "Not required",
		}
	}

	if !pkgvalidation.IsValidKubernetesDistribution(distribution) {
		return ValidationResult{
			Name:    "Kubernetes Distribution",
			Status:  false,
			Message: fmt.Sprintf("Must be one of: %v, or empty", pkgvalidation.ValidKubernetesDistributions),
		}
	}

	return ValidationResult{
		Name:    "Kubernetes Distribution",
		Status:  true,
		Message: "Valid",
	}
}

func ValidateWorkloadRouting(value string) ValidationResult {
	switch value {
	case "", pkgvalidation.WorkloadRoutingPath, pkgvalidation.WorkloadRoutingHost:
		return ValidationResult{
			Name:    "Workload Routing",
			Status:  true,
			Message: "Valid",
		}
	default:
		return ValidationResult{
			Name:    "Workload Routing",
			Status:  false,
			Message: "Must be 'path' or 'host'",
		}
	}
}

// ValidateSSLPaths validates a file path exists
func ValidateSSLPaths(path string) ValidationResult {
	if path == "" {
		return ValidationResult{
			Name:    "SSL Path",
			Status:  false,
			Message: "Path is required",
		}
	}
	if _, err := os.Stat(path); err != nil {
		return ValidationResult{
			Name:    "SSL Path",
			Status:  false,
			Message: fmt.Sprintf("Path not accessible: %v", err),
		}
	}
	return ValidationResult{
		Name:    "SSL Path",
		Status:  true,
		Message: "Path exists",
	}
}

// ValidateCertificate validates PEM cert file, expiry and hostname
func ValidateCertificate(certPath string, host string) ValidationResult {
	if certPath == "" {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: "Certificate path is required"}
	}
	if host == "" {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: "Host (FQDN) is required"}
	}

	pemBytes, err := os.ReadFile(certPath)
	if err != nil {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: fmt.Sprintf("Cannot read certificate: %v", err)}
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: "Invalid PEM certificate"}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: fmt.Sprintf("Parse error: %v", err)}
	}

	now := time.Now()
	if now.Before(cert.NotBefore) {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: "Certificate not yet valid"}
	}
	if now.After(cert.NotAfter) {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: "Certificate expired"}
	}

	hostname := strings.TrimSpace(host)
	hostname = strings.TrimPrefix(hostname, "*.")
	if err := cert.VerifyHostname(hostname); err != nil {
		return ValidationResult{Name: "SSL Certificate", Status: false, Message: fmt.Sprintf("Hostname mismatch: %v", err)}
	}

	return ValidationResult{Name: "SSL Certificate", Status: true, Message: "Valid for host and not expired"}
}

// ValidateKeyPair verifies that the private key matches the certificate (and
// that both decode), purely from disk. It does not check expiry or hostname;
// use ValidateCertificate for those.
func ValidateKeyPair(certPath, keyPath string) ValidationResult {
	if certPath == "" || keyPath == "" {
		return ValidationResult{Name: "TLS Key Pair", Status: false, Message: "Certificate and key paths are required"}
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return ValidationResult{Name: "TLS Key Pair", Status: false, Message: fmt.Sprintf("Certificate and key do not match or cannot be loaded: %v", err)}
	}
	return ValidationResult{Name: "TLS Key Pair", Status: true, Message: "Valid"}
}

func checkReachability(domain string) (reachable bool, err error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false, fmt.Errorf("IP lookup failed: %w", err)
	}
	if len(ips) > 0 {
		return true, nil
	}

	return false, fmt.Errorf("none of the IPs are reachable: %w", err)
}

// ValidateEmail validates email format
func ValidateEmail(email string) ValidationResult {
	if email == "" {
		return ValidationResult{
			Name:    "User Email",
			Status:  true,
			Message: "Not required (optional)",
		}
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationResult{
			Name:    "User Email",
			Status:  false,
			Message: "Invalid email format",
		}
	}

	return ValidationResult{
		Name:    "User Email",
		Status:  true,
		Message: "Valid",
	}
}

// ValidateProjectName validates project name format
func ValidateProjectName(projectName string) ValidationResult {
	if projectName == "" {
		return ValidationResult{
			Name:    "Project Name",
			Status:  true,
			Message: "Not required (will use default: test-project)",
		}
	}

	// Project name validation regex - allows alphanumeric, hyphens, underscores
	projectNameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
	if !projectNameRegex.MatchString(projectName) {
		return ValidationResult{
			Name:    "Project Name",
			Status:  false,
			Message: "Invalid project name format (must start and end with alphanumeric, can contain hyphens and underscores)",
		}
	}

	if len(projectName) > 63 {
		return ValidationResult{
			Name:    "Project Name",
			Status:  false,
			Message: "Project name must be 63 characters or less",
		}
	}

	return ValidationResult{
		Name:    "Project Name",
		Status:  true,
		Message: "Valid",
	}
}

func ValidateNonEmpty(value string) ValidationResult {
	if strings.TrimSpace(value) == "" {
		return ValidationResult{
			Status:  false,
			Message: "Value is required",
		}
	}
	return ValidationResult{
		Status:  true,
		Message: "Valid",
	}
}
