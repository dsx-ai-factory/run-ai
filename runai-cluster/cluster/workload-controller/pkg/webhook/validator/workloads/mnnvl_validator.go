package workloads

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	apijobs "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

var (
	mnnvlUnsupportedErrPrefix = fmt.Sprintf("%s is unsupported for", mnnvl.UserFacingFeatureName)
)

type MnnvlValidator struct {
	clusterMnnvlEnabled     bool
	clusterElasticCDEnabled bool
}

func NewMnnvlValidator(clusterSupportMNNVL bool, clusterSupportElasticCD bool) *MnnvlValidator {
	return &MnnvlValidator{
		clusterMnnvlEnabled:     clusterSupportMNNVL,
		clusterElasticCDEnabled: clusterSupportElasticCD,
	}
}

func (mv *MnnvlValidator) Validate(_ context.Context, job *apijobs.JobWorkload) error {
	if job.Kind != constants.KindDistributedWorkload {
		return nil
	}
	featureLabelValue, exists := job.Labels[mnnvl.FeatureLabelKey]
	if !exists || featureLabelValue == mnnvl.None {
		return nil
	}

	// check that the feature label value can later be converted to a valid mode
	if _, err := mnnvl.GetMode(featureLabelValue); err != nil {
		return fmt.Errorf("%s is an invalid value for the label %s", featureLabelValue, mnnvl.FeatureLabelKey)
	}

	// Deny workloads with MNNVL label when MNNVL is not enabled
	if !mv.clusterMnnvlEnabled {
		return fmt.Errorf("%s is not enabled in the cluster", mnnvl.UserFacingFeatureName)
	}

	// All other validations apply regardless of label value (Required or Preferred)
	if !mv.clusterElasticCDEnabled && apijobs.IsElasticPytorchJob(job) {
		return fmt.Errorf("%s elastic %s", mnnvlUnsupportedErrPrefix, constants.KindPyTorchJob)
	}

	if err := mnnvlutils.ValidateResources(job); err != nil {
		return fmt.Errorf("%s the requested resources, violations: %w", mnnvlUnsupportedErrPrefix, err)
	}

	return nil
}
