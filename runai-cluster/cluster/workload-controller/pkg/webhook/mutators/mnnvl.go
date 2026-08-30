package mutators

import (
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logf.Log.WithName("mnnvl-mutator")

// setMnnvlFeatureLabel set the MNNVL feature label in the workload's labels (ObjectMeta.Labels)
func setMnnvlFeatureLabel(jobWorkload client.Object, clusterSupportMNNVL, elasticCDEnabled bool) error {
	// supported only for Distributed
	dw, isDist := jobWorkload.(*dwv2.DistributedWorkload)
	if !isDist {
		return nil
	}

	metaLabels := dw.GetLabels()
	if metaLabels == nil {
		metaLabels = make(map[string]string)
	}

	// label already exists in meta.labels, do nothing
	_, metaExists := metaLabels[mnnvl.FeatureLabelKey]
	if metaExists {
		return nil
	}

	jw := &jobs.JobWorkload{}
	dw.CopyTo(jw)

	labelFromSpec, err := tryGetMnnvlLabelFromSpec(jw)
	if err != nil {
		return err
	}

	// if the label was provided by the user in the spec use it, else set the default value
	if labelFromSpec != nil {
		metaLabels[mnnvl.FeatureLabelKey] = *labelFromSpec
	} else {
		metaLabels[mnnvl.FeatureLabelKey] = getMnnvlLabelDefaultValue(jw, clusterSupportMNNVL, elasticCDEnabled)
	}

	jobWorkload.SetLabels(metaLabels)
	return nil
}

func getMnnvlLabelDefaultValue(workload *jobs.JobWorkload, clusterSupportMNNVL, elasticCDEnabled bool) string {
	if isValidForMNNVL(workload, clusterSupportMNNVL, elasticCDEnabled) {
		return mnnvl.Preferred
	} else {
		return mnnvl.None
	}
}

func isValidForMNNVL(workload *jobs.JobWorkload, clusterSupportMNNVL, elasticCDEnabled bool) bool {
	if !clusterSupportMNNVL || (!elasticCDEnabled && jobs.IsElasticPytorchJob(workload)) {
		return false
	}
	err := mnnvlutils.ValidateResources(workload)
	if err != nil {
		log.V(4).Info("workload resources are not valid for MNNVL", "violations", err)
		return false
	}
	return true

}

func tryGetMnnvlLabelFromSpec(workload *jobs.JobWorkload) (*string, error) {
	specLabel, specExists := getLabelFromSpec(&workload.Spec, mnnvl.FeatureLabelKey)
	masterSpecLabel, masterSpecExists := getLabelFromSpec(workload.MasterSpec, mnnvl.FeatureLabelKey)

	// user provided different values in Spec and MasterSpec
	if specExists && masterSpecExists && specLabel != masterSpecLabel {
		return nil, fmt.Errorf("providing different values for the label '%s' in Spec and MasterSpec is unsupported", mnnvl.FeatureLabelKey)
	}

	if specExists {
		return ptr.To(specLabel), nil
	}
	if masterSpecExists {
		return ptr.To(masterSpecLabel), nil
	}

	return nil, nil
}

func getLabelFromSpec(spec *jobs.JobSpec, labelKey string) (string, bool) {
	if spec == nil {
		return "", false
	}

	labels := spec.Labels.Deref()
	value, exists := labels[labelKey]
	if !exists {
		return "", false
	}

	return utils.Deref(value.Value), true
}
