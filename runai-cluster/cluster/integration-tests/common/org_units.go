package common

import (
	"strconv"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Org-unit fixtures for the suites. The services under test reconcile and read the
// kai.resources Project and Department, so those are the only org units the suites create.

// SecretsReadyConditionType is the project status condition the secrets deletion blocker
// reports. It is derived by the project-controller from the "Secrets" blocker group's
// display name, so it is not available as a constant on the CRD.
const SecretsReadyConditionType kaiv1alpha1.ProjectConditionType = "SecretsReady"

// BuildProjectObject builds a project with a single queue on the default node pool.
// The project and department id labels keep the run.ai keys: they are the ones the
// consumers (e.g. the external-workload-integrator) read off the kai project.
func BuildProjectObject(projectName, departmentName, projectId, departmentId string) *kaiv1alpha1.Project {
	return &kaiv1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				v2.ProjectIdLabel:                    projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent: departmentName,
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:      projectName,
					Nodepool:  v1alpha1.DefaultNodePoolName,
					Resources: BuildQueueResources(1, 1000, 10000),
				},
			},
		},
	}
}

// BuildDepartmentObject builds a department with a single queue on the default node pool.
func BuildDepartmentObject(departmentName, departmentId string) *kaiv1alpha1.Department {
	return &kaiv1alpha1.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			Labels: map[string]string{
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kaiv1alpha1.DepartmentSpec{
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:      departmentName,
					Nodepool:  v1alpha1.DefaultNodePoolName,
					Resources: BuildQueueResources(1, 1000, 10000),
				},
			},
		},
	}
}

// BuildQueueResources builds the resources of a project or department queue.
func BuildQueueResources(gpuDeserved, cpuDeserved, memoryDeserved int64) *kaiv1alpha1.QueueResourcesConfig {
	return &kaiv1alpha1.QueueResourcesConfig{
		GPU: kaiv1alpha1.SystemResource{
			Deserved: float64(gpuDeserved),
			Limit:    -1,
		},
		CPU: kaiv1alpha1.SystemResource{
			Deserved: float64(cpuDeserved),
			Limit:    -1,
		},
		Memory: kaiv1alpha1.SystemResource{
			Deserved: float64(memoryDeserved),
			Limit:    -1,
		},
	}
}

// SetProjectDeletion sets how the project is to be deleted. blocking makes the
// project-controller run its deletion blockers (the resources remaining in the project
// hold the deletion back); force makes it finalize the project even when a blocker
// reports it is not empty, and lives on an annotation rather than in the spec.
func SetProjectDeletion(project *kaiv1alpha1.Project, blocking, force bool) {
	if blocking {
		blockingType := kaiv1alpha1.Blocking
		project.Spec.DeletionType = &blockingType
	} else {
		project.Spec.DeletionType = nil
	}

	annotations := project.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	if force {
		annotations[clusterconstants.ForceDeleteAnnotation] = strconv.FormatBool(true)
	} else {
		delete(annotations, clusterconstants.ForceDeleteAnnotation)
	}
	project.SetAnnotations(annotations)
}
