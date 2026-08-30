package clusterstatecalculator

import (
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

type ClusterStateCalculator interface {
	CalculateState(cluster *models.Cluster) ClusterState
}

type clusterStateCalculatorImp struct{}

func NewClusterStateCalculator() ClusterStateCalculator {
	return &clusterStateCalculatorImp{}
}

func (c *clusterStateCalculatorImp) CalculateState(cluster *models.Cluster) ClusterState {
	// Cluster states precedence (as defined by @gal-revach):
	// 1. WaitingToConnect
	// 2. Disconnected
	// 3. MissingPrerequisites
	// 4. ServiceIssues
	// 5. Connected

	connectionState := calcClusterConnectionState(cluster)
	if connectionState != ClusterStateConnected {
		return connectionState
	}

	if isMissingPrerequisites(cluster) {
		return ClusterStateMissingPrerequisites
	}

	if isServiceIssues(cluster) {
		return ClusterStateServiceIssues
	}

	return ClusterStateConnected
}

func isMissingPrerequisites(cluster *models.Cluster) bool {
	for _, dependency := range cluster.Status.Dependencies.Required {
		if !dependency.Available {
			return true
		}
	}

	return false
}

func isServiceIssues(cluster *models.Cluster) bool {
	for _, operand := range cluster.Status.Operands {
		if operand.UnreadyThresholdCrossed {
			return true
		}
	}

	return false
}

func calcClusterConnectionState(cluster *models.Cluster) ClusterState {
	if cluster.LastLiveness == nil {
		return ClusterStateWaitingToConnect
	}

	appConfig := runai_config.Get[*config.App]()
	if time.Since(*cluster.LastLiveness) > appConfig.ClusterConnectivityThreshold {
		return ClusterStateDisconnected
	}

	return ClusterStateConnected
}
