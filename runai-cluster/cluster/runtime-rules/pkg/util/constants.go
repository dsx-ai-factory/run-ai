/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

const (
	ProjectResource                           = "project"
	KindRunaiJob                              = "RunaiJob"
	NumberOfWorkersPerQueueEnv                = "NUMBER_OF_WORKERS_PER_QUEUE"
	NumberOfWorkersPerQueueDefault            = 20 // the default value
	WorkerSleepSecondsEnv                     = "WORKER_SLEEP_SECONDS"
	WorkerSleepSecondsDefault                 = 5 // the default value
	NumberOfWorkersAutoDeleteCompletedEnv     = "NUMBER_OF_WORKERS_AUTO_DELETE_COMPLETED"
	NumberOfWorkersAutoDeleteCompletedDefault = 4
	EventList                                 = "EVENT_LIST"
)

var (
	EventListDefault = []int{50}
)
