package pod_identifier

import (
	"strconv"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func GetSharedGpuPriority(podAnnotations map[string]string) (bool, int, error) {
	value, found := podAnnotations[types.RunaiGpuPriority]
	if !found {
		return false, -1, nil
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return false, -1, err
	}

	return true, i, nil
}
