/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runaijobcontroller

import (
	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsJobFinished checks whether the given RunaiJob has finished execution.
// It does not discriminate between successful and failed terminations.
func IsJobFinished(j *runaijob.RunaiJob) bool {
	for _, c := range j.Status.Conditions {
		if (c.Type == runaijob.JobComplete || c.Type == runaijob.JobFailed) && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func GetKey(job *runaijob.RunaiJob) ctrl.Request {
	return ctrl.Request{
		NamespacedName: client.ObjectKeyFromObject(job),
	}
}
