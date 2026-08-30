package common

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/common/priority"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsPreemptible(ctx context.Context, readerClient client.Reader,
	specPriorityClassName string, specPreemptible bool, workloadKind string) (bool, error) {
	if specPriorityClassName == "" {
		if isPreemptibleSpecRelevant(workloadKind) && specPreemptible {
			return true, nil
		}

		// TODO: use default preemptible value from config
		return false, nil
	}

	return priority.IsPreemptible(ctx, readerClient, specPriorityClassName)
}
