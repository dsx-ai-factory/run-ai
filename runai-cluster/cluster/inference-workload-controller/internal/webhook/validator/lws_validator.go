package validator

import (
	"context"
	"errors"

	"github.com/run-ai/runai/runai-cluster/cluster/common/runaipublic"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LwsValidator struct {
	k8sClient client.Client
}

func NewLwsValidator(k8s client.Client) *LwsValidator {
	return &LwsValidator{k8sClient: k8s}
}

func (v *LwsValidator) Validate(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	runaiPublic := runaipublic.NewRunaiPublicHandler(ctx, v.k8sClient)
	isAvailable, reason, err := runaiPublic.IsLwsAvailable()
	if err != nil {
		return err
	}
	if !isAvailable {
		msg := "LWS is not available"
		if len(reason) > 0 {
			msg = msg + ", reason: " + reason
		}
		return errors.New(msg)
	}

	return nil
}
