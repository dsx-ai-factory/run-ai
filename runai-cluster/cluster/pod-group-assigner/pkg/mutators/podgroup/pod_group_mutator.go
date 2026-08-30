package podgroup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-pod-group,mutating=true,failurePolicy=fail,sideEffects=None,groups=scheduling.run.ai,resources=podgroups,verbs=create,versions=*,name=mpodgroup.kb.io,admissionReviewVersions=v1

type PodGroupMutator struct{}

func NewPodGroupMutator() *PodGroupMutator {
	return &PodGroupMutator{}
}

func (pgm *PodGroupMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ctx = log.Logger.WithContext(ctx)

	podGroup := &kaiv2alpha2.PodGroup{}
	if err := json.Unmarshal(req.Object.Raw, &podGroup); err != nil {
		err := fmt.Errorf("unable to unmarshal podgroup")
		log.Ctx(ctx).Error().Msgf("Failure in PodGroupMutator Handle, error: <%s>", err.Error())

		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Ctx(ctx).Info().Msgf("PodGroupMutator: handling mutation of pod group <%s/%s>",
		podGroup.Namespace, podGroup.Name)

	pgm.mutatePodGroup(podGroup)

	marshaledPod, err := json.Marshal(podGroup)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("PodGroupMutator: failed to json marshal pod group <%s/%s>, err: <%s>",
			podGroup.Namespace, podGroup.Name, err.Error())

		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (pgm *PodGroupMutator) mutatePodGroup(podGroup *kaiv2alpha2.PodGroup) {
	podGroup.Spec.MarkUnschedulable = ptr.To(false)
	podGroup.Spec.SchedulingBackoff = ptr.To(int32(common.SingleSchedulingBackoff))

	if podGroup.Labels == nil {
		podGroup.Labels = map[string]string{}
	}

	_, found := podGroup.Labels[config.Config().NodePoolLabelKey]
	if found {
		return
	}

	podGroup.Labels[config.Config().NodePoolLabelKey] = config.Config().UnexistingNodepoolSentinel
}
