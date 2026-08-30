package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// WebhookPath is the HTTP path this mutating webhook is served on.
const WebhookPath = "/mutate-pod"

//+kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=*,name=mpod.kb.io,admissionReviewVersions=v1,reinvocationPolicy=IfNeeded

// PodMutator mutates pods on creation on behalf of the resource-management package:
// it enforces the scheduler, labels the pod with its project, and applies the project's
// default node pools as required node affinity. The three go together - they are the
// package's single admission-time contract for a pod - and are enabled as a unit by
// --enable-pod-webhook.
type PodMutator struct {
	client client.Client
}

func NewPodMutator(k8sClient client.Client) *PodMutator {
	return &PodMutator{client: k8sClient}
}

func (pm *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ctx = log.Logger.WithContext(ctx)

	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		webhookErr := fmt.Errorf("unable to unmarshal pod: %s", err.Error())
		log.Ctx(ctx).Error().Msgf("Failure in PodMutator Handle, error: <%s>", webhookErr.Error())

		return admission.Errored(http.StatusInternalServerError, webhookErr)
	}

	// On CREATE the pod is not yet persisted, so pod.Namespace may be empty.
	namespace := req.Namespace
	if namespace == "" {
		namespace = pod.Namespace
	}

	// A namespace read failure is NOT swallowed - enforcement decides whether a pod is
	// scheduled by KAI at all, and SchedulerName is immutable after creation, so a pod
	// admitted without it can never be corrected.
	enforced, err := pm.isSchedulerEnforcedForNamespace(ctx, namespace)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("PodMutator: failed to resolve scheduler enforcement for pod <%s/%s>, err: <%s>",
			namespace, pod.Name, err.Error())

		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !enforced {
		return admission.Allowed("pod is not managed by the resource-management package")
	}

	log.Ctx(ctx).Info().Msgf("PodMutator: handling mutation of pod <%s/%s>", namespace, pod.Name)

	pm.mutateSchedulerName(ctx, pod, namespace)
	pm.mutateProjectNameLabel(ctx, pod, namespace)
	pm.mutateNodeAffinityForDefaultNodePools(ctx, pod, namespace)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("PodMutator: failed to json marshal pod <%s/%s>, err: <%s>",
			namespace, pod.Name, err.Error())

		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}
