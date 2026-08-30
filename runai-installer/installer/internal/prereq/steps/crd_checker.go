package steps

import "k8s.io/client-go/rest"

// CRDChecker is an interface for checking if a CRD is installed on the cluster
type CRDChecker interface {
	IsCRDInstalled(cfg *rest.Config, crdName string) bool
}

// DefaultCRDChecker implements CRDChecker using the existing function
type DefaultCRDChecker struct{}

func (d *DefaultCRDChecker) IsCRDInstalled(cfg *rest.Config, crdName string) bool {
	return IsCrdsInstalledInstalled(cfg, crdName)
}
