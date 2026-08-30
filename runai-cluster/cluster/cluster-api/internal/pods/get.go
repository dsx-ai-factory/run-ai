package pods

import (
	"context"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPodsWithSelector(
	ctx context.Context, k8sClient client.Client, selector client.ListOption,
) ([]corev1.Pod, error) {
	log.Debug().Msg("Getting pods with selector")

	podList := &corev1.PodList{}

	err := k8sClient.List(ctx, podList, selector)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info().Type("selector", selector).Msg("No pods found with the given selector")
			return []corev1.Pod{}, nil
		}

		log.Error().Err(err).Msg("Error listing pods")

		return nil, err
	}

	log.Info().Int("podCount", len(podList.Items)).Msg("Retrieved pods")

	return podList.Items, nil
}
