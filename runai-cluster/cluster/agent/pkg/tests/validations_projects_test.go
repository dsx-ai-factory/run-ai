package test

import (
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/resourcing"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/projects_openapi"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/handlers"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateProjectExistsEventually(projectName string, k8sClient client.Client) {
	var proj *v2.Project = nil
	EventuallyWithOffset(1, func() *v2.Project {
		proj = getLegacyProjectFromClient(projectName, k8sClient)
		return proj
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateProjectDeletedEventually(projectName string, k8sClient client.Client) {
	var proj *v2.Project = nil
	EventuallyWithOffset(1, func() *v2.Project {
		proj = getLegacyProjectFromClient(projectName, k8sClient)
		return proj
	}, DefaultTimeout).Should(BeNil())
}

func validateExistingProjectMatchesBackendSpec(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingProjectWithBackend(g, projectName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingProjectDoesNotMatchBackendSpec(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingProjectWithBackend(g, projectName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingProjectWithBackend(g Gomega, projectName string, k8sClient client.Client) {
	clusterProject := getLegacyProjectFromClient(projectName, k8sClient)
	g.ExpectWithOffset(1, clusterProject).ToNot(BeNil())

	backendProjSpec := OrgUnitProjectsMap[projectName]
	g.ExpectWithOffset(1, backendProjSpec).ToNot(BeNil(), "Backend project spec should exist")

	g.ExpectWithOffset(1, clusterProject.Name).To(Equal(backendProjSpec.Name), "project name")
	g.ExpectWithOffset(1, clusterProject.Spec.Department).To(Equal(backendProjSpec.Parent.Get().Name), "department name")
	g.ExpectWithOffset(1, clusterProject.Labels[v2.ProjectIdLabel]).To(Equal(backendProjSpec.Id), "project id")
	g.ExpectWithOffset(1, clusterProject.Labels[schedulingv2alpha2.DepartmentIdLabel]).To(Equal(backendProjSpec.Parent.Get().Id), "department id")
	g.ExpectWithOffset(1, clusterProject.Spec.DefaultNodePools).To(Equal(backendProjSpec.DefaultNodePools), "DefaultNodePools")
	schedulingRules := backendProjSpec.SchedulingRules.Get()
	if schedulingRules == nil {
		g.ExpectWithOffset(1, clusterProject.Spec.InteractiveJobMaxIdleDurationSecs).To(Equal(int64(0)), "InteractiveJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(int64(0)), "InteractivePreemptibleJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.TrainingJobMaxIdleDurationSecs).To(Equal(int64(0)), "TrainingJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.InteractiveJobTimeLimitSecs).To(Equal(int64(0)), "InteractiveJobTimeLimitSecs")
	} else {
		g.ExpectWithOffset(1, clusterProject.Spec.InteractiveJobMaxIdleDurationSecs).To(Equal(int64(*schedulingRules.InteractiveJobMaxIdleDurationSeconds.Get())), "InteractiveJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(int64(*schedulingRules.InteractiveJobPreemptIdleDurationSeconds.Get())), "InteractivePreemptibleJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.TrainingJobMaxIdleDurationSecs).To(Equal(int64(*schedulingRules.TrainingJobMaxIdleDurationSeconds.Get())), "TrainingJobMaxIdleDurationSecs")
		g.ExpectWithOffset(1, clusterProject.Spec.InteractiveJobTimeLimitSecs).To(Equal(int64(*schedulingRules.InteractiveJobTimeLimitSeconds.Get())), "InteractiveJobTimeLimitSecs")
	}
	g.ExpectWithOffset(1, clusterProject.Spec.Namespace).To(Equal(common.GetStringVal(backendProjSpec.RequestedNamespace.Get())), "Namespace")

	g.ExpectWithOffset(1, len(clusterProject.Spec.Queues)).To(Equal(len(backendProjSpec.Resources)), "Queues length")
	for _, queue := range clusterProject.Spec.Queues {
		validateQueue(g, queue.Name, queue.Nodepool, queue.Priority, ptr.Deref(queue.Resources, resourcing.QueueResources{}), backendProjSpec.Resources, backendProjSpec.Name)
	}

	validateOrgUnitNodeAffinity(g, clusterProject.Spec, backendProjSpec.NodeTypes)
}

func validateOrgUnitResourcesMatching(g Gomega, resources resourcing.QueueResources, backandResources *openapi.ResourcesSync) {
	validateOrgUnitGpuTypeResourcesMatching(g, resources.GPU, backandResources.Gpu)
	validateOrgUnitCpuTypeResourcesMatching(g, resources.CPU, backandResources.Cpu)
	validateOrgUnitMemoryTypeResourcesMatching(g, resources.Memory, backandResources.Memory)
}

func validateOrgUnitGpuTypeResourcesMatching(g Gomega, clusterGpuTypeResources resourcing.SystemResource, backendGpuTypeResources openapi.ResourcesSyncGpu) {
	g.ExpectWithOffset(3, clusterGpuTypeResources.Deserved).To(Equal(backendGpuTypeResources.Deserved), "clusterGpuTypeResources.Deserved")

	if !backendGpuTypeResources.OverQuotaWeight.IsSet() || backendGpuTypeResources.OverQuotaWeight.Get() == nil {
		g.ExpectWithOffset(3, clusterGpuTypeResources.OverQuotaWeight).To(Equal(float64(1)), "clusterGpuTypeResources.OverQuotaWeight")
	} else {
		g.ExpectWithOffset(3, clusterGpuTypeResources.OverQuotaWeight).To(Equal(*backendGpuTypeResources.OverQuotaWeight.Get()), "clusterGpuTypeResources.OverQuotaWeight")
	}

	if !backendGpuTypeResources.Limit.IsSet() || backendGpuTypeResources.Limit.Get() == nil {
		g.ExpectWithOffset(3, clusterGpuTypeResources.Limit).To(Equal(float64(handlers.NoMaxAllowed)), "clusterGpuTypeResources.Limit")
	} else {
		g.ExpectWithOffset(3, clusterGpuTypeResources.Limit).To(Equal(*backendGpuTypeResources.Limit.Get()), "clusterGpuTypeResources.Limit")
	}
}

func validateOrgUnitCpuTypeResourcesMatching(g Gomega, clusterCpuTypeResources resourcing.SystemResource, backendCpuTypeResources openapi.NullableResourcesCpu) {
	if !backendCpuTypeResources.IsSet() || backendCpuTypeResources.Get() == nil {
		g.ExpectWithOffset(3, clusterCpuTypeResources).To(Equal(resourcing.SystemResource{
			Deserved:        handlers.NoMaxAllowed,
			OverQuotaWeight: handlers.DefaultOverQuotaWeight,
			Limit:           handlers.NoMaxAllowed,
		}), "clusterCpuTypeResources")
		return
	}

	validateCpuMemory(g, backendCpuTypeResources.Get().Deserved, backendCpuTypeResources.Get().Limit, backendCpuTypeResources.Get().OverQuotaWeight, clusterCpuTypeResources)
}

func validateOrgUnitMemoryTypeResourcesMatching(g Gomega, clusterCpuTypeResources resourcing.SystemResource, backendMemoryTypeResources openapi.NullableResourcesSyncMemory) {
	if !backendMemoryTypeResources.IsSet() || backendMemoryTypeResources.Get() == nil {
		g.ExpectWithOffset(3, clusterCpuTypeResources).To(Equal(resourcing.SystemResource{
			Deserved:        handlers.NoMaxAllowed,
			OverQuotaWeight: handlers.DefaultOverQuotaWeight,
			Limit:           handlers.NoMaxAllowed,
		}), "clusterCpuTypeResources")
		return
	}

	validateCpuMemory(g, backendMemoryTypeResources.Get().Deserved, backendMemoryTypeResources.Get().Limit, backendMemoryTypeResources.Get().OverQuotaWeight, clusterCpuTypeResources)
}

func validateCpuMemory(g Gomega, deserved openapi.NullableFloat64, limit openapi.NullableFloat64, OverQuotaWeight openapi.NullableFloat64, clusterCpuTypeResources resourcing.SystemResource) {
	if !deserved.IsSet() || deserved.Get() == nil {
		g.ExpectWithOffset(3, clusterCpuTypeResources.Deserved).To(Equal(float64(handlers.NoMaxAllowed)), "clusterCpuTypeResources.Deserved")
	} else {
		g.ExpectWithOffset(3, clusterCpuTypeResources.Deserved).To(Equal(float64(*deserved.Get())), "clusterCpuTypeResources.Deserved")
	}

	if !OverQuotaWeight.IsSet() || OverQuotaWeight.Get() == nil {
		g.ExpectWithOffset(3, clusterCpuTypeResources.OverQuotaWeight).To(Equal(float64(handlers.DefaultOverQuotaWeight)), "clusterCpuTypeResources.OverQuotaWeight")
	} else {
		g.ExpectWithOffset(3, clusterCpuTypeResources.OverQuotaWeight).To(Equal(float64(*OverQuotaWeight.Get())), "clusterCpuTypeResources.OverQuotaWeight")
	}

	if limit.IsSet() && limit.Get() != nil {
		g.ExpectWithOffset(3, clusterCpuTypeResources.Limit).To(Equal(float64(*limit.Get())), "clusterCpuTypeResources.Limit")
	} else {
		g.ExpectWithOffset(3, clusterCpuTypeResources.Limit).To(Equal(float64(handlers.NoMaxAllowed)), "clusterCpuTypeResources.Limit")
	}
}

func validateOrgUnitNodeAffinity(g Gomega, clusterProjectSpec v2.ProjectSpec, backendJobsNodeAffinity *openapi.NodeTypesPerWorkload) {
	if backendJobsNodeAffinity == nil {
		g.ExpectWithOffset(2, clusterProjectSpec.NodeAffinityTrain).To(BeEmpty(), "clusterProjectSpec.NodeAffinityTrain")
		g.ExpectWithOffset(2, clusterProjectSpec.NodeAffinityInteractive).To(BeEmpty(), "clusterProjectSpec.NodeAffinityTrain")
		return
	}

	validateOrgUnitJobNodeAffinity(g, clusterProjectSpec.NodeAffinityTrain, backendJobsNodeAffinity.Training, backendJobsNodeAffinity.Names)
	validateOrgUnitJobNodeAffinity(g, clusterProjectSpec.NodeAffinityInteractive, backendJobsNodeAffinity.Workspace, backendJobsNodeAffinity.Names)
}

func validateOrgUnitJobNodeAffinity(g Gomega, clusterNodeAffinity []string, nodeTypes []string, names map[string]string) {
	nodeTypeNames := []string{}
	for _, nodeTypeForProject := range nodeTypes {
		nodeTypeNames = append(nodeTypeNames, names[nodeTypeForProject])
	}
	g.ExpectWithOffset(3, clusterNodeAffinity).To(ConsistOf(nodeTypeNames), "clusterNodeAffinity")
}

func _getBackendAssignedResourcesForNodePool(nodePoolName string, nodePoolsResources []projects_openapi.NodePoolAssignedResources) *projects_openapi.NodePoolAssignedResources { //nolint:unused
	for _, npr := range nodePoolsResources {
		if npr.NodePool != nil && npr.NodePool.Name != nil && *npr.NodePool.Name == nodePoolName {
			return &npr
		}
	}
	return nil
}

func getOrgUnitAssignedResourcesForNodePool(nodePoolName string, nodePoolsResources []openapi.ResourcesSync) *openapi.ResourcesSync {
	for _, npr := range nodePoolsResources {
		if npr.NodePool.IsSet() && npr.NodePool.Get() != nil && npr.NodePool.Get().Name != nil && *npr.NodePool.Get().Name == nodePoolName {
			return &npr
		}
	}
	return nil
}

func validateQueue(g Gomega, queueName string, queueNodePool string, priority *int32, queueResources resourcing.QueueResources, resources []openapi.ResourcesSync, name string) {
	backendAssignedResources := getOrgUnitAssignedResourcesForNodePool(queueNodePool, resources)
	g.ExpectWithOffset(1, backendAssignedResources).ToNot(BeNil(), "backendAssignedResources")

	g.ExpectWithOffset(1, queueName).To(Equal(handlers.GetQueueName(name, queueNodePool)), "queue name")
	g.ExpectWithOffset(1, queueNodePool).To(Equal(common.GetStringVal(backendAssignedResources.NodePool.Get().Name)), "queue's nodepool")
	if priority == nil || backendAssignedResources.Rank == nil {
		g.Expect(priority).To(BeNil(), "queue's priority")
		g.Expect(backendAssignedResources.Rank).To(BeNil(), "queue's priority")
	} else {
		g.ExpectWithOffset(1, priority).To(Equal(backendAssignedResources.Rank), "queue's priority")
	}

	validateOrgUnitResourcesMatching(g, queueResources, backendAssignedResources)
}
