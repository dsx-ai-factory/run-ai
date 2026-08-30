/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package frontend

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("frontend.wt")

func InteractiveCreateMergeStackForFrontend(k8sClient client.Client, interactive *v2alpha1.InteractiveWorkload, options WebhookOptions) (*enforcer.MergeStack, error) {

	ctx := logger.IntoContext(context.Background(), log)
	useControlPlanePolicies := policyUtils.SetupControlPlanePoliciesUsage(ctx, k8sClient)

	//
	//   this option is for frontend/testing who wishes to only merge the stack and not do real work
	//
	if options.SuppressBuilding {
		interactive.Spec.Usage = common.UsedAsTemplate
	}

	if options.TargetNamespace != "" {
		interactive.ObjectMeta.Namespace = options.TargetNamespace
	}

	var jobWorkload jobs.JobWorkload
	interactive.CopyTo(&jobWorkload)
	interactiveMergeStackCreator := factory.NewInteractiveWorkloadMergeCreator(k8sClient, useControlPlanePolicies)

	return CreateUnitedMergeStackFrontend(ctx, k8sClient, options, &jobWorkload, interactiveMergeStackCreator, interactive)
}

// CreateMergeStack load the workload and related policy/base workload to a merge stack.
// we can use the stack for merging and enforcing the complete set of values against the complete set of rules

// MergeForFrontend is used by RS, in order for composing a merge stack for the frontend
func MergeForFrontendInteractiveWorkload(k8sClient client.Client, interactive *v2alpha1.InteractiveWorkload, targetNamespace string) ([]byte, error) {

	mergeStack, err := InteractiveCreateMergeStackForFrontend(k8sClient, interactive, WebhookOptions{
		SuppressBuilding: true,
		TargetNamespace:  targetNamespace,
	})
	if err != nil {
		return nil, err
	}
	return mergeStack.MergeWorkloadForFrontend()
}

// InitTest is a function to allow testing with fake client
