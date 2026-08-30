package post_hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HandleQueueVersionUpgrade is a post-upgrade hook that checks if all queues are updated to the correct version.
// It waits for all queues to be updated with the correct version and then enables the scheduler by enabling the nodepool-controller.
func HandleQueueVersionUpgrade(ctx context.Context, clients *clients.Clients,
	queueValidationIntervalMillis int) error {
	logger := log.FromContext(ctx)

	err := waitForAllQueuesWithCorrectVersion(ctx, clients.RemoteClusterClient,
		queueValidationIntervalMillis)
	if err != nil {
		logger.Error(err, "Failed to wait for all queues to be updated with the correct version")
		return fmt.Errorf("error waiting for all queues to be updated, error: %s", err.Error())
	}

	runaiConfigUpgrader := dynamicupgrades.NewRunaiConfigEnableNodepoolController()
	return runaiConfigUpgrader.Upgrade(ctx, clients)
}

func waitForAllQueuesWithCorrectVersion(ctx context.Context, k8s client.Client,
	queueValidationIntervalMillis int) error {
	logger := log.FromContext(ctx)

	queueValidationInterval := time.Duration(queueValidationIntervalMillis) * time.Millisecond

	ticker := time.NewTicker(queueValidationInterval)
	defer ticker.Stop()

	logger.Info(fmt.Sprintf("Waiting for all queues to be updated to stored version v2, interval %d milliseconds",
		queueValidationIntervalMillis))

	var allUpdated bool
	var queueNamesWithBadSpec []string
	var err error

	for {
		select {
		case <-ticker.C:
			allUpdated, queueNamesWithBadSpec, err = areForAllQueuesWithCorrectVersion(ctx, k8s)
			if err != nil {
				return fmt.Errorf("error checking queues: %w", err)
			}
			if allUpdated {
				logger.Info("All queues are updated with the correct version")
				return nil
			}
			logger.Info(fmt.Sprintf("Not all queues are updated yet, queues with bad spec: %v", queueNamesWithBadSpec))
		case <-ctx.Done():
			return fmt.Errorf("timeout reached while waiting for all queues to be updated, queues with bed spec: %v",
				queueNamesWithBadSpec)
		}
	}
}

func areForAllQueuesWithCorrectVersion(ctx context.Context, k8s client.Client) (bool, []string, error) {
	queues := &schedv2.QueueList{}
	err := k8s.List(ctx, queues)
	if err != nil {
		return false, []string{}, fmt.Errorf("failed to list queues: %w", err)
	}

	queueNamesWithBadSpec := []string{}
	for _, queue := range queues.Items {
		if _, ok := queue.Labels[constants.RunaiProjectLabel]; !ok {
			// Skip non-project queues
			continue
		}

		if queue.Spec.ParentQueue == "" {
			queueNamesWithBadSpec = append(queueNamesWithBadSpec, queue.Name)
		}
	}

	if len(queueNamesWithBadSpec) > 0 {
		return false, queueNamesWithBadSpec, nil
	}

	return true, []string{}, nil
}
