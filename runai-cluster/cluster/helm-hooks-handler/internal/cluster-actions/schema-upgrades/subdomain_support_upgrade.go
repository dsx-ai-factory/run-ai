package schema_upgrades

import (
	"strings"

	semver "github.com/Masterminds/semver/v3"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"k8s.io/utils/ptr"
)

var subdomainSupportIntroducedVersion = semver.MustParse("2.25.0")

type subdomainSupportUpgrade struct {
	previousVersion string
}

func newSubdomainSupportUpgrader(previousVersion string) *subdomainSupportUpgrade {
	return &subdomainSupportUpgrade{previousVersion: previousVersion}
}

func (u *subdomainSupportUpgrade) shouldPerformUpgrade(resource *runaiv1.RunaiConfig) bool {
	if !u.isUpgradeFromBeforeSubdomainDefault() {
		return false
	}
	if resource.Spec.Global == nil {
		return true
	}
	return resource.Spec.Global.SubdomainSupport == nil
}

func (u *subdomainSupportUpgrade) updateRunaiConfig(resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error) {
	if resource.Spec.Global == nil {
		resource.Spec.Global = &global.Global{}
	}
	resource.Spec.Global.SubdomainSupport = ptr.To(false)
	return resource, nil
}

func (u *subdomainSupportUpgrade) upgradeHandlerIdentifier() string {
	return "subdomain-support"
}

func (u *subdomainSupportUpgrade) isUpgradeFromBeforeSubdomainDefault() bool {
	if u.previousVersion == "" {
		return true
	}
	if strings.HasPrefix(u.previousVersion, "0.0.0") {
		return false
	}
	v, err := semver.NewVersion(u.previousVersion)
	if err != nil {
		return true
	}
	return v.LessThan(subdomainSupportIntroducedVersion)
}
