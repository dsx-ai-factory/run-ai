/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"fmt"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	v1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	dinfpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"

	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type policyLevel string

const (
	cluster    policyLevel = "cluster"
	department policyLevel = "department"
	project    policyLevel = "project"
)

type policyType string

const (
	KindTrainingPolicy             policyType = clusterCommonConstants.KindTrainingPolicy
	KindTrainingPolicyList         policyType = clusterCommonConstants.KindTrainingPolicyList
	KindInteractivePolicy          policyType = clusterCommonConstants.KindInteractivePolicy
	KindInteractivePolicyList      policyType = clusterCommonConstants.KindInteractivePolicyList
	KindInferencePolicy            policyType = clusterCommonConstants.KindInferencePolicy
	KindDistributedPolicy          policyType = clusterCommonConstants.KindDistributedPolicy
	KindDistributedInferencePolicy policyType = clusterCommonConstants.KindDistributedInferencePolicy
)

type FindPolicyByLevelSearchParams struct {
	policyLevel  policyLevel // the level of the policy to find (project, department, tenant)
	policyType   policyType  // the type of the policy to find (KindTrainingPolicy, KindInteractivePolicy, KindInferencePolicy, KindDistributedPolicy, KindDistributedInferencePolicy)
	namespace    string      // the namespace of the workload (runai for department, tenant level or the project namespace)
	projectId    string      // the project id of the workload (for project level)
	departmentId string      // the department id of the workload (for department level)
}

func GetNumberOfPolicies(listResult client.ObjectList) (int, error) {
	switch itemList := listResult.(type) {
	case *ipv1.InteractivePolicyList:
		interactivePolicyList := itemList
		return len(interactivePolicyList.Items), nil
	case *tpv1.TrainingPolicyList:
		trainingPolicyList := itemList
		return len(trainingPolicyList.Items), nil
	case *infpv1.InferencePolicyList:
		inferencePolicyList := itemList
		return len(inferencePolicyList.Items), nil
	case *dpv1.DistributedPolicyList:
		distributedPolicyList := itemList
		return len(distributedPolicyList.Items), nil
	case *dinfpv1.DistributedInferencePolicyList:
		distributedInferencePolicyList := itemList
		return len(distributedInferencePolicyList.Items), nil
	default:
		return 0, fmt.Errorf("unsupported policy type %v", listResult)
	}

}

func validateNumberOfPolicies(listResult client.ObjectList, expectedNumberOfPolices int) error {
	numerOfPolicies, err := GetNumberOfPolicies(listResult)
	if err != nil {
		return fmt.Errorf("failed to get number of policies, error: %v", err)
	}

	if numerOfPolicies != expectedNumberOfPolices {
		if numerOfPolicies == 0 {
			return errors.NewNotFound(v1.Resource("policy"), "")
		}
		return fmt.Errorf("expected to find %v policies, but found %v policies", expectedNumberOfPolices, numerOfPolicies)
	}

	return nil
}

func getPolicyFromClient(ctx context.Context, k8sClient client.Client, policyToFind FindPolicyByLevelSearchParams) (client.Object, error) {
	var listResult client.ObjectList

	switch policyToFind.policyType {
	case KindInteractivePolicy:
		listResult = &ipv1.InteractivePolicyList{}
	case KindTrainingPolicy:
		listResult = &tpv1.TrainingPolicyList{}
	case KindInferencePolicy:
		listResult = &infv1.InferencePolicyList{}
	case KindDistributedPolicy:
		listResult = &dpv1.DistributedPolicyList{}
	case KindDistributedInferencePolicy:
		listResult = &dinfpv1.DistributedInferencePolicyList{}
	default:
		return nil, fmt.Errorf("unsupported policy type %v", policyToFind.policyType)
	}

	labels := map[string]string{
		clusterCommonConstants.PolicyLabel: clusterCommonConstants.BoolTrueAsString,
		// this label is used to identify new policies in the cluster
		// that were created by the agent from the backend
	}
	inNamespace := client.InNamespace("")

	switch policyToFind.policyLevel {
	case cluster:
		inNamespace = client.InNamespace(clusterCommonConstants.NamespaceRunAi)
		labels[clusterCommonConstants.ClusterWideResourceLabel] = clusterCommonConstants.BoolTrueAsString
	case department:
		inNamespace = client.InNamespace(clusterCommonConstants.NamespaceRunAi)
		labels[clusterCommonConstants.DepartmentScopeLabel] = policyToFind.departmentId
	case project:
		inNamespace = client.InNamespace(policyToFind.namespace)
		labels[clusterCommonConstants.ProjectScopeLabel] = policyToFind.projectId
	}

	err := k8sClient.List(ctx, listResult,
		inNamespace,
		client.MatchingLabels(labels))
	if err != nil {
		return nil, err
	}

	err = validateNumberOfPolicies(listResult, 1)
	if err != nil {
		return nil, err
	}

	switch policyToFind.policyType {
	case KindInteractivePolicy:
		return &listResult.(*ipv1.InteractivePolicyList).Items[0], nil
	case KindTrainingPolicy:
		return &listResult.(*tpv1.TrainingPolicyList).Items[0], nil
	case KindInferencePolicy:
		return &listResult.(*infpv1.InferencePolicyList).Items[0], nil
	case KindDistributedPolicy:
		return &listResult.(*dpv1.DistributedPolicyList).Items[0], nil
	case KindDistributedInferencePolicy:
		return &listResult.(*dinfpv1.DistributedInferencePolicyList).Items[0], nil
	}

	return nil, err
}

func findPolicyByLevel(ctx context.Context, k8sClient client.Client, policyToFind FindPolicyByLevelSearchParams) (client.Object, policyLevel, error) {
	switch policyToFind.policyLevel {
	case cluster:
		policy, err := getPolicyFromClient(ctx, k8sClient, policyToFind)
		if err != nil {
			return policy, cluster, nil
		}
		return policy, cluster, err
	case department:
		policy, err := getPolicyFromClient(ctx, k8sClient, policyToFind)
		if err != nil {
			if errors.IsNotFound(err) {
				policyToFind.policyLevel = cluster
				return findPolicyByLevel(ctx, k8sClient, policyToFind)
			}
		}
		return policy, department, err
	case project:
		policy, err := getPolicyFromClient(ctx, k8sClient, policyToFind)
		if err != nil {
			if errors.IsNotFound(err) {
				policyToFind.policyLevel = department
				policyToFind.namespace = clusterCommonConstants.NamespaceRunAi
				return findPolicyByLevel(ctx, k8sClient, policyToFind)
			}
		}
		return policy, project, err
	default:
		return client.Object(nil), "", fmt.Errorf("unknown policy level %v ", policyToFind.policyLevel)
	}
}

func IsBackendSyncedPolicy(itemLabels map[string]string) bool {
	return itemLabels == nil || itemLabels[clusterCommonConstants.PolicyLabel] != clusterCommonConstants.BoolTrueAsString
}

func FindBackendSyncedPolicy(ctx context.Context, k8sClient client.Client, workloadNamespace string, workloadLabels *map[string]string, policyType policyType) (client.Object, policyLevel, error) {
	var projectName string
	var err error

	if workloadLabels != nil && (*workloadLabels)[clusterCommonConstants.RunaiProjectLabel] != "" {
		projectName = (*workloadLabels)[clusterCommonConstants.RunaiProjectLabel]
	} else {
		projectName, err = workloadproject.ProjectNameOfNamespace(ctx, k8sClient, workloadNamespace)
		if err != nil {
			return nil, "", err
		}
	}

	// get the kai.resources project object to read the department id and project id labels.
	projectObject := &kaiorg.Project{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: projectName}, projectObject); err != nil {
		return nil, "", err
	}
	projectLabels := projectObject.Labels

	var departmentId, projectId string
	if projectLabels != nil {
		departmentId = projectLabels[schedv2alpha2.DepartmentIdLabel]
		projectId = projectLabels[v2.ProjectIdLabel]
	}

	// try to find a policy
	return findPolicyByLevel(ctx, k8sClient, FindPolicyByLevelSearchParams{
		policyLevel:  project,
		policyType:   policyType,
		namespace:    workloadNamespace,
		projectId:    projectId,
		departmentId: departmentId,
	})
}
