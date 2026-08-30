package workload

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
)

const (
	KubeflowReplicaTypeLabel     = "training.kubeflow.org/replica-type"
	KubeflowReplicaMasterMPI     = "launcher"
	KubeflowReplicaMasterMX      = "master"
	KubeflowReplicaMasterPaddle  = "master"
	KubeflowReplicaMasterPytorch = "master"
	KubeflowReplicaMasterTF1     = "master"
	KubeflowReplicaMasterTF2     = "chief"
	KubeflowReplicaMasterXgBoost = "master"

	MpiOperatorJobRoleLabel   = "training.kubeflow.org/job-role"
	MpiOperatorMasterJobRole  = "launcher"
	MpiOperatorRoleTypeLabel  = "mpi_role_type"
	MpiOperatorRoleTypeMaster = "launcher"

	RayPodTypeLabel  = "ray.io/node-type"
	RayMasterPodType = "head"
)

var kubeflowMasterTypes = []string{KubeflowReplicaMasterMPI, KubeflowReplicaMasterMX, KubeflowReplicaMasterPaddle,
	KubeflowReplicaMasterPytorch, KubeflowReplicaMasterTF1, KubeflowReplicaMasterTF2, KubeflowReplicaMasterXgBoost}

func CalcLatestChiefPod(pgPods v1.PodList, workloadPg kaiv2alpha2.PodGroup, currentLatestChiefPod *v1.Pod) *v1.Pod {
	for _, podItem := range pgPods.Items {
		pgPod := podItem
		// For distributed workloads, worker pods can't be a chief pod.
		// As of KAI main, PodGroupSpec.MinMember is a *int32 (nil means unset).
		if workloadPg.Spec.MinMember != nil && *workloadPg.Spec.MinMember > 1 && !isPodDetectedAsMasterPod(pgPod) {
			continue
		}

		if currentLatestChiefPod == nil || pgPod.CreationTimestamp.Before(&currentLatestChiefPod.CreationTimestamp) {
			currentLatestChiefPod = &pgPod
		}
	}
	return currentLatestChiefPod
}

func isPodDetectedAsMasterPod(pod v1.Pod) bool {
	return isMpiOperatorMaster(pod) ||
		isKubeflowMasterRole(pod) ||
		isRayMasterPod(pod)
}

func isMpiOperatorMaster(pod v1.Pod) bool {
	roleType, hasMpiRoleType := pod.ObjectMeta.Labels[MpiOperatorRoleTypeLabel]
	if hasMpiRoleType && roleType == MpiOperatorRoleTypeMaster {
		return true
	}

	roleType, hasMpiJobRoleType := pod.ObjectMeta.Labels[MpiOperatorJobRoleLabel]
	if hasMpiJobRoleType {
		return roleType == MpiOperatorMasterJobRole
	}

	return false
}

func isKubeflowMasterRole(pod v1.Pod) bool {
	roleType, hasKubeflowRoleType := pod.ObjectMeta.Labels[KubeflowReplicaTypeLabel]
	if hasKubeflowRoleType {
		return slices.Contains(kubeflowMasterTypes, roleType)
	}

	return false
}

func isRayMasterPod(pod v1.Pod) bool {
	labelValue, hasRayPodType := pod.ObjectMeta.Labels[RayPodTypeLabel]
	if hasRayPodType {
		return labelValue == RayMasterPodType
	}

	return false
}
