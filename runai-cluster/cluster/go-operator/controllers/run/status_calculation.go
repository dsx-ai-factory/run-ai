package run

import (
	"context"
	"fmt"
	"strings"

	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
)

func calculateOperandStatusMap(ctx context.Context, runaiConfig *runaiv1.RunaiConfig, ops []operands.Operand) (map[string]runaiv1.TimedOperandStatus, error) {
	operandStatusMap := map[string]runaiv1.TimedOperandStatus{}
	multiErr := &me.MultiError{}
	now := metav1.Now()

	for _, operand := range ops {
		if !operand.Enabled() {
			continue
		}

		previousOperandStatus := runaiConfig.Status.OperandStatuses[operand.Name()]

		status, err := operand.Status(ctx)
		if err != nil {
			multiErr.Append(fmt.Errorf("failed to get status for operand %s: %s", operand.Name(), err.Error()))
			status = operands.NotReadyStatus("failed to get status: %s", err.Error())
		}

		lastTransitionTime := previousOperandStatus.LastTransitionTime
		if previousOperandStatus.Ready != status.Ready || previousOperandStatus.LastTransitionTime.IsZero() {
			lastTransitionTime = now
		}

		operandStatusMap[operand.Name()] = runaiv1.TimedOperandStatus{
			OperandStatus:      status,
			LastTransitionTime: lastTransitionTime,
		}
	}

	return operandStatusMap, multiErr.ErrorOrNil()
}

func calculateAggregatedAvailableStatusReason(operandStatusMap map[string]runaiv1.TimedOperandStatus) (bool, string) {
	aggregatedStatus := aggregatedStatusForOperands(operandStatusMap)
	aggregatedReason := strings.Join(aggregatedStatus.Reasons, "\n")

	return aggregatedStatus.Ready, aggregatedReason
}

func aggregatedStatusForOperands(operandStatusMap map[string]runaiv1.TimedOperandStatus) runaiv1.OperandStatus {
	aggregatedStatus := operands.ReadyStatus()

	for _, opStatus := range operandStatusMap {
		if !opStatus.Ready {
			aggregatedStatus.Ready = false
			aggregatedStatus.Reasons = append(aggregatedStatus.Reasons, opStatus.Reasons...)
		}
	}

	return aggregatedStatus
}
