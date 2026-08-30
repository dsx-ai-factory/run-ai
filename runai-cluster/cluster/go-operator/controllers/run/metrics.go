package run

import (
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/metrics"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

func setMetrics(runaiConfig *runaiv1.RunaiConfig, operandStatusMap map[string]runaiv1.TimedOperandStatus) {
	setOperandStatusMetrics(operandStatusMap)
	setClusterInfoMetric(runaiConfig)
}

func setOperandStatusMetrics(operandStatusMap map[string]runaiv1.TimedOperandStatus) {
	for operand, status := range operandStatusMap {
		metrics.OperandStatus.WithLabelValues(operand, metrics.OperandStatusConditionReady).Set(boolToFloat64(status.Ready))
	}
}

func setClusterInfoMetric(runaiConfig *runaiv1.RunaiConfig) {
	tenantID := ""
	if runaiConfig.Spec.Internal.Global.ControlPlane.Config.TenantId != nil {
		tenantID = strconv.FormatInt(int64(*runaiConfig.Spec.Internal.Global.ControlPlane.Config.TenantId), 10)
	}

	clusterName := ""
	if runaiConfig.Spec.Internal.Global.ControlPlane.Config.ClusterName != nil {
		clusterName = *runaiConfig.Spec.Internal.Global.ControlPlane.Config.ClusterName
	}

	platformType := ""
	if runaiConfig.Spec.Internal.Global.Platform != nil {
		platformType = strings.ToLower((*runaiConfig.Spec.Internal.Global.Platform).KubernetesFlavor.String())
	}

	metrics.ClusterInfo.WithLabelValues(
		tenantID,
		runaiConfig.Spec.Internal.Global.ControlPlane.ClusterUUID,
		clusterName,
		runaiConfig.Spec.Internal.Global.Version,
		platformType,
	).Set(1)
}
