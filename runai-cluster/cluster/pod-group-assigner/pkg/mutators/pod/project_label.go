package pod

import (
	"context"

	"github.com/rs/zerolog/log"

	kaipgconstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/utils"

	corev1 "k8s.io/api/core/v1"
)

// mutateProjectNameLabel records the pod's project on the pod itself, resolved from its
// namespace. An explicit label always wins - a workload builder that already set one knows
// better than the namespace default.
//
// It is fail-open, like the default-node-pools mutation: a namespace with no project label
// is logged and the pod left untouched, so a misconfiguration never blocks pod creation.
func (pm *PodMutator) mutateProjectNameLabel(ctx context.Context, pod *corev1.Pod, namespace string) {
	if _, found := pod.Labels[kaipgconstants.ProjectLabelKey]; found {
		return
	}

	projectName, err := utils.GetProjectNameOfNamespace(ctx, pm.client, namespace)
	if err != nil {
		log.Ctx(ctx).Info().Msgf(
			"could not identify the project related to pod <%s/%s> (err: <%s>); skipping project label mutation",
			namespace, pod.Name, err.Error())
		return
	}

	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[kaipgconstants.ProjectLabelKey] = projectName

	log.Ctx(ctx).Info().Msgf("labeled pod <%s/%s> with <%s: %s>",
		namespace, pod.Name, kaipgconstants.ProjectLabelKey, projectName)
}
