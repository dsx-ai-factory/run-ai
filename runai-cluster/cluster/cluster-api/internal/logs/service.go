package logs

import (
	"context"
	"fmt"
	"io"

	clustercontainer "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/container"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/service"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	v1 "k8s.io/api/core/v1"
)

type Streamer interface {
	Stream(ctx context.Context, namespace string, options *Options) (io.ReadCloser, error)
}

type Service struct {
	podsGetter  service.PodsGetter
	podStreamer Streamer
}

func NewService(podsGetter service.PodsGetter, podStreamer Streamer) *Service {
	log.Debug().Msg("Creating new Service")

	s := &Service{
		podsGetter:  podsGetter,
		podStreamer: podStreamer,
	}

	log.Info().Msg("Logs service created successfully")

	return s
}

func (s *Service) LogsReader(
	ctx context.Context, workload *workload.Workload, options *Options,
) (io.ReadCloser, error) {
	log.Debug().Str("workloadName", workload.Name).Str("namespace", workload.Namespace).
		Str("podName", options.Pod).Msg("Starting - Getting pod")

	pod, err := s.podsGetter.GetPod(ctx, workload, options.Pod)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pod")
		return nil, err
	}

	container, err := clustercontainer.GetContainer(pod, options.Container)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get container")
		return nil, err
	}

	log.Debug().Str("container", container).Msg("Container selected")

	builder := NewOptionsBuilder().FromOptions(*options)
	builder.WithPod(pod.Name)
	builder.WithContainer(container)

	finalOption, err := builder.Build()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build final options")
		return nil, err
	}

	log.Debug().Interface("finalOptions", finalOption).Msg("Final options built")

	err = validate(pod, finalOption)
	if err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return nil, err
	}

	reader, err := s.getLogs(ctx, workload.Namespace, finalOption)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get logs")
		return nil, err
	}

	log.Debug().Msg("Successfully obtained logs reader")

	// update origin options with pod and container (needed for logs download file name)
	options.Pod = finalOption.Pod
	options.Container = finalOption.Container

	return reader, nil
}

func validate(pod *v1.Pod, options Options) error {
	log.Debug().Str("podName", pod.Name).Msg("Starting validation")

	if err := validatePod(pod, options.Previous); err != nil {
		log.Error().Err(err).Msg("Pod validation failed")
		return err
	}

	if err := ValidateContainer(pod, options); err != nil {
		log.Error().Err(err).Msg("Container validation failed")
		return err
	}

	log.Debug().Msg("Validation successful")

	return nil
}

func validatePod(pod *v1.Pod, usePrevious bool) error {
	log.Debug().Str("podName", pod.Name).Bool("usePrevious", usePrevious).Msg("Starting pod validation")

	if pod.Status.Phase == v1.PodPending {
		log.Warn().Str("podName", pod.Name).Msg("Pod is not ready")
		return api.NewNotReadyError("pod is not ready")
	}

	log.Debug().Str("podName", pod.Name).Msg("Pod validation successful")

	return nil
}

func (s *Service) getLogs(ctx context.Context, namespace string, options Options) (io.ReadCloser, error) {
	log.Debug().Str("namespace", namespace).Interface("options", options).Msg("Starting to stream logs")

	reader, err := s.podStreamer.Stream(ctx, namespace, &options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to stream logs")
		return nil, err
	}

	log.Debug().Msg("Successfully started streaming logs")

	return reader, nil
}

func ValidateContainer(pod *v1.Pod, options Options) error {
	log.Debug().Str("podName", pod.Name).Str("container", options.Container).Msg("Starting container validation")
	// If the user requested a specific container, we will validate that the container exists in the pod
	if options.Container == "" {
		// should not happen
		return fmt.Errorf("container name is required")
	}

	if options.Previous {
		log.Debug().Str("container", options.Container).Msg("Skipping readiness check for previous logs")
		return nil
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == options.Container {
			return clustercontainer.ValidateContainerExist(pod.Status.ContainerStatuses, container.Name)
		}
	}

	for _, container := range pod.Spec.InitContainers {
		if container.Name == options.Container {
			return clustercontainer.ValidateContainerExist(pod.Status.InitContainerStatuses, container.Name)
		}
	}

	for _, container := range pod.Spec.EphemeralContainers {
		if container.Name == options.Container {
			return clustercontainer.ValidateContainerExist(pod.Status.EphemeralContainerStatuses, container.Name)
		}
	}

	log.Error().Str("container", options.Container).Msg("Requested container not found")

	return api.NewNotFoundError("the requested container not found")
}
