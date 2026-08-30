package workloads

import (
	"context"
	"fmt"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// this only workaround until we will fix the issue with interactive and training workloads

type SingleRunaiJobValidator struct {
	k8sClient client.Client
}

func NewSingleRunaiJobValidator(k8sClient client.Client) *SingleRunaiJobValidator {
	return &SingleRunaiJobValidator{
		k8sClient: k8sClient,
	}
}
func (serve SingleRunaiJobValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	log := logf.FromContext(ctx)
	log.V(4).Info("checking if runai job exists")

	workloadName := workload.GetChildName()

	runaiJob := runaiJobV1.RunaiJob{}
	err := serve.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: workload.Namespace,
		Name:      workloadName,
	}, &runaiJob)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(4).Info("runai job does not exist")
			return nil
		}
		log.Error(err, "failed to get runai job")
		return fmt.Errorf("got error while retriving runai job, error: %v", err)
	}
	log.V(4).Info("runai job exists")
	return fmt.Errorf("runai job with name %s already exists", workloadName)
}
