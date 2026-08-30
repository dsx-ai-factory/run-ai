package config

import (
	"fmt"
	"strconv"

	"os"

	log "github.com/sirupsen/logrus"
)

const (
	EnvVarK8sClientQPS    = "K8S_CLIENT_QPS"
	defaultK8sClientQPS   = float32(50)
	EnvVarK8sClientBurst  = "K8S_CLIENT_BURST"
	defaultK8sClientBurst = 300
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
