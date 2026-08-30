package pod

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// isSchedulerEnforcedForNamespace reports whether every pod of this namespace must be
// scheduled by our scheduler. project-controller writes the annotation from
// Project.Spec.EnforceKaiScheduler; an absent or unparseable value means "not enforced",
// so a malformed annotation never forces a scheduler onto a pod by accident.
//
// Only a failure to read the namespace is returned as an error - the caller rejects the
// admission request rather than guess.
func (pm *PodMutator) isSchedulerEnforcedForNamespace(ctx context.Context, namespaceName string) (bool, error) {
	namespace := &corev1.Namespace{}
	if err := pm.client.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace); err != nil {
		return false, fmt.Errorf("failed to get namespace <%s>, error: %s", namespaceName, err.Error())
	}

	enforce, found := namespace.Annotations[config.Config().EnforceSchedulerAnnotationKey]
	if !found {
		return false, nil
	}

	enforced, err := strconv.ParseBool(enforce)
	if err != nil {
		log.Ctx(ctx).Error().Msgf(
			"Could not parse annotation <%s> value <%s> of namespace <%s>, treating as not enforced, err: <%s>",
			config.Config().EnforceSchedulerAnnotationKey, enforce, namespaceName, err.Error())
		return false, nil
	}

	return enforced, nil
}

// mutateSchedulerName forces the configured scheduler onto the pod. This is the whole point
// of enforcement: without it the pod goes to the default scheduler and no KAI component -
// podgrouper, binder, or the scheduler itself - will ever look at it, so the pod runs outside
// its project's quota.
func (pm *PodMutator) mutateSchedulerName(ctx context.Context, pod *corev1.Pod, namespace string) {
	schedulerName := config.Config().SchedulerName
	if pod.Spec.SchedulerName == schedulerName {
		return
	}

	log.Ctx(ctx).Info().Msgf("Enforcing scheduler <%s> on pod <%s/%s> (was <%s>)",
		schedulerName, namespace, pod.Name, pod.Spec.SchedulerName)

	pod.Spec.SchedulerName = schedulerName
}
