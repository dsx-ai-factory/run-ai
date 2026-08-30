package utils

import (
	"fmt"
	semver "github.com/hashicorp/go-version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = log.Log.WithName("feature-gate")

// FeatureGateUtil is a utility for handling Kubernetes feature gates based on the server version.
type FeatureGateUtil struct {
	client                              *discovery.DiscoveryClient
	kubeVersion                         string
	initOnce                            sync.Once
	featureGateSidecarContainersVersion string
	FeatureGateSidecarContainersEnable  bool
}

var instance *FeatureGateUtil
var once sync.Once

const UNDEFINED int = -2

func GetFeatureGateUtilInstance() *FeatureGateUtil {
	once.Do(func() {
		instance = &FeatureGateUtil{
			// featureGateSidecarContainers is supported starting from Kubernetes version 1.29
			featureGateSidecarContainersVersion: "1.29",
			FeatureGateSidecarContainersEnable:  false,
		}
	})
	return instance
}

func (f *FeatureGateUtil) Initialize(cfg *rest.Config) {
	f.initOnce.Do(func() {
		var err error
		f.client, err = discovery.NewDiscoveryClientForConfig(cfg)
		if err != nil {
			logger.Error(err, "Err in creating DiscoveryClient")
			return
		}

		version, err := f.client.ServerVersion()
		if err != nil {
			logger.Error(err, "Err in getting kubeVersion")
			return
		}
		f.kubeVersion = version.String()
		logger.Info("Identify server k8s version", "kubeVersion", f.kubeVersion)

		// set all featureGatedFlags
		f.setFeatureGateSidecarContainersEnable()

	})
}

func (f *FeatureGateUtil) setFeatureGateSidecarContainersEnable() {
	kubeVersionComparisonResult, err := f.parseAndCompare(f.kubeVersion, f.featureGateSidecarContainersVersion)
	if err != nil {
		logger.Error(err, "Err in comparing featureGateSidecarContainer to kubeVersion")
	}

	if kubeVersionComparisonResult >= 0 {
		f.FeatureGateSidecarContainersEnable = true
	}
}

// parseAndCompare compares two semantic version strings and returns the comparison result.
// It returns 0 if version1 is equal to version2, 1 if version1 is greater than version2, and -1 if version1 is less than version2.
// If either version string is invalid, it returns -2 along with an error.
func (f *FeatureGateUtil) parseAndCompare(version1, version2 string) (int, error) {
	ver1, err := semver.NewVersion(version1)
	if err != nil {
		return UNDEFINED, fmt.Errorf("invalid version1 %s", err)
	}
	ver2, err := semver.NewVersion(version2)
	if err != nil {
		return UNDEFINED, fmt.Errorf("invalid version1 %s", err)
	}
	return ver1.Compare(ver2), nil
}
