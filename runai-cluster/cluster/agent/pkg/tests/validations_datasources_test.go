/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BeCorrectlySyncedPvc(expected PvcAndOptionalUpdate) MatcherCorrectlySyncedPvc {
	return MatcherCorrectlySyncedPvc{
		Expected: expected,
	}
}

type MatcherCorrectlySyncedPvc struct {
	Expected PvcAndOptionalUpdate
}

func (matcher MatcherCorrectlySyncedPvc) Match(actual interface{}) (success bool, err error) {
	fullName := fmt.Sprintf("pvc %s:%s", matcher.Expected.Pvc.Namespace, matcher.Expected.Pvc.Name)

	if openapi.IsNil(actual) {
		return false, fmt.Errorf("%s not created in the cluster", fullName)
	}

	actualPvc := actual.(*corev1.PersistentVolumeClaim)

	if matcher.Expected.Update != nil {
		if actualPvc.Labels == nil {
			return false, fmt.Errorf("%s: missing labels", fullName)
		}

		if actualPvc.Annotations == nil {
			return false, fmt.Errorf("%s: missing annotations", fullName)
		}

		if actualPvc.Labels["run.ai/asset"] != "pvc" {
			return false, fmt.Errorf("%s: run.ai/asset is %s, should be pvc", fullName, actualPvc.Labels["run.ai/asset"])
		}

		if actualPvc.Labels["run.ai/asset-id"] != matcher.Expected.Update.Meta.Id {
			return false, fmt.Errorf("%s: run.ai/asset-id %v does not point to asset %s", fullName, actualPvc.Labels["run.ai/asset-id"], matcher.Expected.Update.Meta.Id)
		}

		updateCountAsString := fmt.Sprintf("%d", matcher.Expected.Update.UpdateCount)

		if actualPvc.Annotations["run.ai/update-count"] != updateCountAsString {
			return false, fmt.Errorf("%s: run.ai/update-count %v does not hold update count %s", fullName, actualPvc.Annotations["run.ai/update-count"], updateCountAsString)
		}
	}

	actualStorageClass := convutils.Deref(actualPvc.Spec.StorageClassName, "")
	expectedStorageClass := convutils.Deref(matcher.Expected.Pvc.Spec.StorageClassName, "")

	if actualStorageClass != expectedStorageClass {
		return false, fmt.Errorf("%s: storage class name %v, expected %v", fullName, actualStorageClass, expectedStorageClass)
	}

	if actualPvc.Spec.Resources.String() != matcher.Expected.Pvc.Spec.Resources.String() {
		return false, fmt.Errorf("%s: resources %v, expected %v", fullName, actualPvc.Spec.Resources.String(), matcher.Expected.Pvc.Spec.Resources.String())
	}

	return true, nil
}

func (matcher MatcherCorrectlySyncedPvc) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "pvc %s:%s is not as expected", matcher.Expected.Pvc.Namespace, matcher.Expected.Pvc.Name)
}

func (matcher MatcherCorrectlySyncedPvc) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "pvc %s:%s match the expected, failing the not(...) condition", matcher.Expected.Pvc.Namespace, matcher.Expected.Pvc.Name)
}

func validatePvcExistsEventually(pvcAndOptionalUpdate PvcAndOptionalUpdate, k8sClient client.Client) {
	var pvcResource *corev1.PersistentVolumeClaim = nil

	if pvcAndOptionalUpdate.Update == nil {
		EventuallyWithOffset(1, func() *corev1.PersistentVolumeClaim {
			pvcResource = getPvcFromClient(pvcAndOptionalUpdate.Pvc.Name, pvcAndOptionalUpdate.Pvc.Namespace, k8sClient)
			return pvcResource
		}, DefaultTimeout).ShouldNot(BeNil())
	} else {
		EventuallyWithOffset(1, func() *corev1.PersistentVolumeClaim {
			pvcResource = getPvcFromClient(pvcAndOptionalUpdate.Pvc.Name, pvcAndOptionalUpdate.Pvc.Namespace, k8sClient)
			return pvcResource
		}, DefaultTimeout).Should(BeCorrectlySyncedPvc(pvcAndOptionalUpdate))
	}
}

func validatePvcDoesNotExistsEventually(pvc *corev1.PersistentVolumeClaim, k8sClient client.Client) {
	var pvcResource *corev1.PersistentVolumeClaim = nil

	EventuallyWithOffset(1, func() *corev1.PersistentVolumeClaim {
		pvcResource = getPvcFromClient(pvc.Name, pvc.Namespace, k8sClient)
		return pvcResource
	}, DefaultTimeout).Should(BeNil())
}

func validatePvcLabelsEventually(pvcAndOptionalUpdate PvcAndOptionalUpdate, k8sClient client.Client) {
	expectedLabelValue := "true"
	if pvcAndOptionalUpdate.Update == nil {
		expectedLabelValue = ""
	} else if pvcAndOptionalUpdate.Update.Spec.Pvc.Get().DataSharing.Get() == nil {
		expectedLabelValue = ""
	}

	EventuallyWithOffset(1, func() string {
		pvc := getPvcFromClient(pvcAndOptionalUpdate.Pvc.Name, pvcAndOptionalUpdate.Pvc.Namespace, k8sClient)
		if pvc != nil {
			return pvc.Labels[clusterCommonConstants.RunaiDataSharingLabel]
		}
		return ""
	}, DefaultTimeout).Should(BeEquivalentTo(expectedLabelValue))
}

func validatePvcAnnotationsEventually(pvcAndOptionalUpdate PvcAndOptionalUpdate, k8sClient client.Client) {
	expectedAnnotations := make(map[string]string)

	if pvcAndOptionalUpdate.Update != nil {
		claimInfo := pvcAndOptionalUpdate.Update.Spec.Pvc.Get().ClaimInfo.Get()
		if claimInfo != nil && claimInfo.AddedAttrValues != nil {
			for _, attr := range claimInfo.AddedAttrValues {
				expectedAnnotations[fmt.Sprintf("attributes.run.ai/%s", attr.Key)] = convutils.Deref(attr.Value, "")
			}
		}
	}

	EventuallyWithOffset(1, func() map[string]string {
		pvc := getPvcFromClient(pvcAndOptionalUpdate.Pvc.Name, pvcAndOptionalUpdate.Pvc.Namespace, k8sClient)
		if pvc != nil && pvc.Annotations != nil {
			attrRelatedAnnotations := make(map[string]string)
			for k, v := range pvc.Annotations {
				if strings.HasPrefix(k, "attributes.run.ai") {
					attrRelatedAnnotations[k] = v
				}
			}
			return attrRelatedAnnotations
		}
		return make(map[string]string)
	}, DefaultTimeout).Should(BeEquivalentTo(expectedAnnotations))
}

func validateAgentConfigWithDatasourceCounter(k8sClient client.Client, counterValue string) {
	validateAgentConfigWithCounter(k8sClient, configmaps.DatasourcesSyncCounter, counterValue)
}

func validateAgentConfigWithCounter(k8sClient client.Client, counterKey string, counterValue string) {
	EventuallyWithOffset(1, func() string {
		configMap := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name:      configmaps.AgentConfigName,
			Namespace: clusterCommonConstants.NamespaceRunAi,
		}, configMap)
		if err != nil {
			return ""
		}

		counter, ok := configMap.Data[counterKey]
		if !ok {
			return ""
		}

		return counter
	}, DefaultTimeout).Should(Equal(counterValue))
}
