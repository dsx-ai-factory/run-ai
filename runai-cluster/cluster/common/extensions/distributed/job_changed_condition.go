/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"golang.org/x/exp/maps"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("StatusChangedPoll")

type StatusChangedPoll struct {
	jobExtensionsStatus extensions.Group
	extensionsClient    extensions.Client
}

func NewStatusChangedPoll(watchedJobFramework extensions.Group, jobExtensionClient extensions.Client) *StatusChangedPoll {
	installedFrameworksStatusCopy := make(extensions.Group, len(watchedJobFramework))
	for k, v := range watchedJobFramework {
		installedFrameworksStatusCopy[k] = v
	}
	return &StatusChangedPoll{
		jobExtensionsStatus: installedFrameworksStatusCopy,
		extensionsClient:    jobExtensionClient,
	}

}
func (scp *StatusChangedPoll) Poll(ctx context.Context) (bool, error) {
	jobList := maps.Keys(scp.jobExtensionsStatus)
	currentState, err := scp.extensionsClient.List(ctx, jobList)
	if err != nil {
		log.Error(err, "failed to check if job frameworks status changed")
		log.Info("preserve current state")
		return false, nil
	}
	if !currentState.Equals(scp.jobExtensionsStatus) {
		log.Info("job framework status changed")
		log.V(4).Info("status information",
			"old", scp.jobExtensionsStatus, "new", currentState)
		return true, nil
	}
	log.V(6).Info(
		"job framework status did not change", "state", scp.jobExtensionsStatus)

	return false, nil
}
