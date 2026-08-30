/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package frontend

import (
	"context"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var logg = logger.Log.WithName("webhook.wt")

// +kubebuilder:webhook:path=/mvalidate-run-ai-v1-trainingworkload,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=trainingworkloas,verbs=create;update,versions=metav1,name=runai-workload-workload-validation.kb.io,admissionReviewVersions=metav1

// as part of validation we need client to load relevant resources

func CreateMergeStackForFrontend(trainingWorkload *v2alpha1.TrainingWorkload, k8sClient client.Client, options WebhookOptions) (*enforcer.MergeStack, error) {
	ctx := logger.IntoContext(context.Background(), logg)
	useControlPlanePolicies := policyUtils.SetupControlPlanePoliciesUsage(ctx, k8sClient)

	//
	//   this option is for frontend/testing who wishes to only merge the stack and not do real work
	//
	if options.SuppressBuilding {
		trainingWorkload.Spec.Usage = common.UsedAsTemplate
	}

	if options.TargetNamespace != "" {
		trainingWorkload.ObjectMeta.Namespace = options.TargetNamespace
	}

	var jobWorkload jobs.JobWorkload
	trainingWorkload.CopyTo(&jobWorkload)
	trainingMergeStackCreator := factory.NewTrainingWorkloadMergeCreator(k8sClient, useControlPlanePolicies)
	return CreateUnitedMergeStackFrontend(ctx, k8sClient, options, &jobWorkload, trainingMergeStackCreator, trainingWorkload)
}

// CreateMergeStack load the workload and related policy/base workload to a merge stack.
// we can use the stack for merging and enforcing the complete set of values against the complete set of rules

// MergeForFrontend is used by RS, in order for composing a merge stack for the frontend
func MergeForFrontendTrainingWorkload(k8sClient client.Client, trainingWorkload *v2alpha1.TrainingWorkload, targetNamespace string) ([]byte, error) {

	mergeStack, err := CreateMergeStackForFrontend(trainingWorkload, k8sClient, WebhookOptions{
		SuppressBuilding: true,
		TargetNamespace:  targetNamespace,
	})
	if err != nil {
		return nil, err
	}
	return mergeStack.MergeWorkloadForFrontend()
}
