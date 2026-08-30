package runaipublic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FrameworkNotFound = errors.New("framework not found")
)

const (
	ErrorMissingDependencies    = "runai-public config, is missing \"Dependencies\" field"
	ErrorMissingOptional        = "runai-public config, is missing \"Optional\" field"
	ErrorRunAIPublicEmpty       = "failed to parse runai-public, got an empty config"
	ErrorParsingRunAIPublic     = "failed to parse runai-public, %v"
	ErrorFailedToGetClusterData = "failed to get cluster %s, data"
	KnativeKey                  = "knative"
	LwsKey                      = "lws"
)

type Handler struct {
	ctx         context.Context
	k8sClient   client.Reader
	runaiPublic *runv1.RunaiPublic
}

func NewRunaiPublicHandler(ctx context.Context, k8sClient client.Reader) *Handler {
	return &Handler{
		ctx:       ctx,
		k8sClient: k8sClient,
	}
}

func (h *Handler) getRunaiPublicConfigMap() error {
	runaiPublicConfigMap := corev1.ConfigMap{}
	err := h.k8sClient.Get(h.ctx, client.ObjectKey{
		Namespace: constants.NamespaceRunAi,
		Name:      constants.RunaiPublicConfigMapName,
	}, &runaiPublicConfigMap)
	if err != nil {
		return err
	}
	return h.generateRunAIPublic(&runaiPublicConfigMap)
}

func (h *Handler) generateRunAIPublic(runaiPublicConfigMap *corev1.ConfigMap) error {
	if runaiPublicJson, ok := runaiPublicConfigMap.Data[constants.RunaiPublicConfigMapName]; ok {
		if len(strings.TrimSpace(runaiPublicJson)) == 0 {
			return fmt.Errorf(ErrorRunAIPublicEmpty)
		}
		err := yaml.Unmarshal([]byte(runaiPublicJson), &h.runaiPublic)
		if err != nil {
			return fmt.Errorf(ErrorParsingRunAIPublic, err)
		}
		return nil
	}
	return fmt.Errorf(ErrorFailedToGetClusterData, constants.RunaiPublicConfigMapName)
}

func (h *Handler) IsKnativeAvailable() (bool, string, error) {
	if h.runaiPublic == nil {
		err := h.getRunaiPublicConfigMap()
		if err != nil {
			return false, "", fmt.Errorf("Invalid workload: %s", err.Error())
		}
	}

	if h.runaiPublic.Dependencies == nil {
		return false, "", fmt.Errorf(ErrorMissingDependencies)
	}
	if h.runaiPublic.Dependencies.Optional == nil {
		return false, "", fmt.Errorf(ErrorMissingOptional)
	}

	knativeAvailableStatus := h.runaiPublic.Dependencies.Optional[KnativeKey]
	jobAvailable := ptr.Deref(knativeAvailableStatus.Available, false)

	return jobAvailable, knativeAvailableStatus.Reason, nil
}

func (h *Handler) IsLwsAvailable() (bool, string, error) {
	if h.runaiPublic == nil {
		err := h.getRunaiPublicConfigMap()
		if err != nil {
			return false, "", fmt.Errorf("Invalid workload: %s", err.Error())
		}
	}

	if h.runaiPublic.Dependencies == nil {
		return false, "", fmt.Errorf(ErrorMissingDependencies)
	}
	if h.runaiPublic.Dependencies.Optional == nil {
		return false, "", fmt.Errorf(ErrorMissingOptional)
	}

	lwsAvailableStatus := h.runaiPublic.Dependencies.Optional[LwsKey]
	jobAvailable := ptr.Deref(lwsAvailableStatus.Available, false)

	return jobAvailable, lwsAvailableStatus.Reason, nil
}

func (h *Handler) IsDistributedJobAvailable(jobType string) (bool, string, bool, error) {
	if h.runaiPublic == nil {
		err := h.getRunaiPublicConfigMap()
		if err != nil {
			return false, "", false, fmt.Errorf("invalid workload: %w", err)
		}
	}

	if h.runaiPublic.Dependencies == nil {
		return false, "", false, fmt.Errorf(ErrorMissingDependencies)
	}
	if h.runaiPublic.Dependencies.Optional == nil {
		return false, "", false, fmt.Errorf(ErrorMissingOptional)
	}
	jobAvailableStatus, ok := h.runaiPublic.Dependencies.Optional[jobType]
	if !ok {
		return false, "", false, fmt.Errorf("job type %s is not supported, %w", jobType, FrameworkNotFound)
	}

	jobAvailable := ptr.Deref(jobAvailableStatus.Available, false)

	jobSuspendable := false
	if suspendComponent, ok := jobAvailableStatus.Components["suspend"]; ok {
		jobSuspendable = ptr.Deref(suspendComponent.Available, false)
	}

	return jobAvailable, jobAvailableStatus.Reason, jobSuspendable, nil

}
