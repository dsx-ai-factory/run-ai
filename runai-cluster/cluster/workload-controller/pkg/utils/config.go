package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	EnvVarK8sClientQPS    = "K8S_CLIENT_QPS"
	defaultK8sClientQPS   = float32(100)
	EnvVarK8sClientBurst  = "K8S_CLIENT_BURST"
	defaultK8sClientBurst = 500

	EnvVarControlPlanePolicies      = "CONTROL_PLANE_POLICIES_ENABLED"
	EnvVarApplyControlPlanePolicies = "APPLY_CONTROL_PLANE_POLICIES"
)

func GetK8sClientQPS() float32 {
	k8sClientQPS := os.Getenv(EnvVarK8sClientQPS)
	if k8sClientQPS != "" {
		qpsFromEnv, err := strconv.ParseFloat(k8sClientQPS, 32)
		if err != nil {
			log.Error(err, fmt.Sprintf("malformed value for QPS '%s'", k8sClientQPS))
		} else {
			return float32(qpsFromEnv)
		}
	}
	return defaultK8sClientQPS
}

func GetK8sClientBurst() int {
	k8sClientBurst := os.Getenv(EnvVarK8sClientBurst)
	if k8sClientBurst != "" {
		burstFromEnv, err := strconv.ParseInt(k8sClientBurst, 10, 32)
		if err != nil {
			log.Error(err, fmt.Sprintf("malformed value for burst '%s'", k8sClientBurst))
		} else {
			return int(burstFromEnv)
		}
	}
	return defaultK8sClientBurst
}

func IsControlPlanePoliciesEnabled() bool {
	return strings.ToLower(os.Getenv(EnvVarControlPlanePolicies)) == "true"
}

func IsApplyControlPlanePoliciesEnabled() bool {
	return strings.ToLower(os.Getenv(EnvVarApplyControlPlanePolicies)) == "true"
}
