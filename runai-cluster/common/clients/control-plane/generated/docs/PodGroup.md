# PodGroup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**JobId** | **string** |  | 
**PodGroupId** | **string** | Unique identifier of the pod group. | 
**PodGroupName** | Pointer to **string** | Name of the pod group. | [optional] 
**WorkloadName** | **string** | Name of the workload. | 
**WorkloadType** | **string** | Type of the workload. | 
**WorkloadKind** | **string** | Specifies the kind of k8s resource that owns the pod group. | 
**ClusterUuid** | **string** | Unique identifier of the cluster. | 
**ImageName** | **string** | The name of the image executed by the pod. | 
**PodGroupPhase** | Pointer to **string** | The phase of the pod group. | [optional] 
**JobPhase** | **string** |  | 
**WorkloadUser** | **string** | The owner of the pod group. | 
**Project** | **string** | The project that the pod group belongs to. | 
**NodeId** | Pointer to **string** | Unique identifier of the node. | [optional] 
**Created** | **int64** | Pod group creation time. | 
**LastUpdated** | **int64** | Time of the latest update to the pod group details. | 
**Parallelism** | **int32** | Specifies the maximum number of pods the workload is able run at any given time. | 
**Completions** | **int32** | Specifies the number of successfully finished pods the created workload should be run with. | 
**CurrentRequestedGpus** | **float32** |  | 
**TotalRequestedGpus** | **float32** |  | 
**RequestedGpusPerPodgroup** | **float32** |  | 
**RequestedGpusMemoryPerPodgroup** | **float32** |  | 
**ParentWorkloadName** | **string** | Specifies the Parent Workload Name | 
**CurrentAllocatedGpus** | Pointer to **float32** |  | [optional] 
**CurrentAllocatedGpusMemory** | Pointer to **float32** |  | [optional] 
**TotalRequestedCpu** | Pointer to **float32** |  | [optional] 
**TotalRequestedMemory** | Pointer to **float32** |  | [optional] 
**TotalLimitCpu** | Pointer to **float32** |  | [optional] 
**TotalLimitMemory** | Pointer to **float32** |  | [optional] 
**CliCommand** | Pointer to **string** |  | [optional] 
**RequestedMigDevices** | Pointer to **string** |  | [optional] 
**DynamicData** | Pointer to **map[string]interface{}** |  | [optional] 
**ExistsInCluster** | Pointer to **bool** |  | [optional] 
**JobUrlType** | Pointer to **string** | Describe the content of jobUrl field, values may be (tensorboard , jupyter or &#39;&#39;) | [optional] [default to ""]
**IsJupyter** | Pointer to **bool** | If true, it indicates that the pod group runs jupyter notebook. | [optional] [default to false]
**JobUrl** | Pointer to **string** |  | [optional] 
**WorkspaceId** | Pointer to **string** | this attribute is obsolete, use workloadId instead. | [optional] 
**WorkloadId** | Pointer to **string** | reference to the workspace which was used for the creation of the job. | [optional] 
**NodePool** | Pointer to **string** | The node pool of the podgroup. | [optional] 

## Methods

### NewPodGroup

`func NewPodGroup(jobId string, podGroupId string, workloadName string, workloadType string, workloadKind string, clusterUuid string, imageName string, jobPhase string, workloadUser string, project string, created int64, lastUpdated int64, parallelism int32, completions int32, currentRequestedGpus float32, totalRequestedGpus float32, requestedGpusPerPodgroup float32, requestedGpusMemoryPerPodgroup float32, parentWorkloadName string, ) *PodGroup`

NewPodGroup instantiates a new PodGroup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodGroupWithDefaults

`func NewPodGroupWithDefaults() *PodGroup`

NewPodGroupWithDefaults instantiates a new PodGroup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetJobId

`func (o *PodGroup) GetJobId() string`

GetJobId returns the JobId field if non-nil, zero value otherwise.

### GetJobIdOk

`func (o *PodGroup) GetJobIdOk() (*string, bool)`

GetJobIdOk returns a tuple with the JobId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobId

`func (o *PodGroup) SetJobId(v string)`

SetJobId sets JobId field to given value.


### GetPodGroupId

`func (o *PodGroup) GetPodGroupId() string`

GetPodGroupId returns the PodGroupId field if non-nil, zero value otherwise.

### GetPodGroupIdOk

`func (o *PodGroup) GetPodGroupIdOk() (*string, bool)`

GetPodGroupIdOk returns a tuple with the PodGroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupId

`func (o *PodGroup) SetPodGroupId(v string)`

SetPodGroupId sets PodGroupId field to given value.


### GetPodGroupName

`func (o *PodGroup) GetPodGroupName() string`

GetPodGroupName returns the PodGroupName field if non-nil, zero value otherwise.

### GetPodGroupNameOk

`func (o *PodGroup) GetPodGroupNameOk() (*string, bool)`

GetPodGroupNameOk returns a tuple with the PodGroupName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupName

`func (o *PodGroup) SetPodGroupName(v string)`

SetPodGroupName sets PodGroupName field to given value.

### HasPodGroupName

`func (o *PodGroup) HasPodGroupName() bool`

HasPodGroupName returns a boolean if a field has been set.

### GetWorkloadName

`func (o *PodGroup) GetWorkloadName() string`

GetWorkloadName returns the WorkloadName field if non-nil, zero value otherwise.

### GetWorkloadNameOk

`func (o *PodGroup) GetWorkloadNameOk() (*string, bool)`

GetWorkloadNameOk returns a tuple with the WorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadName

`func (o *PodGroup) SetWorkloadName(v string)`

SetWorkloadName sets WorkloadName field to given value.


### GetWorkloadType

`func (o *PodGroup) GetWorkloadType() string`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *PodGroup) GetWorkloadTypeOk() (*string, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *PodGroup) SetWorkloadType(v string)`

SetWorkloadType sets WorkloadType field to given value.


### GetWorkloadKind

`func (o *PodGroup) GetWorkloadKind() string`

GetWorkloadKind returns the WorkloadKind field if non-nil, zero value otherwise.

### GetWorkloadKindOk

`func (o *PodGroup) GetWorkloadKindOk() (*string, bool)`

GetWorkloadKindOk returns a tuple with the WorkloadKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadKind

`func (o *PodGroup) SetWorkloadKind(v string)`

SetWorkloadKind sets WorkloadKind field to given value.


### GetClusterUuid

`func (o *PodGroup) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *PodGroup) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *PodGroup) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetImageName

`func (o *PodGroup) GetImageName() string`

GetImageName returns the ImageName field if non-nil, zero value otherwise.

### GetImageNameOk

`func (o *PodGroup) GetImageNameOk() (*string, bool)`

GetImageNameOk returns a tuple with the ImageName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageName

`func (o *PodGroup) SetImageName(v string)`

SetImageName sets ImageName field to given value.


### GetPodGroupPhase

`func (o *PodGroup) GetPodGroupPhase() string`

GetPodGroupPhase returns the PodGroupPhase field if non-nil, zero value otherwise.

### GetPodGroupPhaseOk

`func (o *PodGroup) GetPodGroupPhaseOk() (*string, bool)`

GetPodGroupPhaseOk returns a tuple with the PodGroupPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupPhase

`func (o *PodGroup) SetPodGroupPhase(v string)`

SetPodGroupPhase sets PodGroupPhase field to given value.

### HasPodGroupPhase

`func (o *PodGroup) HasPodGroupPhase() bool`

HasPodGroupPhase returns a boolean if a field has been set.

### GetJobPhase

`func (o *PodGroup) GetJobPhase() string`

GetJobPhase returns the JobPhase field if non-nil, zero value otherwise.

### GetJobPhaseOk

`func (o *PodGroup) GetJobPhaseOk() (*string, bool)`

GetJobPhaseOk returns a tuple with the JobPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobPhase

`func (o *PodGroup) SetJobPhase(v string)`

SetJobPhase sets JobPhase field to given value.


### GetWorkloadUser

`func (o *PodGroup) GetWorkloadUser() string`

GetWorkloadUser returns the WorkloadUser field if non-nil, zero value otherwise.

### GetWorkloadUserOk

`func (o *PodGroup) GetWorkloadUserOk() (*string, bool)`

GetWorkloadUserOk returns a tuple with the WorkloadUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadUser

`func (o *PodGroup) SetWorkloadUser(v string)`

SetWorkloadUser sets WorkloadUser field to given value.


### GetProject

`func (o *PodGroup) GetProject() string`

GetProject returns the Project field if non-nil, zero value otherwise.

### GetProjectOk

`func (o *PodGroup) GetProjectOk() (*string, bool)`

GetProjectOk returns a tuple with the Project field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProject

`func (o *PodGroup) SetProject(v string)`

SetProject sets Project field to given value.


### GetNodeId

`func (o *PodGroup) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *PodGroup) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *PodGroup) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *PodGroup) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetCreated

`func (o *PodGroup) GetCreated() int64`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *PodGroup) GetCreatedOk() (*int64, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *PodGroup) SetCreated(v int64)`

SetCreated sets Created field to given value.


### GetLastUpdated

`func (o *PodGroup) GetLastUpdated() int64`

GetLastUpdated returns the LastUpdated field if non-nil, zero value otherwise.

### GetLastUpdatedOk

`func (o *PodGroup) GetLastUpdatedOk() (*int64, bool)`

GetLastUpdatedOk returns a tuple with the LastUpdated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUpdated

`func (o *PodGroup) SetLastUpdated(v int64)`

SetLastUpdated sets LastUpdated field to given value.


### GetParallelism

`func (o *PodGroup) GetParallelism() int32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *PodGroup) GetParallelismOk() (*int32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *PodGroup) SetParallelism(v int32)`

SetParallelism sets Parallelism field to given value.


### GetCompletions

`func (o *PodGroup) GetCompletions() int32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *PodGroup) GetCompletionsOk() (*int32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *PodGroup) SetCompletions(v int32)`

SetCompletions sets Completions field to given value.


### GetCurrentRequestedGpus

`func (o *PodGroup) GetCurrentRequestedGpus() float32`

GetCurrentRequestedGpus returns the CurrentRequestedGpus field if non-nil, zero value otherwise.

### GetCurrentRequestedGpusOk

`func (o *PodGroup) GetCurrentRequestedGpusOk() (*float32, bool)`

GetCurrentRequestedGpusOk returns a tuple with the CurrentRequestedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentRequestedGpus

`func (o *PodGroup) SetCurrentRequestedGpus(v float32)`

SetCurrentRequestedGpus sets CurrentRequestedGpus field to given value.


### GetTotalRequestedGpus

`func (o *PodGroup) GetTotalRequestedGpus() float32`

GetTotalRequestedGpus returns the TotalRequestedGpus field if non-nil, zero value otherwise.

### GetTotalRequestedGpusOk

`func (o *PodGroup) GetTotalRequestedGpusOk() (*float32, bool)`

GetTotalRequestedGpusOk returns a tuple with the TotalRequestedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedGpus

`func (o *PodGroup) SetTotalRequestedGpus(v float32)`

SetTotalRequestedGpus sets TotalRequestedGpus field to given value.


### GetRequestedGpusPerPodgroup

`func (o *PodGroup) GetRequestedGpusPerPodgroup() float32`

GetRequestedGpusPerPodgroup returns the RequestedGpusPerPodgroup field if non-nil, zero value otherwise.

### GetRequestedGpusPerPodgroupOk

`func (o *PodGroup) GetRequestedGpusPerPodgroupOk() (*float32, bool)`

GetRequestedGpusPerPodgroupOk returns a tuple with the RequestedGpusPerPodgroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedGpusPerPodgroup

`func (o *PodGroup) SetRequestedGpusPerPodgroup(v float32)`

SetRequestedGpusPerPodgroup sets RequestedGpusPerPodgroup field to given value.


### GetRequestedGpusMemoryPerPodgroup

`func (o *PodGroup) GetRequestedGpusMemoryPerPodgroup() float32`

GetRequestedGpusMemoryPerPodgroup returns the RequestedGpusMemoryPerPodgroup field if non-nil, zero value otherwise.

### GetRequestedGpusMemoryPerPodgroupOk

`func (o *PodGroup) GetRequestedGpusMemoryPerPodgroupOk() (*float32, bool)`

GetRequestedGpusMemoryPerPodgroupOk returns a tuple with the RequestedGpusMemoryPerPodgroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedGpusMemoryPerPodgroup

`func (o *PodGroup) SetRequestedGpusMemoryPerPodgroup(v float32)`

SetRequestedGpusMemoryPerPodgroup sets RequestedGpusMemoryPerPodgroup field to given value.


### GetParentWorkloadName

`func (o *PodGroup) GetParentWorkloadName() string`

GetParentWorkloadName returns the ParentWorkloadName field if non-nil, zero value otherwise.

### GetParentWorkloadNameOk

`func (o *PodGroup) GetParentWorkloadNameOk() (*string, bool)`

GetParentWorkloadNameOk returns a tuple with the ParentWorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentWorkloadName

`func (o *PodGroup) SetParentWorkloadName(v string)`

SetParentWorkloadName sets ParentWorkloadName field to given value.


### GetCurrentAllocatedGpus

`func (o *PodGroup) GetCurrentAllocatedGpus() float32`

GetCurrentAllocatedGpus returns the CurrentAllocatedGpus field if non-nil, zero value otherwise.

### GetCurrentAllocatedGpusOk

`func (o *PodGroup) GetCurrentAllocatedGpusOk() (*float32, bool)`

GetCurrentAllocatedGpusOk returns a tuple with the CurrentAllocatedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentAllocatedGpus

`func (o *PodGroup) SetCurrentAllocatedGpus(v float32)`

SetCurrentAllocatedGpus sets CurrentAllocatedGpus field to given value.

### HasCurrentAllocatedGpus

`func (o *PodGroup) HasCurrentAllocatedGpus() bool`

HasCurrentAllocatedGpus returns a boolean if a field has been set.

### GetCurrentAllocatedGpusMemory

`func (o *PodGroup) GetCurrentAllocatedGpusMemory() float32`

GetCurrentAllocatedGpusMemory returns the CurrentAllocatedGpusMemory field if non-nil, zero value otherwise.

### GetCurrentAllocatedGpusMemoryOk

`func (o *PodGroup) GetCurrentAllocatedGpusMemoryOk() (*float32, bool)`

GetCurrentAllocatedGpusMemoryOk returns a tuple with the CurrentAllocatedGpusMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentAllocatedGpusMemory

`func (o *PodGroup) SetCurrentAllocatedGpusMemory(v float32)`

SetCurrentAllocatedGpusMemory sets CurrentAllocatedGpusMemory field to given value.

### HasCurrentAllocatedGpusMemory

`func (o *PodGroup) HasCurrentAllocatedGpusMemory() bool`

HasCurrentAllocatedGpusMemory returns a boolean if a field has been set.

### GetTotalRequestedCpu

`func (o *PodGroup) GetTotalRequestedCpu() float32`

GetTotalRequestedCpu returns the TotalRequestedCpu field if non-nil, zero value otherwise.

### GetTotalRequestedCpuOk

`func (o *PodGroup) GetTotalRequestedCpuOk() (*float32, bool)`

GetTotalRequestedCpuOk returns a tuple with the TotalRequestedCpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedCpu

`func (o *PodGroup) SetTotalRequestedCpu(v float32)`

SetTotalRequestedCpu sets TotalRequestedCpu field to given value.

### HasTotalRequestedCpu

`func (o *PodGroup) HasTotalRequestedCpu() bool`

HasTotalRequestedCpu returns a boolean if a field has been set.

### GetTotalRequestedMemory

`func (o *PodGroup) GetTotalRequestedMemory() float32`

GetTotalRequestedMemory returns the TotalRequestedMemory field if non-nil, zero value otherwise.

### GetTotalRequestedMemoryOk

`func (o *PodGroup) GetTotalRequestedMemoryOk() (*float32, bool)`

GetTotalRequestedMemoryOk returns a tuple with the TotalRequestedMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedMemory

`func (o *PodGroup) SetTotalRequestedMemory(v float32)`

SetTotalRequestedMemory sets TotalRequestedMemory field to given value.

### HasTotalRequestedMemory

`func (o *PodGroup) HasTotalRequestedMemory() bool`

HasTotalRequestedMemory returns a boolean if a field has been set.

### GetTotalLimitCpu

`func (o *PodGroup) GetTotalLimitCpu() float32`

GetTotalLimitCpu returns the TotalLimitCpu field if non-nil, zero value otherwise.

### GetTotalLimitCpuOk

`func (o *PodGroup) GetTotalLimitCpuOk() (*float32, bool)`

GetTotalLimitCpuOk returns a tuple with the TotalLimitCpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalLimitCpu

`func (o *PodGroup) SetTotalLimitCpu(v float32)`

SetTotalLimitCpu sets TotalLimitCpu field to given value.

### HasTotalLimitCpu

`func (o *PodGroup) HasTotalLimitCpu() bool`

HasTotalLimitCpu returns a boolean if a field has been set.

### GetTotalLimitMemory

`func (o *PodGroup) GetTotalLimitMemory() float32`

GetTotalLimitMemory returns the TotalLimitMemory field if non-nil, zero value otherwise.

### GetTotalLimitMemoryOk

`func (o *PodGroup) GetTotalLimitMemoryOk() (*float32, bool)`

GetTotalLimitMemoryOk returns a tuple with the TotalLimitMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalLimitMemory

`func (o *PodGroup) SetTotalLimitMemory(v float32)`

SetTotalLimitMemory sets TotalLimitMemory field to given value.

### HasTotalLimitMemory

`func (o *PodGroup) HasTotalLimitMemory() bool`

HasTotalLimitMemory returns a boolean if a field has been set.

### GetCliCommand

`func (o *PodGroup) GetCliCommand() string`

GetCliCommand returns the CliCommand field if non-nil, zero value otherwise.

### GetCliCommandOk

`func (o *PodGroup) GetCliCommandOk() (*string, bool)`

GetCliCommandOk returns a tuple with the CliCommand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCliCommand

`func (o *PodGroup) SetCliCommand(v string)`

SetCliCommand sets CliCommand field to given value.

### HasCliCommand

`func (o *PodGroup) HasCliCommand() bool`

HasCliCommand returns a boolean if a field has been set.

### GetRequestedMigDevices

`func (o *PodGroup) GetRequestedMigDevices() string`

GetRequestedMigDevices returns the RequestedMigDevices field if non-nil, zero value otherwise.

### GetRequestedMigDevicesOk

`func (o *PodGroup) GetRequestedMigDevicesOk() (*string, bool)`

GetRequestedMigDevicesOk returns a tuple with the RequestedMigDevices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedMigDevices

`func (o *PodGroup) SetRequestedMigDevices(v string)`

SetRequestedMigDevices sets RequestedMigDevices field to given value.

### HasRequestedMigDevices

`func (o *PodGroup) HasRequestedMigDevices() bool`

HasRequestedMigDevices returns a boolean if a field has been set.

### GetDynamicData

`func (o *PodGroup) GetDynamicData() map[string]interface{}`

GetDynamicData returns the DynamicData field if non-nil, zero value otherwise.

### GetDynamicDataOk

`func (o *PodGroup) GetDynamicDataOk() (*map[string]interface{}, bool)`

GetDynamicDataOk returns a tuple with the DynamicData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDynamicData

`func (o *PodGroup) SetDynamicData(v map[string]interface{})`

SetDynamicData sets DynamicData field to given value.

### HasDynamicData

`func (o *PodGroup) HasDynamicData() bool`

HasDynamicData returns a boolean if a field has been set.

### GetExistsInCluster

`func (o *PodGroup) GetExistsInCluster() bool`

GetExistsInCluster returns the ExistsInCluster field if non-nil, zero value otherwise.

### GetExistsInClusterOk

`func (o *PodGroup) GetExistsInClusterOk() (*bool, bool)`

GetExistsInClusterOk returns a tuple with the ExistsInCluster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistsInCluster

`func (o *PodGroup) SetExistsInCluster(v bool)`

SetExistsInCluster sets ExistsInCluster field to given value.

### HasExistsInCluster

`func (o *PodGroup) HasExistsInCluster() bool`

HasExistsInCluster returns a boolean if a field has been set.

### GetJobUrlType

`func (o *PodGroup) GetJobUrlType() string`

GetJobUrlType returns the JobUrlType field if non-nil, zero value otherwise.

### GetJobUrlTypeOk

`func (o *PodGroup) GetJobUrlTypeOk() (*string, bool)`

GetJobUrlTypeOk returns a tuple with the JobUrlType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobUrlType

`func (o *PodGroup) SetJobUrlType(v string)`

SetJobUrlType sets JobUrlType field to given value.

### HasJobUrlType

`func (o *PodGroup) HasJobUrlType() bool`

HasJobUrlType returns a boolean if a field has been set.

### GetIsJupyter

`func (o *PodGroup) GetIsJupyter() bool`

GetIsJupyter returns the IsJupyter field if non-nil, zero value otherwise.

### GetIsJupyterOk

`func (o *PodGroup) GetIsJupyterOk() (*bool, bool)`

GetIsJupyterOk returns a tuple with the IsJupyter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsJupyter

`func (o *PodGroup) SetIsJupyter(v bool)`

SetIsJupyter sets IsJupyter field to given value.

### HasIsJupyter

`func (o *PodGroup) HasIsJupyter() bool`

HasIsJupyter returns a boolean if a field has been set.

### GetJobUrl

`func (o *PodGroup) GetJobUrl() string`

GetJobUrl returns the JobUrl field if non-nil, zero value otherwise.

### GetJobUrlOk

`func (o *PodGroup) GetJobUrlOk() (*string, bool)`

GetJobUrlOk returns a tuple with the JobUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobUrl

`func (o *PodGroup) SetJobUrl(v string)`

SetJobUrl sets JobUrl field to given value.

### HasJobUrl

`func (o *PodGroup) HasJobUrl() bool`

HasJobUrl returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *PodGroup) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *PodGroup) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *PodGroup) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.

### HasWorkspaceId

`func (o *PodGroup) HasWorkspaceId() bool`

HasWorkspaceId returns a boolean if a field has been set.

### GetWorkloadId

`func (o *PodGroup) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *PodGroup) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *PodGroup) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.

### HasWorkloadId

`func (o *PodGroup) HasWorkloadId() bool`

HasWorkloadId returns a boolean if a field has been set.

### GetNodePool

`func (o *PodGroup) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *PodGroup) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *PodGroup) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *PodGroup) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


