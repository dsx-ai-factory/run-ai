/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"time"

	workloadsMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/adapters"
	"sigs.k8s.io/controller-runtime/pkg/client"

	auth_openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	openapi2 "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	assets_openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	org_unit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	policy_openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/datavolumes_openapi"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/backend_openapi"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/projects_openapi"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/handlers"
	"k8s.io/apimachinery/pkg/types"
)

const (
	Project1Name  = "team-a"
	Proj1Id       = 1
	Proj1IdStr    = "1"
	Project2Name  = "team-b"
	Proj2Id       = 2
	Proj2IdStr    = "2"
	Project3Name  = "team-c"
	Project3bName = "team-c"
	Project4Name  = "team-d"
	Project5Name  = "proj1"

	NodePoolDName        = "default"
	NodePool1Name        = "np-1"
	NodePool2Name        = "np-2"
	NodePool3Name        = "np-3"
	NodePoolDeletingName = "np-deleting"

	DataVolume1Name        = "dv1"
	DataVolume2Name        = "dv2"
	DataVolume3Name        = "dv3"
	DataVolume4Name        = "dv4"
	DataVolumeDeletingName = "dv-deleting"

	Department1Name       = "dep-1"
	Dep1Id                = 1
	Dep1IdStr             = "1"
	Department2Name       = "dep-2"
	Dep2Id                = 2
	Dep2IdStr             = "2"
	Department3Name       = "dep-3"
	Dep3Id                = 3
	Dep3IdStr             = "3"
	Department4Name       = "dep-4"
	Dep4Id                = 4
	Dep4IdStr             = "4"
	BadDeletionDepartment = "badDeletion"

	User1Name  = "user1"
	User2Name  = "user2"
	User3Name  = "user3"
	Group1Name = "group1"
	Group2Name = "group2"
	Group3Name = "group3"

	NamespaceForSpec1 = "team-a-ns"
	NamespaceForSpec2 = "runai-team-b"

	NodeAffinity1 = "some-key1"
	NodeAffinity2 = "some-key2"
	NodeAffinity3 = "some-key3"

	ResearchManagerUserRole = "research_manager"
	ResearcherUserRole      = "researcher"

	policy1ImageBeforeUpdate = "ubuntu"
	policyBackendImage       = "updated-image"
	policyBackendImageMaster = "updated-master-image"

	priorityClassAddedId      = "862e3223-c58a-4e27-b243-815e837fe846"
	priorityClassAddedName    = "added"
	priorityClassUpdatedId    = "c0136158-d54b-44ae-8b90-d05384cbd9fd"
	priorityClassUpdatedName  = "updated"
	priorityClassNonRunaiId   = "21cd0384-7f22-4a1d-8b10-2c7c63d6425d"
	priorityClassNonRunaiName = "nonRunai"
	priorityClassDeletedId    = "091b082a-93fa-47eb-843d-7e39cf671e2b"
	priorityClassDeletedName  = "deleted"
	priorityOriginalValue     = 1

	// Network Topology constants
	NetworkTopology1Name        = "topology-1"
	NetworkTopology2Name        = "topology-2"
	NetworkTopology3Name        = "topology-3"
	NetworkTopology4Name        = "topology-4"
	NetworkTopologyDeletingName = "topology-deleting"
	NetworkTopologyCluster1Name = "topology-cluster-1"
	NetworkTopologyCluster2Name = "topology-cluster-2"

	NodeLevelHostName  = "topology-hostname"
	NodeLevelHostRack  = "topology-rack"
	NodeLevelHostBlock = "topology-block"

	ClusterSource = "Cluster"

	// Karta constants
	KartaVersionV1          = "v1"
	KartaVersionV1Alpha2    = "v1alpha2"
	KartaKindDeployment     = "Deployment"
	KartaKindJob            = "Job"
	KartaKindRayCluster     = "RayCluster"
	KartaKindRayService     = "RayService"
	KartaGroupApps          = "apps"
	KartaGroupBatch         = "batch"
	KartaGroupRayIo         = "ray.io"
	KartaAppsDeploymentName = "krt-v1-apps-deployment"
	KartaBatchJobName       = "krt-v1-batch-job"
	KartaRayClusterName     = "krt-v1alpha2-ray-io-raycluster"
	KartaRayServiceName     = "krt-v1alpha2-ray-io-rayservice"
)

type PvcAndOptionalUpdate struct {
	Pvc    *corev1.PersistentVolumeClaim
	Update *assets_openapi.DatasourceUpdate
}

type SecretAndOptionalUpdate struct {
	Secret *corev1.Secret
	Update *assets_openapi.CredentialsUpdate
}

type WorkloadV2AndUpdate struct {
	Workload client.Object
	Update   *workloadsMgrApi.WorkloadV2Update
}

type policyTestData struct {
	// use -1 for unset fields
	name               string
	scope              policy_openapi.ScopeType
	scopeId            int // (-1 for tenant scope)
	namespace          string
	afterUpdateCounter int
	preUpdatedCounter  int // only if exist in the cluster before test
	effective          policy_openapi.SupersetPolicy
	distributed        policy_openapi.DistributedPolicyDefaultsAndRulesV2
	spec               interface{}
	masterSpec         interface{}
	toDelete           bool
}

var (
	trainingSpec = tpv1.TrainingSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Image: flds.NewStringField(policy1ImageBeforeUpdate),
			},
		},
	}

	inferenceSpec = infpv1.InferenceSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Image: flds.NewStringField(policy1ImageBeforeUpdate),
			},
		},
	}

	policy1 = policyTestData{ // exists - should be updated
		name:               "policy1",
		scope:              policy_openapi.SCOPETYPE_TENANT,
		scopeId:            -1,
		namespace:          constants.NamespaceRunAi,
		afterUpdateCounter: 11,
		preUpdatedCounter:  1,
		effective:          backendPolicyEffective,
		spec:               trainingSpec,
		toDelete:           false,
	}
	policy2 = policyTestData{ // don't exist - should be created
		name:               "policy2",
		scope:              policy_openapi.SCOPETYPE_DEPARTMENT,
		scopeId:            -1,
		namespace:          constants.NamespaceRunAi,
		afterUpdateCounter: 12,
		preUpdatedCounter:  -1,
		effective:          backendPolicyEffective,
		toDelete:           false,
	}
	policy3 = policyTestData{ // don't exist - should be created
		name:               "policy3",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 13,
		preUpdatedCounter:  -1,
		effective:          backendPolicyEffective,
		toDelete:           false,
	}
	policy4 = policyTestData{ // don't exist - should be created
		name:               "policy4",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 14,
		preUpdatedCounter:  -1,
		toDelete:           false,
	}
	policy5 = policyTestData{ // exists - should be deleted
		name:               "policy5",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 15,
		preUpdatedCounter:  -1,
		effective:          policy_openapi.SupersetPolicy{},
		spec:               tpv1.TrainingSpec{},
		toDelete:           true,
	}
	policy6 = policyTestData{ // existing distributed - should be deleted
		name:               "policy6",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 16,
		preUpdatedCounter:  -1,
		spec:               dpv1.DistributedSpec{},
		masterSpec:         &dpv1.DistributedMasterSpec{},
		toDelete:           true,
	}
	policy7 = policyTestData{ // newly created distributed - should be created
		name:               "policy7",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 17,
		preUpdatedCounter:  -1,
		distributed:        distributedPolicyEffective,
		spec:               dpv1.DistributedSpec{},
		masterSpec:         &dpv1.DistributedMasterSpec{},
		toDelete:           false,
	}
	policy8 = policyTestData{ // existing distributed
		name:               "policy8",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 10,
		preUpdatedCounter:  10,
		spec:               dpv1.DistributedSpec{},
		masterSpec:         &dpv1.DistributedMasterSpec{},
		toDelete:           true,
	}
	policy9 = policyTestData{ // exists - should be updated
		name:               "policy9",
		scope:              policy_openapi.SCOPETYPE_TENANT,
		scopeId:            -1,
		namespace:          constants.NamespaceRunAi,
		afterUpdateCounter: 11,
		preUpdatedCounter:  1,
		effective:          backendPolicyEffective,
		spec:               inferenceSpec,
		toDelete:           false,
	}
	policy10 = policyTestData{ // exists - should be deleted
		name:               "policy10",
		scope:              policy_openapi.SCOPETYPE_PROJECT,
		scopeId:            1,
		namespace:          NamespaceForSpec1,
		afterUpdateCounter: 15,
		preUpdatedCounter:  -1,
		effective:          policy_openapi.SupersetPolicy{},
		spec:               infpv1.InferenceSpec{},
		toDelete:           true,
	}

	BackendProject1 = projects_openapi.Project{
		Name:                              ptr.To(Project1Name),
		DepartmentName:                    ptr.To(Department1Name),
		DepartmentId:                      ptr.To(int32(Dep1Id)),
		DefaultNodePools:                  []string{NodePoolDName},
		InteractiveJobTimeLimitSecs:       ptr.To(float32(4)),
		InteractiveJobMaxIdleDurationSecs: ptr.To(float32(5)),
		InteractivePreemptibleJobMaxIdleDurationSecs: ptr.To(float32(10)),
		TrainingJobMaxIdleDurationSecs:               ptr.To(float32(2)),
		NodeAffinity: &projects_openapi.JobsNodeAffinity{
			Train: &projects_openapi.NodeAffinity{
				AffinityType: ptr.To("no_limit"),
				SelectedTypes: []projects_openapi.NodeTypeForProject{
					{
						Id:   ptr.To(float32(2)),
						Name: ptr.To(NodeAffinity2),
					},
					{
						Id:   ptr.To(float32(3)),
						Name: ptr.To(NodeAffinity3),
					},
				},
			},
			Interactive: &projects_openapi.NodeAffinity{
				AffinityType: ptr.To("only_selected"),
				SelectedTypes: []projects_openapi.NodeTypeForProject{
					{
						Id:   ptr.To(float32(1)),
						Name: ptr.To(NodeAffinity1),
					},
				},
			},
		},
		NodePoolsResources: []projects_openapi.NodePoolAssignedResources{
			{
				Id: ptr.To(float32(1)),
				NodePool: &projects_openapi.NodePoolResources{
					Id:   ptr.To(float32(1)),
					Name: ptr.To(NodePoolDName),
				},
				Gpu: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1)),
					MaxAllowed:      ptr.To(float32(-1)),
					OverQuotaWeight: ptr.To(float32(2)),
				},
				Cpu: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1000)),
					MaxAllowed:      nil,
					OverQuotaWeight: nil,
				},
				Memory: nil,
			},
			{
				Id: ptr.To(float32(2)),
				NodePool: &projects_openapi.NodePoolResources{
					Id:   ptr.To(float32(2)),
					Name: ptr.To(NodePool2Name),
				},
				Gpu: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1)),
					MaxAllowed:      ptr.To(float32(-1)),
					OverQuotaWeight: ptr.To(float32(2)),
				},
				Cpu: nil,
				Memory: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1000)),
					MaxAllowed:      ptr.To(float32(-1)),
					OverQuotaWeight: nil,
				},
			},
		},
		Namespace: ptr.To(NamespaceForSpec2),
		Id:        ptr.To(int32(Proj1Id)),

		Permissions: &projects_openapi.ResourcePermissions{
			Users:        []string{User1Name},
			Groups:       []string{Group1Name, Group2Name},
			Applications: []string{"some-app"},
		},
	}

	BackendProject2 = projects_openapi.Project{
		Name:                              ptr.To(Project2Name),
		DepartmentName:                    ptr.To(Department2Name),
		DepartmentId:                      ptr.To(int32(Dep2Id)),
		DefaultNodePools:                  []string{NodePool1Name},
		InteractiveJobTimeLimitSecs:       ptr.To(float32(1)),
		InteractiveJobMaxIdleDurationSecs: ptr.To(float32(1)),
		InteractivePreemptibleJobMaxIdleDurationSecs: ptr.To(float32(1000)),
		TrainingJobMaxIdleDurationSecs:               ptr.To(float32(2000)),
		NodeAffinity: &projects_openapi.JobsNodeAffinity{
			Train: &projects_openapi.NodeAffinity{
				AffinityType: ptr.To("only_selected"),
				SelectedTypes: []projects_openapi.NodeTypeForProject{
					{
						Id:   ptr.To(float32(2)),
						Name: ptr.To(NodeAffinity1),
					},
				},
			},
			Interactive: &projects_openapi.NodeAffinity{
				AffinityType:  ptr.To("no_limit"),
				SelectedTypes: []projects_openapi.NodeTypeForProject{},
			},
		},
		NodePoolsResources: []projects_openapi.NodePoolAssignedResources{
			{
				Id: ptr.To(float32(1)),
				NodePool: &projects_openapi.NodePoolResources{
					Id:   ptr.To(float32(1)),
					Name: ptr.To(NodePoolDName),
				},
				Gpu: &projects_openapi.Resource{},
				Cpu: nil,
				Memory: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(456464)),
					MaxAllowed:      ptr.To(float32(456464)),
					OverQuotaWeight: ptr.To(float32(2)),
				},
			},
			{
				Id: ptr.To(float32(2)),
				NodePool: &projects_openapi.NodePoolResources{
					Id:   ptr.To(float32(2)),
					Name: ptr.To(NodePool2Name),
				},
				Gpu: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1)),
					MaxAllowed:      ptr.To(float32(-1)),
					OverQuotaWeight: ptr.To(float32(2)),
				},
				Cpu: &projects_openapi.Resource{
					Deserved:        ptr.To(float32(1000)),
					MaxAllowed:      nil,
					OverQuotaWeight: nil,
				},
				Memory: nil,
			},
		},
		Namespace: ptr.To(NamespaceForSpec1),
		Id:        ptr.To(int32(Proj2Id)),

		Permissions: &projects_openapi.ResourcePermissions{
			Users:        []string{User1Name, User2Name},
			Applications: []string{"some-app2"},
		},
	}

	BackendProject3 = projects_openapi.Project{
		Name:           ptr.To(Project3Name),
		DepartmentName: ptr.To(Department1Name),
		DepartmentId:   ptr.To(int32(Dep1Id)),
		NodePoolsResources: []projects_openapi.NodePoolAssignedResources{
			{
				Id: ptr.To(float32(1)),
				NodePool: &projects_openapi.NodePoolResources{
					Id:   ptr.To(float32(1)),
					Name: ptr.To(NodePoolDName),
				},
				Gpu: nil,
				Cpu: &projects_openapi.Resource{
					Deserved:        nil,
					MaxAllowed:      ptr.To(float32(1000)),
					OverQuotaWeight: nil,
				},
			},
		},
		Id: ptr.To(int32(2)),
	}

	OrgUnitProject1 = org_unit_openapi.ProjectSync{
		Id:   Proj1IdStr,
		Name: Project1Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep1IdStr,
			Name: Department1Name,
		}),
		RequestedNamespace: *org_unit_openapi.NewNullableString(ptr.To(NamespaceForSpec2)),
		DefaultNodePools:   []string{NodePoolDName},
		SchedulingRules: *org_unit_openapi.NewNullableSchedulingRules(&org_unit_openapi.SchedulingRules{
			InteractiveJobTimeLimitSeconds:           *org_unit_openapi.NewNullableInt32(ptr.To(int32(4))),
			InteractiveJobMaxIdleDurationSeconds:     *org_unit_openapi.NewNullableInt32(ptr.To(int32(5))),
			InteractiveJobPreemptIdleDurationSeconds: *org_unit_openapi.NewNullableInt32(ptr.To(int32(10))),
			TrainingJobMaxIdleDurationSeconds:        *org_unit_openapi.NewNullableInt32(ptr.To(int32(2))),
		}),
		Resources: []org_unit_openapi.ResourcesSync{
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "1",
					Name: ptr.To(NodePoolDName),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved:        1,
					Limit:           nullableFloat(-1),
					OverQuotaWeight: nullableFloat(2),
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
					Deserved: nullableFloat(1000),
				}),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
				Rank:   ptr.To(int32(180)),
			},
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "2",
					Name: ptr.To(NodePool2Name),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved:        1,
					Limit:           nullableFloat(-1),
					OverQuotaWeight: nullableFloat(2),
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(nil),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(&org_unit_openapi.ResourcesSyncMemory{
					Deserved: nullableFloat(1000),
					Limit:    nullableFloat(-1),
				}),
				Rank: ptr.To(int32(180)),
			},
		},
		NodeTypes: &org_unit_openapi.NodeTypesPerWorkload{
			Training:  []string{"2", "3"},
			Workspace: []string{"1"},
			Names: map[string]string{
				"2": NodeAffinity2,
				"3": NodeAffinity3,
				"1": NodeAffinity1,
			},
		},
		Permissions: &org_unit_openapi.ProjectSyncPermissions{
			Users:        []string{User1Name},
			Groups:       []string{Group1Name, Group2Name},
			Applications: []string{"some-app"},
		},
	}

	OrgUnitProject2 = org_unit_openapi.ProjectSync{
		Id:   Proj2IdStr,
		Name: Project2Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep2IdStr,
			Name: Department2Name,
		}),
		RequestedNamespace: *org_unit_openapi.NewNullableString(ptr.To(NamespaceForSpec1)),
		DefaultNodePools:   []string{NodePool1Name},
		SchedulingRules: *org_unit_openapi.NewNullableSchedulingRules(&org_unit_openapi.SchedulingRules{
			InteractiveJobTimeLimitSeconds:           *org_unit_openapi.NewNullableInt32(ptr.To(int32(1))),
			InteractiveJobMaxIdleDurationSeconds:     *org_unit_openapi.NewNullableInt32(ptr.To(int32(1))),
			InteractiveJobPreemptIdleDurationSeconds: *org_unit_openapi.NewNullableInt32(ptr.To(int32(1000))),
			TrainingJobMaxIdleDurationSeconds:        *org_unit_openapi.NewNullableInt32(ptr.To(int32(2000))),
		}),
		Resources: []org_unit_openapi.ResourcesSync{
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "1",
					Name: ptr.To(NodePoolDName),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved: 0,
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(nil),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(&org_unit_openapi.ResourcesSyncMemory{
					Deserved:        nullableFloat(456464),
					Limit:           nullableFloat(456464),
					OverQuotaWeight: nullableFloat(2),
				}),
				Rank: ptr.To(int32(210)),
			},
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "2",
					Name: ptr.To(NodePool2Name),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved:        1,
					Limit:           nullableFloat(-1),
					OverQuotaWeight: nullableFloat(2),
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
					Deserved: nullableFloat(1000),
				}),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
				Rank:   ptr.To(int32(100)),
			},
		},
		NodeTypes: &org_unit_openapi.NodeTypesPerWorkload{
			Training: []string{"2"},
			Names: map[string]string{
				"2": NodeAffinity1,
			},
		},
		Permissions: &org_unit_openapi.ProjectSyncPermissions{
			Users:        []string{User1Name, User2Name},
			Applications: []string{"some-app2"},
		},
	}

	OrgUnitProject3del = org_unit_openapi.ProjectSync{
		Id:   "22",
		Name: Project3Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep1IdStr,
			Name: Department1Name,
		}),
		DeletedAt: *org_unit_openapi.NewNullableTime(ptr.To(time.Now())),
		Resources: []org_unit_openapi.ResourcesSync{
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "1",
					Name: ptr.To(NodePoolDName),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved: 0,
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
					Limit: nullableFloat(1000),
				}),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
				Rank:   ptr.To(int32(100)),
			},
		},
	}

	OrgUnitProject4delOtherId = org_unit_openapi.ProjectSync{
		Id:   "22000",
		Name: Project1Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep1IdStr,
			Name: Department1Name,
		}),
		RequestedNamespace: *org_unit_openapi.NewNullableString(ptr.To(NamespaceForSpec2)),
		DeletedAt:          *org_unit_openapi.NewNullableTime(ptr.To(time.Now())),
		Resources: []org_unit_openapi.ResourcesSync{
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "1",
					Name: ptr.To(NodePoolDName),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved: 0,
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
					Limit: nullableFloat(1000),
				}),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
				Rank:   ptr.To(int32(100)),
			},
		},
	}

	OrgUnitProject3 = org_unit_openapi.ProjectSync{
		Id:   "2",
		Name: Project3Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep1IdStr,
			Name: Department1Name,
		}),
		Resources: []org_unit_openapi.ResourcesSync{
			{
				NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
					Id:   "1",
					Name: ptr.To(NodePoolDName),
				}),
				Gpu: org_unit_openapi.ResourcesSyncGpu{
					Deserved: 0,
				},
				Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
					Limit: nullableFloat(1000),
				}),
				Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
				Rank:   ptr.To(int32(100)),
			},
		},
	}

	OrgUnitProject4 = org_unit_openapi.ProjectSync{
		Id:   "4",
		Name: Project4Name,
		Parent: *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
			Id:   Dep1IdStr,
			Name: Department1Name,
		}),
		DeletedAt: *org_unit_openapi.NewNullableTime(ptr.To(time.Now())),
	}

	BackendProjectsMap = map[string]*projects_openapi.Project{
		Project1Name: &BackendProject1,
		Project2Name: &BackendProject2,
		Project3Name: &BackendProject3,
	}

	OrgUnitProjectsMap = map[string]*org_unit_openapi.ProjectSync{
		Project1Name: &OrgUnitProject1,
		Project2Name: &OrgUnitProject2,
		Project3Name: &OrgUnitProject3,
		Project4Name: &OrgUnitProject4,
	}

	BackendDepartment1 = org_unit_openapi.DepartmentSync{
		Id:        Dep1IdStr,
		Name:      ptr.To(Department1Name),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		DeletedAt: org_unit_openapi.NullableTime{},
		Resources: []org_unit_openapi.ResourcesSync{{
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "1",
				Name: ptr.To(NodePoolDName),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(10),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(2),
			},
			Cpu: *org_unit_openapi.NewNullableResourcesCpu(nil),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(&org_unit_openapi.ResourcesSyncMemory{
				Deserved: nullableFloat(1000),
			}),
			Rank: ptr.To(int32(50)),
		}},
	}

	BackendDepartment2 = org_unit_openapi.DepartmentSync{
		Id:        Dep2IdStr,
		Name:      ptr.To(Department2Name),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		DeletedAt: org_unit_openapi.NullableTime{},
		Resources: []org_unit_openapi.ResourcesSync{{
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "1",
				Name: ptr.To(NodePoolDName),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(1),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(2),
			},
			Cpu: *org_unit_openapi.NewNullableResourcesCpu(nil),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(&org_unit_openapi.ResourcesSyncMemory{
				Deserved:        nullableFloat(456464),
				Limit:           nullableFloat(456464),
				OverQuotaWeight: nullableFloat(2),
			}),
			Rank: ptr.To(int32(50)),
		}, {
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "2",
				Name: ptr.To(NodePool2Name),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(1),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(2),
			},
			Cpu: *org_unit_openapi.NewNullableResourcesCpu(&org_unit_openapi.ResourcesCpu{
				Deserved: nullableFloat(1000),
			}),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
			Rank:   ptr.To(int32(50)),
		}},
	}

	BackendDepartment3 = org_unit_openapi.DepartmentSync{
		Name: ptr.To(Department3Name),
		Id:   Dep3IdStr,
		Resources: []org_unit_openapi.ResourcesSync{{
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "1",
				Name: ptr.To(NodePoolDName),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(0),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(0),
			},
			Cpu:    *org_unit_openapi.NewNullableResourcesCpu(nil),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
			Rank:   ptr.To(int32(50)),
		}},
	}

	BackendDepartment4 = org_unit_openapi.DepartmentSync{
		Name:      ptr.To(Department4Name),
		Id:        Dep4IdStr,
		DeletedAt: *org_unit_openapi.NewNullableTime(ptr.To(time.Now())),
		Resources: []org_unit_openapi.ResourcesSync{{
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "1",
				Name: ptr.To(NodePoolDName),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(0),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(0),
			},
			Cpu:    *org_unit_openapi.NewNullableResourcesCpu(nil),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
			Rank:   ptr.To(int32(50)),
		}},
	}

	BackendDepartment5 = org_unit_openapi.DepartmentSync{
		Name:      ptr.To(Department1Name),
		Id:        "121212",
		DeletedAt: *org_unit_openapi.NewNullableTime(ptr.To(time.Now())),
		Resources: []org_unit_openapi.ResourcesSync{{
			NodePool: *org_unit_openapi.NewNullableResourcesNodePool(&org_unit_openapi.ResourcesNodePool{
				Id:   "1",
				Name: ptr.To(NodePoolDName),
			}),
			Gpu: org_unit_openapi.ResourcesSyncGpu{
				Deserved:        float64(0),
				Limit:           nullableFloat(-1),
				OverQuotaWeight: nullableFloat(0),
			},
			Cpu:    *org_unit_openapi.NewNullableResourcesCpu(nil),
			Memory: *org_unit_openapi.NewNullableResourcesSyncMemory(nil),
			Rank:   ptr.To(int32(50)),
		}},
	}

	OrgUnitDepartmentsMap = map[string]*org_unit_openapi.DepartmentSync{
		Department1Name:       &BackendDepartment1,
		Department2Name:       &BackendDepartment2,
		Department3Name:       &BackendDepartment3,
		Department4Name:       &BackendDepartment4,
		BadDeletionDepartment: &BackendDepartment5,
	}

	OpenApiNodepoolD = openapi2.Nodepool{
		Name:  NodePoolDName,
		Id:    "100",
		Phase: ptr.To(openapi2.NODEPOOLPHASE_UPDATING),
	}

	OpenApiNodepool1 = openapi2.Nodepool{
		Name:         NodePool1Name,
		Id:           "1",
		LabelKey:     "some-label1",
		LabelValue:   "some-value1",
		Phase:        ptr.To(openapi2.NODEPOOLPHASE_READY),
		PhaseMessage: ptr.To("some-status-message"),
		Status: &openapi2.NodepoolSyncFieldsStatus{
			Nodes: []string{"node1, node2"},
		},
		IsDefault:                       true,
		GpuNetworkAccelerationDetection: *openapi2.NewNullableNodepoolGPUNetworkAccelerationDetection(ptr.To(openapi2.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE)),
		GpuNetworkAccelerationLabelKey:  ptr.To("GPU-DETECTION-LABEL"),
		SchedulingConfiguration: *openapi2.NewNullableSchedulingConfiguration(&openapi2.SchedulingConfiguration{
			PlacementStrategy: &openapi2.SchedulingConfigurationPlacementStrategy{
				Gpu: *openapi2.NewNullablePlacementStrategy(ptr.To(openapi2.PLACEMENTSTRATEGY_SPREAD)),
			},
			TimeBasedFairShare: *openapi2.NewNullableTimeBasedFairShare(&openapi2.TimeBasedFairShare{
				Enabled:                 *openapi2.NewNullableBool(ptr.To(true)),
				HistoricalUsageWeight:   *openapi2.NewNullableFloat32(ptr.To(float32(1.3))),
				DecayHalfLife:           *openapi2.NewNullableString(ptr.To("48h")),
				WindowDuration:          *openapi2.NewNullableString(ptr.To("168h")),
				WindowType:              *openapi2.NewNullableTimeBasedFairShareWindowType(ptr.To(openapi2.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING)),
				TumblingWindowStartTime: *openapi2.NewNullableTime(ptr.To(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))),
			}),
		}),
	}

	OpenApiNodepool2 = openapi2.Nodepool{
		Name:       NodePool2Name,
		Id:         "2",
		LabelKey:   "some-label2",
		LabelValue: "some-value2",
		Phase:      ptr.To(openapi2.NODEPOOLPHASE_CREATING),
	}

	OpenApiNodepool3 = openapi2.Nodepool{
		Name:       NodePool3Name,
		Id:         "3",
		LabelKey:   "some-label3",
		LabelValue: "some-value3",
		Phase:      ptr.To(openapi2.NODEPOOLPHASE_CREATING),
	}

	OpenApiNodepoolDeleting = openapi2.Nodepool{
		Name:       NodePoolDeletingName,
		Id:         "10",
		LabelKey:   "some-label-d",
		LabelValue: "some-value-d",
		Phase:      ptr.To(openapi2.NODEPOOLPHASE_DELETING),
	}

	OpenApiNodepoolsMap = map[string]*openapi2.Nodepool{
		NodePoolDName:        &OpenApiNodepoolD,
		NodePool1Name:        &OpenApiNodepool1,
		NodePool2Name:        &OpenApiNodepool2,
		NodePool3Name:        &OpenApiNodepool3,
		NodePoolDeletingName: &OpenApiNodepoolDeleting,
	}

	BackendDv1 = datavolumes_openapi.DatavolumeInternal{
		Id:                         "1",
		CreatedAt:                  ptr.To(time.Now()),
		UpdatedAt:                  ptr.To(time.Now()),
		CreatedBy:                  ptr.To("tester-a"),
		UpdatedBy:                  ptr.To("tester-b"),
		Name:                       DataVolume1Name,
		Description:                ptr.To("dv1-desc"),
		OriginPvcName:              "pvc1",
		ProjectId:                  "1",
		ShouldDeleteOriginalVolume: nil,
		ProjectName:                Project1Name,
		DepartmentId:               ptr.To(Department1Name),
		ClusterId:                  ptr.To(uuid.NewString()),
		SharedScopes: []datavolumes_openapi.SharedScope{
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_PROJECT,
				ScopeId:   "2",
			},
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_DEPARTMENT,
				ScopeId:   "2",
			},
		},
	}

	BackendDv2 = datavolumes_openapi.DatavolumeInternal{
		Id:                         "2",
		CreatedAt:                  ptr.To(time.Now()),
		UpdatedAt:                  ptr.To(time.Now()),
		CreatedBy:                  ptr.To("tester-a"),
		UpdatedBy:                  ptr.To("tester-b"),
		Name:                       DataVolume2Name,
		Description:                ptr.To("dv2 desc"),
		OriginPvcName:              "pvc2",
		ProjectId:                  "1",
		ShouldDeleteOriginalVolume: ptr.To(false),
		ProjectName:                Project1Name,
		DepartmentId:               ptr.To(Department1Name),
		ClusterId:                  ptr.To(uuid.NewString()),
		SharedScopes: []datavolumes_openapi.SharedScope{
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_DEPARTMENT,
				ScopeId:   "2",
			},
		},
	}

	BackendDv3 = datavolumes_openapi.DatavolumeInternal{
		Id:                         "3",
		CreatedAt:                  nil,
		UpdatedAt:                  nil,
		CreatedBy:                  nil,
		UpdatedBy:                  nil,
		Name:                       DataVolume3Name,
		Description:                nil,
		OriginPvcName:              "pvc1",
		ProjectId:                  "1",
		ShouldDeleteOriginalVolume: ptr.To(true),
		ProjectName:                Project2Name,
		DepartmentId:               nil,
		ClusterId:                  nil,
		SharedScopes: []datavolumes_openapi.SharedScope{
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_PROJECT,
				ScopeId:   "1",
			},
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_CLUSTER,
				ScopeId:   "2",
			},
		},
		Phase: ptr.To(datavolumes_openapi.DATAVOLUMEPHASE_READY),
	}

	BackendDv4 = datavolumes_openapi.DatavolumeInternal{
		Id:                         "4",
		CreatedAt:                  ptr.To(time.Now()),
		UpdatedAt:                  ptr.To(time.Now()),
		CreatedBy:                  ptr.To("tester-a"),
		UpdatedBy:                  ptr.To("tester-b"),
		Name:                       DataVolume4Name,
		Description:                ptr.To("dv1-desc"),
		OriginPvcName:              "pvc1",
		ProjectId:                  "1",
		ShouldDeleteOriginalVolume: nil,
		ProjectName:                Project1Name,
		DepartmentId:               ptr.To(Department1Name),
		ClusterId:                  ptr.To(uuid.NewString()),
		SharedScopes: []datavolumes_openapi.SharedScope{
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_PROJECT,
				ScopeId:   "2",
			},
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_DEPARTMENT,
				ScopeId:   "2",
			},
		},
		Phase: ptr.To(datavolumes_openapi.DATAVOLUMEPHASE_CREATING),
	}

	BackendDvDeleting = datavolumes_openapi.DatavolumeInternal{
		Id:                         "5",
		CreatedAt:                  ptr.To(time.Now()),
		UpdatedAt:                  ptr.To(time.Now()),
		CreatedBy:                  ptr.To("tester-a"),
		UpdatedBy:                  ptr.To("tester-b"),
		Name:                       DataVolumeDeletingName,
		Description:                ptr.To("dv1-desc"),
		OriginPvcName:              "pvc1",
		ProjectId:                  "1",
		ShouldDeleteOriginalVolume: nil,
		ProjectName:                Project1Name,
		DepartmentId:               ptr.To(Department1Name),
		ClusterId:                  ptr.To(uuid.NewString()),
		SharedScopes: []datavolumes_openapi.SharedScope{
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_PROJECT,
				ScopeId:   "2",
			},
			{
				ScopeType: datavolumes_openapi.SCOPETYPE_DEPARTMENT,
				ScopeId:   "2",
			},
		},
		Phase: ptr.To(datavolumes_openapi.DATAVOLUMEPHASE_DELETING),
	}

	BackendDatavolumesMap = map[string]*datavolumes_openapi.DatavolumeInternal{
		DataVolume1Name:        &BackendDv1,
		DataVolume2Name:        &BackendDv2,
		DataVolume3Name:        &BackendDv3,
		DataVolume4Name:        &BackendDv4,
		DataVolumeDeletingName: &BackendDvDeleting,
	}

	backendPolicyEffective = policy_openapi.SupersetPolicy{
		Defaults: *policy_openapi.NewNullableSupersetDefaults(&policy_openapi.SupersetDefaults{
			Image: *policy_openapi.NewNullableString(ptr.To(policyBackendImage)),
		}),
	}

	distributedPolicyEffective = policy_openapi.DistributedPolicyDefaultsAndRulesV2{
		Defaults: *policy_openapi.NewNullableDistributedPolicyDefaultsV2(&policy_openapi.DistributedPolicyDefaultsV2{
			Worker: *policy_openapi.NewNullableDistributedPolicyDefaultsV2Worker(&policy_openapi.DistributedPolicyDefaultsV2Worker{
				Image: *policy_openapi.NewNullableString(ptr.To(policyBackendImage)),
			}),
			Master: *policy_openapi.NewNullableReplicaDefaultsV2(&policy_openapi.ReplicaDefaultsV2{
				Image: *policy_openapi.NewNullableString(ptr.To(policyBackendImageMaster)),
			}),
		}),
	}

	BackendPolicy1  = generateBackendClusterPoliciesInner(policy1.name, policy1.scope, policy1.afterUpdateCounter, policy1.scopeId, policy1.toDelete, &policy1.effective, nil)
	BackendPolicy2  = generateBackendClusterPoliciesInner(policy2.name, policy2.scope, policy2.afterUpdateCounter, policy2.scopeId, policy2.toDelete, &policy2.effective, nil)
	BackendPolicy3  = generateBackendClusterPoliciesInner(policy3.name, policy3.scope, policy3.afterUpdateCounter, policy3.scopeId, policy3.toDelete, &policy3.effective, nil)
	BackendPolicy4  = generateBackendClusterPoliciesInner(policy4.name, policy4.scope, policy4.afterUpdateCounter, policy4.scopeId, policy4.toDelete, &policy4.effective, nil)
	BackendPolicy5  = generateBackendClusterPoliciesInner(policy5.name, policy5.scope, policy5.afterUpdateCounter, policy5.scopeId, policy5.toDelete, &policy5.effective, nil)
	BackendPolicy6  = generateBackendClusterPoliciesInner(policy6.name, policy6.scope, policy6.afterUpdateCounter, policy6.scopeId, policy6.toDelete, nil, &policy6.distributed)
	BackendPolicy7  = generateBackendClusterPoliciesInner(policy7.name, policy7.scope, policy7.afterUpdateCounter, policy7.scopeId, policy7.toDelete, nil, &policy7.distributed)
	BackendPolicy9  = generateBackendClusterPoliciesInner(policy9.name, policy9.scope, policy9.afterUpdateCounter, policy9.scopeId, policy9.toDelete, &policy9.effective, nil)
	BackendPolicy10 = generateBackendClusterPoliciesInner(policy10.name, policy10.scope, policy10.afterUpdateCounter, policy10.scopeId, policy10.toDelete, &policy10.effective, nil)

	pytorch = workloads_openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH

	/////////////////////////////////////////////////////////////
	// Workload 1: master and spec differ
	/////////////////////////////////////////////////////////////
	restartPolicyNever, _  = workloads_openapi.NewRestartPolicyFromValue("Never")
	restartPolicyAlways, _ = workloads_openapi.NewRestartPolicyFromValue("Always")

	BackendDistributedWorkload1 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(3)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-d1-1234",
			RequestedName: "workload-d1",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image:                *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				NumWorkers:           *workloads_openapi.NewNullableInt32(ptr.To(int32(1))),
				DistributedFramework: *workloads_openapi.NewNullableDistributedFramework(&pytorch),
				RestartPolicy:        *workloads_openapi.NewNullableString((*string)(restartPolicyNever)),
			},
			MasterSpec: *workloads_openapi.NewNullableMasterSpec(&workloads_openapi.MasterSpec{
				Image:         *workloads_openapi.NewNullableString(ptr.To("redhat")),
				RestartPolicy: *workloads_openapi.NewNullableRestartPolicy(restartPolicyAlways),
			}),
		},
	}

	BackendNonDistributedWorkload1 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(3)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-nd1-1234",
			RequestedName: "workload-nd1",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
			},
		},
	}

	BackendInferenceWorkload1 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(3)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-inf1-1234",
			RequestedName: "workload-inf1",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				ServingPort: *workloads_openapi.NewNullableSupersetServingPort(&workloads_openapi.SupersetServingPort{
					Container: *workloads_openapi.NewNullableInt32(ptr.To(int32(8080))),
				}),
			},
		},
	}

	ClusterInteractiveWorkload1 = &iwv2.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload1.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload1.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterTrainingWorkload1 = &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload1.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload1.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterDistributedWorkload1 = &dwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendDistributedWorkload1.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: dpv1.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendDistributedWorkload1.Workload.Spec.Image.Get()),
				},
			},
		},
		MasterSpec: &dpv1.DistributedMasterSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendDistributedWorkload1.Workload.MasterSpec.Get().Image.Get()),
				},
			},
		},
	}

	ClusterInferenceWorkload1 = &infwv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendInferenceWorkload1.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: infpv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendInferenceWorkload1.Workload.Spec.Image.Get()),
				},
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &flds.PortsField{
						Value: &map[string]flds.PortItemAndRules{
							"serving-port": {
								Value: &flds.PortItemValue{
									ContainerPort: ptr.To(8080),
									ServiceType:   ptr.To(constants.ServiceTypeServingPort),
								},
							},
						},
					},
				},
			},
		},
	}

	/////////////////////////////////////////////////////////////
	// Workload 2: no master
	/////////////////////////////////////////////////////////////

	BackendDistributedWorkload2 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(8)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-d2-1234",
			RequestedName: "workload-d2",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image:                *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				NumWorkers:           *workloads_openapi.NewNullableInt32(ptr.To(int32(2))),
				DistributedFramework: *workloads_openapi.NewNullableDistributedFramework(&pytorch),
			},
		},
	}

	BackendNonDistributedWorkload2 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(8)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-nd2-1234",
			RequestedName: "workload-nd2",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
			},
		},
	}

	BackendInferenceWorkload2 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(8)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-inf2-1234",
			RequestedName: "workload-inf2",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				ServingPort: *workloads_openapi.NewNullableSupersetServingPort(&workloads_openapi.SupersetServingPort{
					Container: *workloads_openapi.NewNullableInt32(ptr.To(int32(8080))),
				}),
			},
		},
	}

	ClusterInteractiveWorkload2 = &iwv2.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload2.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload2.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterTrainingWorkload2 = &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload2.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload2.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterDistributedWorkload2 = &dwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendDistributedWorkload2.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: dpv1.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendDistributedWorkload2.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterInferenceWorkload2 = &infwv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendInferenceWorkload2.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: infpv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendInferenceWorkload2.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	/////////////////////////////////////////////////////////////
	// Workload 3: master same as spec
	/////////////////////////////////////////////////////////////

	BackendDistributedWorkload3 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(9)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-d3-1234",
			RequestedName: "workload-d3",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image:                *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				NumWorkers:           *workloads_openapi.NewNullableInt32(ptr.To(int32(3))),
				DistributedFramework: *workloads_openapi.NewNullableDistributedFramework(&pytorch),
			},
			MasterSpecSameAsWorker: *workloads_openapi.NewNullableBool(ptr.To(true)),
		},
	}

	BackendNonDistributedWorkload3 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(9)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-nd3-1234",
			RequestedName: "workload-nd3",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
			},
		},
	}

	BackendInferenceWorkload3 = workloads_openapi.WorkloadsSyncInfo{
		UpdateCount: ptr.To(int64(9)),
		Workload: &workloads_openapi.SupersetWorkload{
			Name:          "workload-inf3-1234",
			RequestedName: "workload-inf3",
			WorkloadId:    uuid.NewString(),
			ProjectId:     "1",
			ClusterId:     TestClusterId,
			CreatedBy:     "tester-a",
			CreatedAt:     time.Now(),
			DesiredPhase:  workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			Spec: &workloads_openapi.SupersetSpec{
				Image: *workloads_openapi.NewNullableString(ptr.To("ubuntu")),
				ServingPort: *workloads_openapi.NewNullableSupersetServingPort(&workloads_openapi.SupersetServingPort{
					Container: *workloads_openapi.NewNullableInt32(ptr.To(int32(8080))),
				}),
			},
		},
	}

	ClusterInteractiveWorkload3 = &iwv2.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload3.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload3.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterTrainingWorkload3 = &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendNonDistributedWorkload3.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendNonDistributedWorkload3.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	ClusterDistributedWorkload3 = &dwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendDistributedWorkload3.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: dpv1.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendDistributedWorkload3.Workload.Spec.Image.Get()),
				},
			},
		},
		MasterSpec: &dpv1.DistributedMasterSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendDistributedWorkload3.Workload.Spec.Image.Get()), // taken from the spec!
				},
			},
		},
	}

	ClusterInferenceWorkload3 = &infwv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendInferenceWorkload3.Workload.Name,
			Namespace: NamespaceForSpec1,
		},
		Spec: infpv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: flds.NewStringField(*BackendInferenceWorkload3.Workload.Spec.Image.Get()),
				},
			},
		},
	}

	BackendAccessRule1 = auth_openapi.AccessRuleSync{
		ScopeId:     "8",
		ScopeType:   auth_openapi.SCOPETYPE_PROJECT,
		SubjectId:   "l1-researcher-587a-60ed061d6eb5@run.ai",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_USER,
		RoleId:      5,
		RoleName:    "L1 researcher",
		ScopeName:   Project5Name,
		Id:          59,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(-100 * time.Second),
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe",
		ClusterId:   ptr.To(TestClusterId),
	}

	BackendAccessRule2 = auth_openapi.AccessRuleSync{
		ScopeId:     "2",
		ScopeType:   auth_openapi.SCOPETYPE_TENANT,
		SubjectId:   "5",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_APP,
		RoleId:      2,
		RoleName:    "l1-researcher",
		ScopeName:   "",
		Id:          2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(-90 * time.Second),
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe2",
	}

	BackendAccessRule3 = auth_openapi.AccessRuleSync{
		ScopeId:     "3",
		ScopeType:   auth_openapi.SCOPETYPE_CLUSTER,
		SubjectId:   "4",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_GROUP,
		RoleId:      31,
		RoleName:    "custom-l2-researcher",
		ScopeName:   "",
		Id:          3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(-80 * time.Second),
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe3",
		PredefinedRole: *auth_openapi.NewNullableAccessRuleSyncAllOfPredefinedRole(&auth_openapi.AccessRuleSyncAllOfPredefinedRole{
			Id:   ptr.To("3"),
			Name: ptr.To("l2-researcher"),
		}),
	}

	BackendAccessRule4 = auth_openapi.AccessRuleSync{
		ScopeId:     "4",
		ScopeType:   auth_openapi.SCOPETYPE_DEPARTMENT,
		SubjectId:   "5",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_USER,
		RoleId:      4,
		RoleName:    "viewer",
		ScopeName:   Department4Name,
		Id:          4,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(-70 * time.Second),
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe4",
	}

	halfAMinuteAgo                = time.Now().Add(-30 * time.Second)
	BackendAccessRule5ToBeDeleted = auth_openapi.AccessRuleSync{
		ScopeId:     "5",
		ScopeType:   auth_openapi.SCOPETYPE_TENANT,
		SubjectId:   "6",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_USER,
		RoleId:      5,
		RoleName:    "user",
		ScopeName:   "",
		Id:          5,
		CreatedAt:   time.Now(),
		UpdatedAt:   halfAMinuteAgo,
		DeletedAt:   *auth_openapi.NewNullableTime(&halfAMinuteAgo),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe5",
	}
	BackendAccessRule6AlreadyDeleted = auth_openapi.AccessRuleSync{
		ScopeId:     "6",
		ScopeType:   auth_openapi.SCOPETYPE_PROJECT,
		SubjectId:   "7",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_GROUP,
		RoleId:      8,
		RoleName:    "admin",
		ScopeName:   Project4Name,
		Id:          6,
		CreatedAt:   time.Now(),
		UpdatedAt:   halfAMinuteAgo,
		DeletedAt:   *auth_openapi.NewNullableTime(&halfAMinuteAgo),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe6",
	}
	BackendAccessRule7FutureUpdateTime = auth_openapi.AccessRuleSync{
		ScopeId:     "7",
		ScopeType:   auth_openapi.SCOPETYPE_DEPARTMENT,
		SubjectId:   "8",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_USER,
		RoleId:      9,
		RoleName:    "researcher",
		ScopeName:   Department2Name,
		Id:          7,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(2 * time.Hour), // in the future - will be updated in the test
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
		TenantId:    ptr.To(int32(1)),
		CreatedBy:   "jane_doe6",
	}

	AccessRule1Name             = handlers.GenerateAccessRuleName(&BackendAccessRule1)
	AccessRule2Name             = handlers.GenerateAccessRuleName(&BackendAccessRule2)
	AccessRule3Name             = handlers.GenerateAccessRuleName(&BackendAccessRule3)
	AccessRule4Name             = handlers.GenerateAccessRuleName(&BackendAccessRule4)
	AccessRule5ToBeDeleted      = handlers.GenerateAccessRuleName(&BackendAccessRule5ToBeDeleted)
	AccessRule6AlreadyDeleted   = handlers.GenerateAccessRuleName(&BackendAccessRule6AlreadyDeleted)
	AccessRule7FutureUpdateTime = handlers.GenerateAccessRuleName(&BackendAccessRule7FutureUpdateTime)

	BackendAccessRulesMap = map[string]*auth_openapi.AccessRuleSync{
		AccessRule1Name:             &BackendAccessRule1,
		AccessRule2Name:             &BackendAccessRule2,
		AccessRule3Name:             &BackendAccessRule3,
		AccessRule4Name:             &BackendAccessRule4,
		AccessRule5ToBeDeleted:      &BackendAccessRule5ToBeDeleted,
		AccessRule6AlreadyDeleted:   &BackendAccessRule6AlreadyDeleted,
		AccessRule7FutureUpdateTime: &BackendAccessRule7FutureUpdateTime,
	}

	BackendUser1 = backend_openapi.User{
		UserId: User1Name,
		Role:   ResearcherUserRole,
	}
	BackendUser2 = backend_openapi.User{
		UserId: User2Name,
		Role:   ResearchManagerUserRole,
	}
	BackendUser3 = backend_openapi.User{
		UserId: User3Name,
		Role:   "random-role",
	}
	BackendGroup1 = backend_openapi.User{
		UserId: Group1Name,
		Role:   ResearcherUserRole,
	}
	BackendGroup2 = backend_openapi.User{
		UserId: Group2Name,
		Role:   ResearchManagerUserRole,
	}
	BackendGroup3 = backend_openapi.User{
		UserId: Group3Name,
		Role:   "random-role",
	}

	BackendUsersGroupsMap = map[string]*backend_openapi.User{
		User1Name:  &BackendUser1,
		User2Name:  &BackendUser2,
		User3Name:  &BackendUser3,
		Group1Name: &BackendGroup1,
		Group2Name: &BackendGroup2,
		Group3Name: &BackendGroup3,
	}

	//////////////////////////////////
	//   PVCs
	//////////////////////////////////

	existingProjPvcA = getBasicPvcObjectWithAnnotations("existing-proj-pvc-a-claim", NamespaceForSpec1,
		nil,
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		map[string]string{
			"attributes.run.ai/dnsip": "1.2.3.4",
		},
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("existing-proj-pvc-a-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1G"),
				},
			},
		},
	)

	existingDeptPvcA = getBasicPvcObject("existing-dept-pvc-a-claim", constants.NamespaceRunAi,
		nil,
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name, &corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("existing-dept-pvc-a-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("2G"),
				},
			},
		},
	)

	existingProjPvcB = getBasicPvcObject("existing-proj-pvc-b-claim", NamespaceForSpec1,
		convutils.Ref(int64(1)), // this pvc was previously synced as update #1
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("existing-proj-pvc-b-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3G"),
				},
			},
		},
	)

	existingDeptPvcB = getBasicPvcObject("existing-dept-pvc-b-claim", constants.NamespaceRunAi,
		convutils.Ref(int64(2)), // this pvc was previously synced as update #2
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("existing-dept-pvc-b-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("4G"),
				},
			},
		},
	)

	//
	//   update #3 we create a new project pvc
	//
	toCreateProjPvc = assets_openapi.DatasourceUpdate{
		UpdateCount: 3,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "newly-synced-proj-pvc",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_PVC,
		},
		Spec: &assets_openapi.DatasourceUpdateSpec{
			Pvc: *assets_openapi.NewNullablePVCAssetSpec(&assets_openapi.PVCAssetSpec{
				ClaimName: *assets_openapi.NewNullableString(convutils.Ref("newly-synced-proj-pvc-claim")),
				ClaimInfo: *assets_openapi.NewNullableClaimInfo(&assets_openapi.ClaimInfo{
					StorageClass: *assets_openapi.NewNullableString(convutils.Ref("newly-synced-proj-pvc-class")),
					Size:         *assets_openapi.NewNullableString(convutils.Ref("4G")),
					VolumeMode:   *assets_openapi.NewNullablePvcVolumeMode(convutils.Ref(assets_openapi.PVCVOLUMEMODE_FILESYSTEM)),
					AddedAttrValues: []assets_openapi.PvcAddedAttrValue{
						{
							Key:   "dnsname",
							Value: convutils.Ref("dns.nvidia.com"),
						},
						{
							Key:   "dnsip",
							Value: convutils.Ref("1.2.3.5"),
						},
					},
				}),
			}),
		},
	}

	syncedNewProjPvc = getBasicPvcObject("newly-synced-proj-pvc-claim", NamespaceForSpec1,
		convutils.Ref(int64(3)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("newly-synced-proj-pvc-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("4G"),
				},
			},
		},
	)

	//
	//   update #4 we delete project pvc b
	//
	toDeleteProjPvcB = assets_openapi.DatasourceUpdate{
		UpdateCount: 4,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "existing-proj-pvc-b",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_PVC,
		},
		IsDeleted: *assets_openapi.NewNullableBool(convutils.Ref(true)),
		Spec: &assets_openapi.DatasourceUpdateSpec{
			Pvc: *assets_openapi.NewNullablePVCAssetSpec(&assets_openapi.PVCAssetSpec{
				ClaimName: *assets_openapi.NewNullableString(convutils.Ref("existing-proj-pvc-b-claim")),
			}),
		},
	}

	//
	//   update #5 we create a new department pvc
	//
	toCreateDeptPvc = assets_openapi.DatasourceUpdate{
		UpdateCount: 5,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:        assets_openapi.SCOPE_DEPARTMENT,
			ClusterId:    *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			DepartmentId: *assets_openapi.NewNullableString(convutils.Ref(Department1Name)),
			Name:         "newly-synced-dept-pvc",
			Id:           uuid.NewString(),
			Kind:         assets_openapi.ASSETKIND_PVC,
		},
		Spec: &assets_openapi.DatasourceUpdateSpec{
			Pvc: *assets_openapi.NewNullablePVCAssetSpec(&assets_openapi.PVCAssetSpec{
				ClaimName: *assets_openapi.NewNullableString(convutils.Ref("newly-synced-dept-pvc-claim")),
				ClaimInfo: *assets_openapi.NewNullableClaimInfo(&assets_openapi.ClaimInfo{
					StorageClass: *assets_openapi.NewNullableString(convutils.Ref("newly-synced-dept-pvc-class")),
					Size:         *assets_openapi.NewNullableString(convutils.Ref("5G")),
					VolumeMode:   *assets_openapi.NewNullablePvcVolumeMode(convutils.Ref(assets_openapi.PVCVOLUMEMODE_FILESYSTEM)),
				}),
				DataSharing: *assets_openapi.NewNullableBool(utils.Ref(true)),
			}),
		},
	}

	syncedNewDeptPvc = getBasicPvcObject("newly-synced-dept-pvc-claim", constants.NamespaceRunAi,
		convutils.Ref(int64(5)),
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("newly-synced-dept-pvc-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5G"),
				},
			},
		},
	)

	//
	//   update #6 we modify the size of proj-pvc-a
	//
	toUpdateProjPvcA = assets_openapi.DatasourceUpdate{
		UpdateCount: 6,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "existing-proj-pvc-a",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_PVC,
		},
		Spec: &assets_openapi.DatasourceUpdateSpec{
			Pvc: *assets_openapi.NewNullablePVCAssetSpec(&assets_openapi.PVCAssetSpec{
				ClaimName: *assets_openapi.NewNullableString(convutils.Ref("existing-proj-pvc-a-claim")),
				ClaimInfo: *assets_openapi.NewNullableClaimInfo(&assets_openapi.ClaimInfo{
					StorageClass: *assets_openapi.NewNullableString(convutils.Ref("newly-synced-dept-pvc-class")),
					Size:         *assets_openapi.NewNullableString(convutils.Ref("6G")),
					VolumeMode:   *assets_openapi.NewNullablePvcVolumeMode(convutils.Ref(assets_openapi.PVCVOLUMEMODE_FILESYSTEM)),
				}),
			}),
		},
	}

	updatedProjPvcA = getBasicPvcObject("existing-proj-pvc-a-claim", NamespaceForSpec1,
		convutils.Ref(int64(6)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		&corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref("newly-synced-dept-pvc-class"),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("6G"),
				},
			},
		},
	)

	allDatasourceUpdates = []assets_openapi.DatasourceUpdate{
		toCreateProjPvc,
		toCreateDeptPvc,
		toDeleteProjPvcB,
		toUpdateProjPvcA,
	}

	postSyncUpdatePvcs = []PvcAndOptionalUpdate{
		{Pvc: syncedNewProjPvc, Update: &toCreateProjPvc},
		{Pvc: syncedNewDeptPvc, Update: &toCreateDeptPvc},
		{Pvc: updatedProjPvcA, Update: &toUpdateProjPvcA},
		{Pvc: existingDeptPvcA},
		{Pvc: existingDeptPvcB},
	}

	postSyncDeletedPvcs = []*corev1.PersistentVolumeClaim{
		existingProjPvcB,
	}

	//////////////////////////////////
	//   Secrets
	//////////////////////////////////

	existingProjPasswordA = getBasicSecretObject("password-pa", NamespaceForSpec1,
		nil,
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"password",
		map[string][]byte{
			"user":     []byte("my-user"),
			"password": []byte("my-password"),
		},
	)

	existingProjPasswordC = getBasicSecretObject("password-pc", NamespaceForSpec1,
		nil,
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"password",
		map[string][]byte{
			"user":     []byte("his-user"),
			"password": []byte("his-password"),
		},
	)

	existingDeptAccessKey = getBasicSecretObject("accesskey-aka", constants.NamespaceRunAi,
		nil,
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		"access-key",
		map[string][]byte{
			"accessKeyId":     []byte("my-id"),
			"accessSecretKey": []byte("my-secret-key"),
		},
	)

	existingProjDockerRegistry = getBasicSecretObject("dockerregistry-dra", NamespaceForSpec1,
		convutils.Ref(int64(7)), // this secret was previously synced as update #7
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"docker-registry",
		map[string][]byte{
			"user":     []byte("my-user"),
			"password": []byte("my-password"),
			"url":      []byte("http://docker-registry.com"),
		},
	)

	existingDeptGenericSecret = getBasicSecretObject("genericsecret-gsa", constants.NamespaceRunAi,
		convutils.Ref(int64(8)), // this secret was previously synced as update #8
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		"generic",
		map[string][]byte{
			"key-1": []byte("my-value-1"),
			"key-2": []byte("my-value-2"),
		},
	)

	//
	//   update #9 we create a new project password
	//
	toCreateProjPassword = assets_openapi.CredentialsUpdate{
		UpdateCount: 9,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "pb",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_PASSWORD,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			Password: *assets_openapi.NewNullablePasswordCreationSpec(&assets_openapi.PasswordCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("my-user-1")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("my-password-1")),
			}),
		},
	}

	syncedNewPasswordSecret = getBasicSecretObject("password-pb", NamespaceForSpec1,
		convutils.Ref(int64(9)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"password",
		map[string][]byte{
			"user":     []byte("my-user-2"),
			"password": []byte("my-password-2"),
		},
	)

	//
	//   update #10 we delete project access key
	//
	toDeleteDeptAccessKey = assets_openapi.CredentialsUpdate{
		UpdateCount: 10,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:        assets_openapi.SCOPE_DEPARTMENT,
			ClusterId:    *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			DepartmentId: *assets_openapi.NewNullableString(convutils.Ref(Department1Name)),
			Name:         "aka",
			Id:           uuid.NewString(),
			Kind:         assets_openapi.ASSETKIND_ACCESS_KEY,
		},
		IsDeleted: *assets_openapi.NewNullableBool(convutils.Ref(true)),
		Spec:      &assets_openapi.CredentialsUpdateSpec{},
	}

	//
	//   update #11 we create a new department docker registry
	//
	toCreateDeptDockerRegistry = assets_openapi.CredentialsUpdate{
		UpdateCount: 11,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:        assets_openapi.SCOPE_DEPARTMENT,
			ClusterId:    *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			DepartmentId: *assets_openapi.NewNullableString(convutils.Ref(Department1Name)),
			Name:         "drb",
			Id:           uuid.NewString(),
			Kind:         assets_openapi.ASSETKIND_DOCKER_REGISTRY,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			DockerRegistry: *assets_openapi.NewNullableDockerRegistryCreationSpec(&assets_openapi.DockerRegistryCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("my-dr-user")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("my-dr-pass")),
				Url:      *assets_openapi.NewNullableString(convutils.Ref("https://docker-registry.com")),
			}),
		},
	}

	syncedNewDockerRegistrySecret = getBasicSecretObject("dockerregistry-drb", constants.NamespaceRunAi,
		convutils.Ref(int64(11)),
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		"docker-registry",
		map[string][]byte{
			"user":     []byte("my-user-3"),
			"password": []byte("my-password-3"),
			"url":      []byte("https://docker-registy.com"),
		},
	)

	//
	//   update #12 we modify the data of the generic secret
	//
	toUpdateDeptGenericSecret = assets_openapi.CredentialsUpdate{
		UpdateCount: 12,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:        assets_openapi.SCOPE_DEPARTMENT,
			ClusterId:    *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			DepartmentId: *assets_openapi.NewNullableString(convutils.Ref(Department1Name)),
			Name:         "gsa",
			Id:           uuid.NewString(),
			Kind:         assets_openapi.ASSETKIND_GENERIC_SECRET,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			GenericSecret: *assets_openapi.NewNullableGenericSecretCreationSpec(&assets_openapi.GenericSecretCreationSpec{
				KeyValuePairs: []assets_openapi.KeyValuePair{
					{
						Key:   *assets_openapi.NewNullableString(convutils.Ref("aaa")),
						Value: *assets_openapi.NewNullableString(convutils.Ref("bbb")),
					},
				},
			}),
		},
	}

	updatedGenericSecret = getBasicSecretObject("genericsecret-gsa", constants.NamespaceRunAi,
		convutils.Ref(int64(6)),
		assets_openapi.SCOPETYPE_DEPARTMENT, Department1Name,
		"generic",
		map[string][]byte{
			"aaa": []byte("bbb"),
		},
	)

	//
	//   update #13 we rename existing password-pc
	//
	toRenamePasswordC = assets_openapi.CredentialsUpdate{
		UpdateCount: 13,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "renamed-pc",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_PASSWORD,
			PrevName:  convutils.Ref("pc"),
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			Password: *assets_openapi.NewNullablePasswordCreationSpec(&assets_openapi.PasswordCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("changed-user")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("changed-pass")),
			}),
		},
	}

	updatedPasswordC = getBasicSecretObject("password-renamed-pc", NamespaceForSpec1,
		convutils.Ref(int64(13)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"password",
		map[string][]byte{
			"changed-user": []byte("changed-pass"),
		},
	)

	//
	//   update #14 we create a new private generic secret
	//
	toCreatePrivateGenericSecret = assets_openapi.CredentialsUpdate{
		UpdateCount: 14,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_TENANT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			Name:      "private-gs",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_GENERIC_SECRET,
			Private:   *assets_openapi.NewNullableBool(convutils.Ref(true)),
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			GenericSecret: *assets_openapi.NewNullableGenericSecretCreationSpec(&assets_openapi.GenericSecretCreationSpec{
				KeyValuePairs: []assets_openapi.KeyValuePair{
					{
						Key:   *assets_openapi.NewNullableString(convutils.Ref("private-key-1")),
						Value: *assets_openapi.NewNullableString(convutils.Ref("private-value-1")),
					},
					{
						Key:   *assets_openapi.NewNullableString(convutils.Ref("private-key-2")),
						Value: *assets_openapi.NewNullableString(convutils.Ref("private-value-2")),
					},
				},
			}),
		},
	}

	syncedPrivateGenericSecret = getBasicSecretObject("private-gs-"+toCreatePrivateGenericSecret.Meta.Id[:10], constants.NamespaceRunAi,
		convutils.Ref(int64(14)),
		assets_openapi.SCOPETYPE_TENANT, "",
		"generic",
		map[string][]byte{
			"private-key-1": []byte("private-value-1"),
			"private-key-2": []byte("private-value-2"),
		},
	)

	//
	//   update #15 we create a new private docker registry
	//
	toCreatePrivateDockerRegistry = assets_openapi.CredentialsUpdate{
		UpdateCount: 15,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_TENANT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			Name:      "private-dr",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_DOCKER_REGISTRY,
			Private:   *assets_openapi.NewNullableBool(convutils.Ref(true)),
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			DockerRegistry: *assets_openapi.NewNullableDockerRegistryCreationSpec(&assets_openapi.DockerRegistryCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("private-user")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("private-pass")),
				Url:      *assets_openapi.NewNullableString(convutils.Ref("https://private-registry.com")),
			}),
		},
	}

	syncedPrivateDockerRegistry = getBasicSecretObject(
		"private-dr-"+toCreatePrivateDockerRegistry.Meta.Id[:10], constants.NamespaceRunAi,
		convutils.Ref(int64(15)), assets_openapi.SCOPETYPE_TENANT, "", "docker-registry", map[string][]byte{
			"user":     []byte("private-user"),
			"password": []byte("private-pass"),
			"url":      []byte("https://private-registry.com"),
		})

	toCreateNgcApiKey = assets_openapi.CredentialsUpdate{
		UpdateCount: 16,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "ngc-1",
			Id:        uuid.NewString(),
			Kind:      assets_openapi.ASSETKIND_NGC_API_KEY,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			GenericSecret: *assets_openapi.NewNullableGenericSecretCreationSpec(&assets_openapi.GenericSecretCreationSpec{
				KeyValuePairs: []assets_openapi.KeyValuePair{
					{
						Key:   *assets_openapi.NewNullableString(convutils.Ref("NGC_API_KEY")),
						Value: *assets_openapi.NewNullableString(convutils.Ref("ngc-key-1")),
					},
				},
			}),
			DockerRegistry: *assets_openapi.NewNullableDockerRegistryCreationSpec(&assets_openapi.DockerRegistryCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("user")),
				Url:      *assets_openapi.NewNullableString(convutils.Ref("url")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("ngc-key-1")),
			}),
		},
	}

	ngcGenericSecretExpected = assets_openapi.CredentialsUpdate{
		UpdateCount: 16,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "ngcgs-ngc-1",
			Id:        toCreateNgcApiKey.Meta.Id,
			Kind:      assets_openapi.ASSETKIND_GENERIC_SECRET,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			GenericSecret: *assets_openapi.NewNullableGenericSecretCreationSpec(&assets_openapi.GenericSecretCreationSpec{
				KeyValuePairs: []assets_openapi.KeyValuePair{
					{
						Key:   *assets_openapi.NewNullableString(convutils.Ref("NGC_API_KEY")),
						Value: *assets_openapi.NewNullableString(convutils.Ref("ngc-key-1")),
					},
				},
			}),
		},
	}

	ngcDrExpected = assets_openapi.CredentialsUpdate{
		UpdateCount: 16,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      "ngcdr-ngc-1",
			Id:        toCreateNgcApiKey.Meta.Id,
			Kind:      assets_openapi.ASSETKIND_DOCKER_REGISTRY,
		},
		Spec: &assets_openapi.CredentialsUpdateSpec{
			DockerRegistry: *assets_openapi.NewNullableDockerRegistryCreationSpec(&assets_openapi.DockerRegistryCreationSpec{
				User:     *assets_openapi.NewNullableString(convutils.Ref("user")),
				Password: *assets_openapi.NewNullableString(convutils.Ref("ngc-key-1")),
				Url:      *assets_openapi.NewNullableString(convutils.Ref("url")),
			}),
		},
	}

	syncedNgcGenericSecret = getBasicSecretObject("genericsecret-ngcgs-"+toCreateNgcApiKey.Meta.Name, NamespaceForSpec1,
		convutils.Ref(int64(14)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"generic",
		map[string][]byte{
			"NGC_API_KEY": []byte("ngc-key-1"),
		},
	)

	syncedNgcDockerRegister = getBasicSecretObject("dockerregistry-ngcdr-"+toCreateNgcApiKey.Meta.Name, NamespaceForSpec1,
		convutils.Ref(int64(14)),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		"docker-registry",
		map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("ngc-key-1"),
			"url":      []byte("url"),
		},
	)

	allCredentialsUpdates = []assets_openapi.CredentialsUpdate{
		toCreateProjPassword,
		toCreateDeptDockerRegistry,
		toDeleteDeptAccessKey,
		toUpdateDeptGenericSecret,
		toRenamePasswordC,
		toCreatePrivateGenericSecret,
		toCreatePrivateDockerRegistry,
		toCreateNgcApiKey,
	}

	postSyncUpdateSecrets = []SecretAndOptionalUpdate{
		{Secret: existingProjDockerRegistry},
		{Secret: existingProjPasswordA},
		{Secret: syncedNewPasswordSecret, Update: &toCreateProjPassword},
		{Secret: syncedNewDockerRegistrySecret, Update: &toCreateDeptDockerRegistry},
		{Secret: updatedGenericSecret, Update: &toUpdateDeptGenericSecret},
		{Secret: updatedPasswordC, Update: &toRenamePasswordC},
		{Secret: syncedPrivateGenericSecret, Update: &toCreatePrivateGenericSecret},
		{Secret: syncedPrivateDockerRegistry, Update: &toCreatePrivateDockerRegistry},
		{Secret: syncedNgcGenericSecret, Update: &ngcGenericSecretExpected},
		{Secret: syncedNgcDockerRegister, Update: &ngcDrExpected},
	}

	postSyncDeletedSecrets = []*corev1.Secret{
		existingDeptAccessKey,
		existingProjPasswordC,
	}
	// Network Topology backend data
	BackendNetworkTopology1 = openapi2.NetworkTopology{
		Name:   NetworkTopology1Name,
		Levels: []string{NodeLevelHostBlock, NodeLevelHostName},
		Source: ptr.To(adapters.NetworkTopologyRunAiSource),
	}

	BackendNetworkTopology2 = openapi2.NetworkTopology{
		Name:   NetworkTopology2Name,
		Levels: []string{NodeLevelHostRack, NodeLevelHostBlock, NodeLevelHostName},
		Source: ptr.To(adapters.NetworkTopologyRunAiSource),
	}

	BackendNetworkTopology3 = openapi2.NetworkTopology{
		Name:   NetworkTopology3Name,
		Levels: []string{NodeLevelHostName},
		Source: ptr.To(adapters.NetworkTopologyRunAiSource),
	}

	BackendNetworkTopology4 = openapi2.NetworkTopology{
		Name:   NetworkTopology4Name,
		Levels: []string{NodeLevelHostRack, NodeLevelHostName},
		Source: ptr.To(adapters.NetworkTopologyRunAiSource),
	}

	BackendNetworkTopologyDeleting = openapi2.NetworkTopology{
		Name:   NetworkTopologyDeletingName,
		Levels: []string{NodeLevelHostBlock},
		Source: ptr.To(adapters.NetworkTopologyRunAiSource),
	}

	BackendNetworkTopologyCluster1 = openapi2.NetworkTopology{
		Name:   NetworkTopologyCluster1Name,
		Levels: []string{NodeLevelHostBlock},
		Source: ptr.To(ClusterSource),
	}

	BackendNetworkTopologyCluster2 = openapi2.NetworkTopology{
		Name:   NetworkTopologyCluster2Name,
		Levels: []string{NodeLevelHostBlock},
		Source: ptr.To(ClusterSource),
	}

	BackendNetworkTopologiesMap = map[string]*openapi2.NetworkTopology{
		NetworkTopology1Name:        &BackendNetworkTopology1,
		NetworkTopology2Name:        &BackendNetworkTopology2,
		NetworkTopology3Name:        &BackendNetworkTopology3,
		NetworkTopology4Name:        &BackendNetworkTopology4,
		NetworkTopologyDeletingName: &BackendNetworkTopologyDeleting,
		NetworkTopologyCluster1Name: &BackendNetworkTopologyCluster1,
		NetworkTopologyCluster2Name: &BackendNetworkTopologyCluster2,
	}

	//////////////////////////////////
	//   Workloads V2
	//////////////////////////////////

	deploymentGvk = workloadsMgrApi.GVK{
		Version: "v1",
		Group:   *workloadsMgrApi.NewNullableString(convutils.Ref("apps")),
		Kind:    "Deployment",
	}

	existingV2WorkloadA = getBasicWorkloadV2("existing-v2-workload-a", NamespaceForSpec1, int64(1), generateSampleDeployment("app-a"))

	existingV2WorkloadB = getBasicWorkloadV2("existing-v2-workload-b", NamespaceForSpec1, int64(2), generateSampleDeployment("app-b"))

	//
	//   update #3 we create a new workload
	//
	toCreateWorkloadC = workloadsMgrApi.WorkloadV2Update{
		Id:               uuid.NewString(),
		Name:             "new-v2-workload-c",
		ProjectId:        convutils.Ref(Proj1IdStr),
		Gvk:              deploymentGvk,
		UpdateCount:      3,
		DesiredPhase:     workloadsMgrApi.DESIREDPHASE_RUNNING,
		WorkloadManifest: generateSampleDeployment("app-c"),
	}

	createdWorkloadC = getBasicWorkloadV2("new-v2-workload-c", NamespaceForSpec1, int64(3), generateSampleDeployment("app-c"))

	//
	//   update #4 delete a workload
	//
	toDeleteWorkloadA = workloadsMgrApi.WorkloadV2Update{
		Id:           uuid.NewString(),
		Name:         "existing-v2-workload-a",
		ProjectId:    convutils.Ref(Proj1IdStr),
		Gvk:          deploymentGvk,
		UpdateCount:  4,
		DesiredPhase: workloadsMgrApi.DESIREDPHASE_DELETED,
	}

	//----------------------------------------------------------
	// NOTE --> KEEP THIS ONE UPDATED WHEN YOU ADD UPDATES
	//----------------------------------------------------------
	lastWorkloadV2Update = toDeleteWorkloadA

	allWorkloadsV2Updates = []workloadsMgrApi.WorkloadV2Update{
		toCreateWorkloadC,
		toDeleteWorkloadA,
	}

	workloadsV2WithUpdates = []WorkloadV2AndUpdate{
		{Workload: existingV2WorkloadA, Update: &toDeleteWorkloadA},
		{Workload: existingV2WorkloadB},
		{Workload: createdWorkloadC, Update: &toCreateWorkloadC},
	}

	backendWorkloadTypeConfigs = []workloads_openapi.WorkloadTypeConfig{
		{
			CategoryId:     "cat1",
			PriorityId:     "pri1",
			Name:           "Deployment",
			Id:             ptr.To("id1"),
			CategoryName:   ptr.To("CategoryA"),
			PriorityName:   ptr.To("PriorityA"),
			Preemptibility: workloads_openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
			UpdatedAt:      ptr.To(time.Now()),
			UpdatedBy:      ptr.To("userA"),
			Group:          "apps",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: ptr.To("v1"), UpdatedAt: ptr.To(time.Now()), Karta: buildResourceDefinitionForGvk("apps", "v1", "Deployment")},
			},
		},
		{
			CategoryId:     "cat2",
			PriorityId:     "pri2",
			Name:           "RayCluster",
			Id:             ptr.To("id2"),
			CategoryName:   ptr.To("CategoryB"),
			PriorityName:   ptr.To("PriorityB"),
			Preemptibility: workloads_openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
			UpdatedAt:      ptr.To(time.Now()),
			UpdatedBy:      ptr.To("userB"),
			Group:          "ray.io",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: ptr.To("v1alpha2"), UpdatedAt: ptr.To(time.Now()), Karta: buildResourceDefinitionForGvk("ray.io", "v1alpha2", "RayCluster")},
			},
		},
		{
			CategoryId:     "cat2",
			PriorityId:     "pri2",
			Name:           "RayService",
			Id:             ptr.To("id3"),
			CategoryName:   ptr.To("CategoryB"),
			PriorityName:   ptr.To("PriorityB"),
			Preemptibility: workloads_openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
			UpdatedAt:      ptr.To(time.Now()),
			UpdatedBy:      ptr.To("userB"),
			Group:          "ray.io",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: ptr.To("v1alpha2"), UpdatedAt: ptr.To(time.Now()), Karta: buildResourceDefinitionForGvk("ray.io", "v1alpha2", "RayService")},
			},
		},
	}
)

func nullableFloat(val float64) org_unit_openapi.NullableFloat64 {
	return *org_unit_openapi.NewNullableFloat64(ptr.To(val))
}

func getBasicNetworkTopologyObject(name string) *kaiv1alpha1.Topology {
	// Get the backend spec for this topology to create a matching object
	backendTopology, exists := BackendNetworkTopologiesMap[name]
	if !exists {
		// Fallback to default levels if backend spec doesn't exist
		return &kaiv1alpha1.Topology{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					adapters.NetworkTopologySourceLabel: adapters.NetworkTopologyRunAiSource,
				},
			},
			Spec: kaiv1alpha1.TopologySpec{
				Levels: []kaiv1alpha1.TopologyLevel{
					{NodeLabel: NodeLevelHostBlock},
				},
			},
		}
	}

	// Create levels from backend spec
	levels := adapters.NewNetworkTopologyLevelsAdapter(backendTopology.Levels).
		ToClusterTopologyLevels()
	return &kaiv1alpha1.Topology{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				adapters.NetworkTopologySourceLabel: adapters.NetworkTopologyRunAiSource,
			},
		},
		Spec: kaiv1alpha1.TopologySpec{
			Levels: levels,
		},
	}
}

func getBasicClusterNetworkTopologyObject(name string) *kaiv1alpha1.Topology {
	return &kaiv1alpha1.Topology{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				adapters.NetworkTopologySourceLabel: ClusterSource,
			},
		},
		Spec: kaiv1alpha1.TopologySpec{
			Levels: []kaiv1alpha1.TopologyLevel{
				{NodeLabel: "topology-block"},
			},
		},
	}
}

// getBasicKartaObject creates a basic Karta object with the given name and GVK
func getBasicKartaObject(name, group, version, kind string) *kartav1alpha1.Karta {
	r := &kartav1alpha1.Karta{}
	r.Name = name
	r.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	r.CreationTimestamp = metav1.Now()
	r.UID = types.UID(uuid.New().String())
	return r
}
