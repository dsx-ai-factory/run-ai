/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package frameworks

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var DistributedJobCRDsSupportedMap = map[string]client.Object{
	constants.KindMpiJob: &v2beta1.MPIJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindMpiJob,
			APIVersion: "kubeflow.org/v2beta1",
		},
	},
	constants.KindPyTorchJob: &kubeflowv1.PyTorchJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindPyTorchJob,
			APIVersion: kubeflowv1.GroupVersion.String(),
		},
	},
	constants.KindTFJob: &kubeflowv1.TFJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindTFJob,
			APIVersion: kubeflowv1.GroupVersion.String(),
		},
	},
	constants.KindXGBoostJob: &kubeflowv1.XGBoostJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindXGBoostJob,
			APIVersion: kubeflowv1.GroupVersion.String(),
		},
	},
	constants.KindJAXJob: &kubeflowv1.JAXJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindJAXJob,
			APIVersion: kubeflowv1.GroupVersion.String(),
		},
	},
}

type DistributedJobFrameworksStatus struct {
	JobObject      client.Object
	SuspendEnabled bool
}

func GetDistributedJobFrameworksStatus(extensionsGroupStatus extensions.Group) (map[string]DistributedJobFrameworksStatus, error) {
	enabledJobFrameworks := make(map[string]DistributedJobFrameworksStatus, len(extensionsGroupStatus))

	for _, extension := range extensionsGroupStatus {
		if !extension.IsEnabled() {
			continue
		}
		jobKind, ok := distributed.ExtensionToJobKindMap[extension.GetType()]
		if !ok {
			return nil, fmt.Errorf("the job kind for the extension %s were not found", extension.GetType().String())
		}
		jobObject, ok := DistributedJobCRDsSupportedMap[jobKind]
		if !ok {
			return nil, fmt.Errorf("the job object for the extension %s were not found", extension.GetType().String())
		}
		enabledJobFrameworks[jobKind] = DistributedJobFrameworksStatus{
			jobObject,
			extension.IsCapable(distributed.Suspension),
		}
	}

	return enabledJobFrameworks, nil
}
