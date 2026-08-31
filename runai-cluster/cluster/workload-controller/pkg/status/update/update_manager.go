/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type StatusUpdaterManager struct {
	updater                 StatusUpdater
	preUpdateStatusCallback PreUpdateStatusFunc
	isStatusChangedFunc     IsStatusChanged
}

func NewSimpleStatusUpdaterManager(updater StatusUpdater) *StatusUpdaterManager {
	return &StatusUpdaterManager{
		updater:                 updater,
		isStatusChangedFunc:     IsStatusChangedInWorkload,
		preUpdateStatusCallback: UpdateLastChangedStatus,
	}
}

func NewStatusUpdaterManager(updater StatusUpdater, isStatusChanged IsStatusChanged, preUpdateStatusCallback PreUpdateStatusFunc) *StatusUpdaterManager {
	return &StatusUpdaterManager{
		updater:                 updater,
		isStatusChangedFunc:     isStatusChanged,
		preUpdateStatusCallback: preUpdateStatusCallback,
	}
}

func (self *StatusUpdaterManager) Update(ctx context.Context, modifiedWorkload, originalWorkload *jobs.JobWorkload) error {
	log := logger.FromContext(ctx).WithValues("name", modifiedWorkload.Name, "namespace", modifiedWorkload.Namespace)
	log.V(4).Info("Checking if status is changed")
	if self.isStatusChangedFunc(ctx, modifiedWorkload, originalWorkload) {
		log.V(4).Info("Status is changed, updating status")
		if self.preUpdateStatusCallback != nil {
			log.V(4).Info("Running pre update status function")
			err := self.preUpdateStatusCallback(ctx, modifiedWorkload)
			if err != nil {
				log.Error(err, "Failed to run pre update status callback")
				return err
			}
		}
		log.V(4).Info("Updating status")
		return self.updater.UpdateStatus(ctx, modifiedWorkload)
	}
	log.V(4).Info("Status is not changed, skipping update")
	return nil
}

func UpdateLastChangedStatus(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error {
	log := logger.FromContext(ctx).WithValues("name", modifiedWorkload.Name, "namespace", modifiedWorkload.Namespace)
	timeNow := metav1.Now()
	log.V(4).Info("Updating last status change time")
	log.V(6).Info("Setting last status change time to %v", timeNow.Format(time.RFC3339))
	modifiedWorkload.Status.LastStatusChangeTime = &timeNow
	return nil
}

func IsStatusChangedInWorkload(ctx context.Context, modifiedWorkload, originalWorkload *jobs.JobWorkload) bool {
	log := logger.FromContext(ctx).WithValues("name", modifiedWorkload.Name, "namespace", modifiedWorkload.Namespace)
	log.V(4).Info("Checking if status is changed")
	return modifiedWorkload.IsStatusChanged(originalWorkload)
}
