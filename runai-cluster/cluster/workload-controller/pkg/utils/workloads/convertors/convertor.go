package convertors

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InteractiveWorkloadConvertor struct {
}

func (self *InteractiveWorkloadConvertor) Convert(jobWorkload *jobs.JobWorkload) client.Object {
	return iwv2.NewFromJobWorkload(jobWorkload)
}

// for training workload convertor
type TrainingWorkloadConvertor struct {
}

func (self *TrainingWorkloadConvertor) Convert(jobWorkload *jobs.JobWorkload) client.Object {
	return twv2.NewFromJobWorkload(jobWorkload)
}

// for distributed workload convertor
type DistributedWorkloadConvertor struct {
}

func (self *DistributedWorkloadConvertor) Convert(jobWorkload *jobs.JobWorkload) client.Object {
	return dwv2.NewFromJobWorkload(jobWorkload)
}

func NewInteractiveWorkloadConvertor() *InteractiveWorkloadConvertor {
	return &InteractiveWorkloadConvertor{}
}

func NewTrainingWorkloadConvertor() *TrainingWorkloadConvertor {
	return &TrainingWorkloadConvertor{}
}

func NewDistributedWorkloadConvertor() *DistributedWorkloadConvertor {
	return &DistributedWorkloadConvertor{}
}

// GeneralWorkloadConverter type-agnostic converter from job to workload
type GeneralWorkloadConverter struct {
}

func NewGeneralWorkloadConverter() *GeneralWorkloadConverter {
	return &GeneralWorkloadConverter{}
}

func (self *GeneralWorkloadConverter) Convert(jobWorkload *jobs.JobWorkload) client.Object {
	if jobWorkload == nil {
		return nil
	}

	switch jobWorkload.Kind {
	case constants.KindTrainingWorkload:
		return twv2.NewFromJobWorkload(jobWorkload)
	case constants.KindInteractiveWorkload:
		return iwv2.NewFromJobWorkload(jobWorkload)
	case constants.KindDistributedWorkload:
		return dwv2.NewFromJobWorkload(jobWorkload)
	default:
		return nil
	}
}
