package converter

// for inference workload convertor
import (
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InferenceWorkloadConvertor struct {
}

func (self *InferenceWorkloadConvertor) Convert(jobWorkload *jobs.JobWorkload) client.Object {
	return infwv2.NewFromJobWorkload(jobWorkload)
}

func NewInferenceWorkloadConvertor() *InferenceWorkloadConvertor {
	return &InferenceWorkloadConvertor{}
}
