package test

import (
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/datavolumes_openapi"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateDatavolumeExistsEventually(nodePoolName string, k8sClient client.Client) {
	var dep *v1alpha1.DataVolume = nil
	EventuallyWithOffset(1, func() *v1alpha1.DataVolume {
		dep = getDatavolumeFromClient(nodePoolName, k8sClient)
		return dep
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateDatavolumeDeletedEventually(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		np := getDatavolumeFromClient(nodePoolName, k8sClient)
		g.Expect(np).To(BeNil())
	}, DefaultTimeout).Should(Succeed())
}

func validateExistingDatavolumeMatchesBackendSpec(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingDatavolumeWithBackendSpec(g, nodePoolName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingDatavolumeDoesNotMatchBackendSpec(datavolumeName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingDatavolumeWithBackendSpec(g, datavolumeName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingDatavolumeWithBackendSpec(g Gomega, datavolumeName string, k8sClient client.Client) {
	clusterDatavolume := getDatavolumeFromClient(datavolumeName, k8sClient)
	g.ExpectWithOffset(1, clusterDatavolume).ToNot(BeNil(), "should exist")

	backendDvSpec := BackendDatavolumesMap[datavolumeName]
	g.ExpectWithOffset(1, backendDvSpec).ToNot(BeNil(), "Backend DV spec should exist")

	g.ExpectWithOffset(1, clusterDatavolume.Name).To(Equal(backendDvSpec.Name), "name")
	if backendDvSpec.Description != nil {
		g.ExpectWithOffset(1, clusterDatavolume.Spec.Description).To(Equal(*backendDvSpec.Description), "description")
	} else {
		g.ExpectWithOffset(1, clusterDatavolume.Spec.Description).To(BeEmpty(), "description")
	}
	g.ExpectWithOffset(1, clusterDatavolume.Spec.OriginPvcName).To(Equal(backendDvSpec.OriginPvcName), "origin PVC name")
	g.ExpectWithOffset(1, clusterDatavolume.Spec.ProjectName).To(Equal(backendDvSpec.ProjectName), "project name")
	g.ExpectWithOffset(1, clusterDatavolume.Spec.DatavolumeId).To(Equal(backendDvSpec.Id), "datavolume ID")

	if backendDvSpec.ShouldDeleteOriginalVolume != nil {
		g.ExpectWithOffset(1, clusterDatavolume.Spec.ShouldDeleteOriginalVolume).To(Equal(*backendDvSpec.ShouldDeleteOriginalVolume), "should delete original volume")
	} else {
		g.ExpectWithOffset(1, clusterDatavolume.Spec.ShouldDeleteOriginalVolume).To(BeFalse(), "should delete original volume")
	}

	validateSharedScopes(g, clusterDatavolume.Spec.SharedScopes, backendDvSpec.SharedScopes)
}

func validateSharedScopes(g Gomega, clusterSharedScopes []v1alpha1.DataVolumeSharedScope, backendSharedScopes []datavolumes_openapi.SharedScope) {
	g.ExpectWithOffset(2, len(clusterSharedScopes)).To(Equal(len(backendSharedScopes)), "shared scopes count")
	for i, clusterSharedScope := range clusterSharedScopes {
		g.ExpectWithOffset(2, clusterSharedScope.ScopeName).To(Equal(backendSharedScopes[i].ScopeName), "scope ID")
		g.ExpectWithOffset(2, string(clusterSharedScope.ScopeType)).To(Equal(string(backendSharedScopes[i].ScopeType)), "scope type")
	}
}
