/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

import (
	"errors"
)

type WorkloadType string

const (
	InteractiveWorkload WorkloadType = "Interactive"
	TrainingWorkload    WorkloadType = "Training"
	DistributedWorkload WorkloadType = "Distributed"
)

func (workloadTypeStr WorkloadType) String() string {
	return string(workloadTypeStr)
}

type ReplicaType int

const (
	Worker ReplicaType = iota
	Master
	Chief
	Launcher
)
const (
	MasterStr   = "Master"
	WorkerStr   = "Worker"
	ChiefStr    = "Chief"
	LauncherStr = "Launcher"
)

func (replicaTypeStr ReplicaType) String() string {
	return [...]string{WorkerStr, MasterStr, ChiefStr, LauncherStr}[replicaTypeStr]
}

func ReplicaTypeFromString(replicaType string) (ReplicaType, error) {
	switch replicaType {
	case WorkerStr:
		return Worker, nil
	case MasterStr:
		return Master, nil
	case ChiefStr:
		return Chief, nil
	case LauncherStr:

		return Launcher, nil
	default:
		return 0, errors.New("unknown replica type: " + replicaType)
	}
}
