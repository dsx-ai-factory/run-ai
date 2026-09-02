# DisplayedJob1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**JobId** | Pointer to **string** | Unique identifier of the job. | [optional] 
**PodGroupId** | Pointer to **string** | Unique identifier of the pod group. | [optional] 
**JobName** | Pointer to **string** | The name of the job. | [optional] 
**JobType** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**StatusTooltip** | Pointer to **string** |  | [optional] 
**ImageName** | Pointer to **string** | The name of the image executed by the pod. | [optional] 
**User** | Pointer to **string** | The owner of the job. | [optional] 
**Project** | Pointer to **string** | The project that the pod group belongs to. | [optional] 
**NodeId** | Pointer to **string** | Unique identifier of the node. | [optional] 
**CreationTime** | Pointer to **string** | Creation time of the job. | [optional] 
**CompletionTime** | Pointer to **string** | Completion time of the job. | [optional] 
**TotalRuntime** | Pointer to **string** |  | [optional] 
**TotalWaitTime** | Pointer to **string** |  | [optional] 
**LastStatusUpdateTime** | Pointer to **float32** |  | [optional] 
**MsSinceLastStatusUpdate** | Pointer to **float32** |  | [optional] 
**Pending** | Pointer to **float32** |  | [optional] 
**Running** | Pointer to **float32** |  | [optional] 
**Parallelism** | Pointer to **float32** |  | [optional] 
**Completions** | Pointer to **float32** |  | [optional] 
**Failed** | Pointer to **float32** |  | [optional] 
**Succeeded** | Pointer to **float32** |  | [optional] 
**CurrentAllocatedGPUs** | Pointer to **float32** |  | [optional] 
**CurrentAllocatedGPUsMemory** | Pointer to **float32** |  | [optional] 
**CurrentRequestedGPUs** | Pointer to **float32** |  | [optional] 
**TotalRequestedGPUs** | Pointer to **float32** |  | [optional] 
**RequestedGPUsPerPodGroup** | Pointer to **float32** |  | [optional] 
**RequestedGPUsMemoryPerPodGroup** | Pointer to **float32** |  | [optional] 
**ParentWorkloadName** | Pointer to **string** |  | [optional] 
**TotalRequestedMemory** | Pointer to **float32** |  | [optional] 
**TotalRequestedCPU** | Pointer to **float32** |  | [optional] 
**TotalLimitCPU** | Pointer to **float32** |  | [optional] 
**TotalLimitMemory** | Pointer to **float32** |  | [optional] 
**WorkloadKind** | Pointer to **float32** | Specifies the kind of k8s resource that owns the pod group. | [optional] 
**CliCommand** | Pointer to **string** |  | [optional] 
**RequestedMigDevices** | Pointer to **string** |  | [optional] 
**DynamicData** | Pointer to **map[string]interface{}** |  | [optional] 
**ExistsInCluster** | Pointer to **bool** |  | [optional] 
**IsJupyter** | Pointer to **bool** | If true, it indicates that the pod group runs jupyter notebook. | [optional] [default to false]
**JobUrl** | Pointer to **string** |  | [optional] 
**NodePool** | Pointer to **string** | The node pool of the job. | [optional] 

## Methods

### NewDisplayedJob1

`func NewDisplayedJob1() *DisplayedJob1`

NewDisplayedJob1 instantiates a new DisplayedJob1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDisplayedJob1WithDefaults

`func NewDisplayedJob1WithDefaults() *DisplayedJob1`

NewDisplayedJob1WithDefaults instantiates a new DisplayedJob1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetJobId

`func (o *DisplayedJob1) GetJobId() string`

GetJobId returns the JobId field if non-nil, zero value otherwise.

### GetJobIdOk

`func (o *DisplayedJob1) GetJobIdOk() (*string, bool)`

GetJobIdOk returns a tuple with the JobId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobId

`func (o *DisplayedJob1) SetJobId(v string)`

SetJobId sets JobId field to given value.

### HasJobId

`func (o *DisplayedJob1) HasJobId() bool`

HasJobId returns a boolean if a field has been set.

### GetPodGroupId

`func (o *DisplayedJob1) GetPodGroupId() string`

GetPodGroupId returns the PodGroupId field if non-nil, zero value otherwise.

### GetPodGroupIdOk

`func (o *DisplayedJob1) GetPodGroupIdOk() (*string, bool)`

GetPodGroupIdOk returns a tuple with the PodGroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupId

`func (o *DisplayedJob1) SetPodGroupId(v string)`

SetPodGroupId sets PodGroupId field to given value.

### HasPodGroupId

`func (o *DisplayedJob1) HasPodGroupId() bool`

HasPodGroupId returns a boolean if a field has been set.

### GetJobName

`func (o *DisplayedJob1) GetJobName() string`

GetJobName returns the JobName field if non-nil, zero value otherwise.

### GetJobNameOk

`func (o *DisplayedJob1) GetJobNameOk() (*string, bool)`

GetJobNameOk returns a tuple with the JobName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobName

`func (o *DisplayedJob1) SetJobName(v string)`

SetJobName sets JobName field to given value.

### HasJobName

`func (o *DisplayedJob1) HasJobName() bool`

HasJobName returns a boolean if a field has been set.

### GetJobType

`func (o *DisplayedJob1) GetJobType() string`

GetJobType returns the JobType field if non-nil, zero value otherwise.

### GetJobTypeOk

`func (o *DisplayedJob1) GetJobTypeOk() (*string, bool)`

GetJobTypeOk returns a tuple with the JobType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobType

`func (o *DisplayedJob1) SetJobType(v string)`

SetJobType sets JobType field to given value.

### HasJobType

`func (o *DisplayedJob1) HasJobType() bool`

HasJobType returns a boolean if a field has been set.

### GetClusterId

`func (o *DisplayedJob1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DisplayedJob1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DisplayedJob1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DisplayedJob1) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetStatus

`func (o *DisplayedJob1) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DisplayedJob1) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DisplayedJob1) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DisplayedJob1) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetStatusTooltip

`func (o *DisplayedJob1) GetStatusTooltip() string`

GetStatusTooltip returns the StatusTooltip field if non-nil, zero value otherwise.

### GetStatusTooltipOk

`func (o *DisplayedJob1) GetStatusTooltipOk() (*string, bool)`

GetStatusTooltipOk returns a tuple with the StatusTooltip field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusTooltip

`func (o *DisplayedJob1) SetStatusTooltip(v string)`

SetStatusTooltip sets StatusTooltip field to given value.

### HasStatusTooltip

`func (o *DisplayedJob1) HasStatusTooltip() bool`

HasStatusTooltip returns a boolean if a field has been set.

### GetImageName

`func (o *DisplayedJob1) GetImageName() string`

GetImageName returns the ImageName field if non-nil, zero value otherwise.

### GetImageNameOk

`func (o *DisplayedJob1) GetImageNameOk() (*string, bool)`

GetImageNameOk returns a tuple with the ImageName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageName

`func (o *DisplayedJob1) SetImageName(v string)`

SetImageName sets ImageName field to given value.

### HasImageName

`func (o *DisplayedJob1) HasImageName() bool`

HasImageName returns a boolean if a field has been set.

### GetUser

`func (o *DisplayedJob1) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DisplayedJob1) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DisplayedJob1) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *DisplayedJob1) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetProject

`func (o *DisplayedJob1) GetProject() string`

GetProject returns the Project field if non-nil, zero value otherwise.

### GetProjectOk

`func (o *DisplayedJob1) GetProjectOk() (*string, bool)`

GetProjectOk returns a tuple with the Project field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProject

`func (o *DisplayedJob1) SetProject(v string)`

SetProject sets Project field to given value.

### HasProject

`func (o *DisplayedJob1) HasProject() bool`

HasProject returns a boolean if a field has been set.

### GetNodeId

`func (o *DisplayedJob1) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *DisplayedJob1) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *DisplayedJob1) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *DisplayedJob1) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetCreationTime

`func (o *DisplayedJob1) GetCreationTime() string`

GetCreationTime returns the CreationTime field if non-nil, zero value otherwise.

### GetCreationTimeOk

`func (o *DisplayedJob1) GetCreationTimeOk() (*string, bool)`

GetCreationTimeOk returns a tuple with the CreationTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTime

`func (o *DisplayedJob1) SetCreationTime(v string)`

SetCreationTime sets CreationTime field to given value.

### HasCreationTime

`func (o *DisplayedJob1) HasCreationTime() bool`

HasCreationTime returns a boolean if a field has been set.

### GetCompletionTime

`func (o *DisplayedJob1) GetCompletionTime() string`

GetCompletionTime returns the CompletionTime field if non-nil, zero value otherwise.

### GetCompletionTimeOk

`func (o *DisplayedJob1) GetCompletionTimeOk() (*string, bool)`

GetCompletionTimeOk returns a tuple with the CompletionTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletionTime

`func (o *DisplayedJob1) SetCompletionTime(v string)`

SetCompletionTime sets CompletionTime field to given value.

### HasCompletionTime

`func (o *DisplayedJob1) HasCompletionTime() bool`

HasCompletionTime returns a boolean if a field has been set.

### GetTotalRuntime

`func (o *DisplayedJob1) GetTotalRuntime() string`

GetTotalRuntime returns the TotalRuntime field if non-nil, zero value otherwise.

### GetTotalRuntimeOk

`func (o *DisplayedJob1) GetTotalRuntimeOk() (*string, bool)`

GetTotalRuntimeOk returns a tuple with the TotalRuntime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRuntime

`func (o *DisplayedJob1) SetTotalRuntime(v string)`

SetTotalRuntime sets TotalRuntime field to given value.

### HasTotalRuntime

`func (o *DisplayedJob1) HasTotalRuntime() bool`

HasTotalRuntime returns a boolean if a field has been set.

### GetTotalWaitTime

`func (o *DisplayedJob1) GetTotalWaitTime() string`

GetTotalWaitTime returns the TotalWaitTime field if non-nil, zero value otherwise.

### GetTotalWaitTimeOk

`func (o *DisplayedJob1) GetTotalWaitTimeOk() (*string, bool)`

GetTotalWaitTimeOk returns a tuple with the TotalWaitTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalWaitTime

`func (o *DisplayedJob1) SetTotalWaitTime(v string)`

SetTotalWaitTime sets TotalWaitTime field to given value.

### HasTotalWaitTime

`func (o *DisplayedJob1) HasTotalWaitTime() bool`

HasTotalWaitTime returns a boolean if a field has been set.

### GetLastStatusUpdateTime

`func (o *DisplayedJob1) GetLastStatusUpdateTime() float32`

GetLastStatusUpdateTime returns the LastStatusUpdateTime field if non-nil, zero value otherwise.

### GetLastStatusUpdateTimeOk

`func (o *DisplayedJob1) GetLastStatusUpdateTimeOk() (*float32, bool)`

GetLastStatusUpdateTimeOk returns a tuple with the LastStatusUpdateTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastStatusUpdateTime

`func (o *DisplayedJob1) SetLastStatusUpdateTime(v float32)`

SetLastStatusUpdateTime sets LastStatusUpdateTime field to given value.

### HasLastStatusUpdateTime

`func (o *DisplayedJob1) HasLastStatusUpdateTime() bool`

HasLastStatusUpdateTime returns a boolean if a field has been set.

### GetMsSinceLastStatusUpdate

`func (o *DisplayedJob1) GetMsSinceLastStatusUpdate() float32`

GetMsSinceLastStatusUpdate returns the MsSinceLastStatusUpdate field if non-nil, zero value otherwise.

### GetMsSinceLastStatusUpdateOk

`func (o *DisplayedJob1) GetMsSinceLastStatusUpdateOk() (*float32, bool)`

GetMsSinceLastStatusUpdateOk returns a tuple with the MsSinceLastStatusUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsSinceLastStatusUpdate

`func (o *DisplayedJob1) SetMsSinceLastStatusUpdate(v float32)`

SetMsSinceLastStatusUpdate sets MsSinceLastStatusUpdate field to given value.

### HasMsSinceLastStatusUpdate

`func (o *DisplayedJob1) HasMsSinceLastStatusUpdate() bool`

HasMsSinceLastStatusUpdate returns a boolean if a field has been set.

### GetPending

`func (o *DisplayedJob1) GetPending() float32`

GetPending returns the Pending field if non-nil, zero value otherwise.

### GetPendingOk

`func (o *DisplayedJob1) GetPendingOk() (*float32, bool)`

GetPendingOk returns a tuple with the Pending field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPending

`func (o *DisplayedJob1) SetPending(v float32)`

SetPending sets Pending field to given value.

### HasPending

`func (o *DisplayedJob1) HasPending() bool`

HasPending returns a boolean if a field has been set.

### GetRunning

`func (o *DisplayedJob1) GetRunning() float32`

GetRunning returns the Running field if non-nil, zero value otherwise.

### GetRunningOk

`func (o *DisplayedJob1) GetRunningOk() (*float32, bool)`

GetRunningOk returns a tuple with the Running field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunning

`func (o *DisplayedJob1) SetRunning(v float32)`

SetRunning sets Running field to given value.

### HasRunning

`func (o *DisplayedJob1) HasRunning() bool`

HasRunning returns a boolean if a field has been set.

### GetParallelism

`func (o *DisplayedJob1) GetParallelism() float32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *DisplayedJob1) GetParallelismOk() (*float32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *DisplayedJob1) SetParallelism(v float32)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *DisplayedJob1) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### GetCompletions

`func (o *DisplayedJob1) GetCompletions() float32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *DisplayedJob1) GetCompletionsOk() (*float32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *DisplayedJob1) SetCompletions(v float32)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *DisplayedJob1) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### GetFailed

`func (o *DisplayedJob1) GetFailed() float32`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *DisplayedJob1) GetFailedOk() (*float32, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *DisplayedJob1) SetFailed(v float32)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *DisplayedJob1) HasFailed() bool`

HasFailed returns a boolean if a field has been set.

### GetSucceeded

`func (o *DisplayedJob1) GetSucceeded() float32`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *DisplayedJob1) GetSucceededOk() (*float32, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *DisplayedJob1) SetSucceeded(v float32)`

SetSucceeded sets Succeeded field to given value.

### HasSucceeded

`func (o *DisplayedJob1) HasSucceeded() bool`

HasSucceeded returns a boolean if a field has been set.

### GetCurrentAllocatedGPUs

`func (o *DisplayedJob1) GetCurrentAllocatedGPUs() float32`

GetCurrentAllocatedGPUs returns the CurrentAllocatedGPUs field if non-nil, zero value otherwise.

### GetCurrentAllocatedGPUsOk

`func (o *DisplayedJob1) GetCurrentAllocatedGPUsOk() (*float32, bool)`

GetCurrentAllocatedGPUsOk returns a tuple with the CurrentAllocatedGPUs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentAllocatedGPUs

`func (o *DisplayedJob1) SetCurrentAllocatedGPUs(v float32)`

SetCurrentAllocatedGPUs sets CurrentAllocatedGPUs field to given value.

### HasCurrentAllocatedGPUs

`func (o *DisplayedJob1) HasCurrentAllocatedGPUs() bool`

HasCurrentAllocatedGPUs returns a boolean if a field has been set.

### GetCurrentAllocatedGPUsMemory

`func (o *DisplayedJob1) GetCurrentAllocatedGPUsMemory() float32`

GetCurrentAllocatedGPUsMemory returns the CurrentAllocatedGPUsMemory field if non-nil, zero value otherwise.

### GetCurrentAllocatedGPUsMemoryOk

`func (o *DisplayedJob1) GetCurrentAllocatedGPUsMemoryOk() (*float32, bool)`

GetCurrentAllocatedGPUsMemoryOk returns a tuple with the CurrentAllocatedGPUsMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentAllocatedGPUsMemory

`func (o *DisplayedJob1) SetCurrentAllocatedGPUsMemory(v float32)`

SetCurrentAllocatedGPUsMemory sets CurrentAllocatedGPUsMemory field to given value.

### HasCurrentAllocatedGPUsMemory

`func (o *DisplayedJob1) HasCurrentAllocatedGPUsMemory() bool`

HasCurrentAllocatedGPUsMemory returns a boolean if a field has been set.

### GetCurrentRequestedGPUs

`func (o *DisplayedJob1) GetCurrentRequestedGPUs() float32`

GetCurrentRequestedGPUs returns the CurrentRequestedGPUs field if non-nil, zero value otherwise.

### GetCurrentRequestedGPUsOk

`func (o *DisplayedJob1) GetCurrentRequestedGPUsOk() (*float32, bool)`

GetCurrentRequestedGPUsOk returns a tuple with the CurrentRequestedGPUs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentRequestedGPUs

`func (o *DisplayedJob1) SetCurrentRequestedGPUs(v float32)`

SetCurrentRequestedGPUs sets CurrentRequestedGPUs field to given value.

### HasCurrentRequestedGPUs

`func (o *DisplayedJob1) HasCurrentRequestedGPUs() bool`

HasCurrentRequestedGPUs returns a boolean if a field has been set.

### GetTotalRequestedGPUs

`func (o *DisplayedJob1) GetTotalRequestedGPUs() float32`

GetTotalRequestedGPUs returns the TotalRequestedGPUs field if non-nil, zero value otherwise.

### GetTotalRequestedGPUsOk

`func (o *DisplayedJob1) GetTotalRequestedGPUsOk() (*float32, bool)`

GetTotalRequestedGPUsOk returns a tuple with the TotalRequestedGPUs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedGPUs

`func (o *DisplayedJob1) SetTotalRequestedGPUs(v float32)`

SetTotalRequestedGPUs sets TotalRequestedGPUs field to given value.

### HasTotalRequestedGPUs

`func (o *DisplayedJob1) HasTotalRequestedGPUs() bool`

HasTotalRequestedGPUs returns a boolean if a field has been set.

### GetRequestedGPUsPerPodGroup

`func (o *DisplayedJob1) GetRequestedGPUsPerPodGroup() float32`

GetRequestedGPUsPerPodGroup returns the RequestedGPUsPerPodGroup field if non-nil, zero value otherwise.

### GetRequestedGPUsPerPodGroupOk

`func (o *DisplayedJob1) GetRequestedGPUsPerPodGroupOk() (*float32, bool)`

GetRequestedGPUsPerPodGroupOk returns a tuple with the RequestedGPUsPerPodGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedGPUsPerPodGroup

`func (o *DisplayedJob1) SetRequestedGPUsPerPodGroup(v float32)`

SetRequestedGPUsPerPodGroup sets RequestedGPUsPerPodGroup field to given value.

### HasRequestedGPUsPerPodGroup

`func (o *DisplayedJob1) HasRequestedGPUsPerPodGroup() bool`

HasRequestedGPUsPerPodGroup returns a boolean if a field has been set.

### GetRequestedGPUsMemoryPerPodGroup

`func (o *DisplayedJob1) GetRequestedGPUsMemoryPerPodGroup() float32`

GetRequestedGPUsMemoryPerPodGroup returns the RequestedGPUsMemoryPerPodGroup field if non-nil, zero value otherwise.

### GetRequestedGPUsMemoryPerPodGroupOk

`func (o *DisplayedJob1) GetRequestedGPUsMemoryPerPodGroupOk() (*float32, bool)`

GetRequestedGPUsMemoryPerPodGroupOk returns a tuple with the RequestedGPUsMemoryPerPodGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedGPUsMemoryPerPodGroup

`func (o *DisplayedJob1) SetRequestedGPUsMemoryPerPodGroup(v float32)`

SetRequestedGPUsMemoryPerPodGroup sets RequestedGPUsMemoryPerPodGroup field to given value.

### HasRequestedGPUsMemoryPerPodGroup

`func (o *DisplayedJob1) HasRequestedGPUsMemoryPerPodGroup() bool`

HasRequestedGPUsMemoryPerPodGroup returns a boolean if a field has been set.

### GetParentWorkloadName

`func (o *DisplayedJob1) GetParentWorkloadName() string`

GetParentWorkloadName returns the ParentWorkloadName field if non-nil, zero value otherwise.

### GetParentWorkloadNameOk

`func (o *DisplayedJob1) GetParentWorkloadNameOk() (*string, bool)`

GetParentWorkloadNameOk returns a tuple with the ParentWorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentWorkloadName

`func (o *DisplayedJob1) SetParentWorkloadName(v string)`

SetParentWorkloadName sets ParentWorkloadName field to given value.

### HasParentWorkloadName

`func (o *DisplayedJob1) HasParentWorkloadName() bool`

HasParentWorkloadName returns a boolean if a field has been set.

### GetTotalRequestedMemory

`func (o *DisplayedJob1) GetTotalRequestedMemory() float32`

GetTotalRequestedMemory returns the TotalRequestedMemory field if non-nil, zero value otherwise.

### GetTotalRequestedMemoryOk

`func (o *DisplayedJob1) GetTotalRequestedMemoryOk() (*float32, bool)`

GetTotalRequestedMemoryOk returns a tuple with the TotalRequestedMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedMemory

`func (o *DisplayedJob1) SetTotalRequestedMemory(v float32)`

SetTotalRequestedMemory sets TotalRequestedMemory field to given value.

### HasTotalRequestedMemory

`func (o *DisplayedJob1) HasTotalRequestedMemory() bool`

HasTotalRequestedMemory returns a boolean if a field has been set.

### GetTotalRequestedCPU

`func (o *DisplayedJob1) GetTotalRequestedCPU() float32`

GetTotalRequestedCPU returns the TotalRequestedCPU field if non-nil, zero value otherwise.

### GetTotalRequestedCPUOk

`func (o *DisplayedJob1) GetTotalRequestedCPUOk() (*float32, bool)`

GetTotalRequestedCPUOk returns a tuple with the TotalRequestedCPU field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequestedCPU

`func (o *DisplayedJob1) SetTotalRequestedCPU(v float32)`

SetTotalRequestedCPU sets TotalRequestedCPU field to given value.

### HasTotalRequestedCPU

`func (o *DisplayedJob1) HasTotalRequestedCPU() bool`

HasTotalRequestedCPU returns a boolean if a field has been set.

### GetTotalLimitCPU

`func (o *DisplayedJob1) GetTotalLimitCPU() float32`

GetTotalLimitCPU returns the TotalLimitCPU field if non-nil, zero value otherwise.

### GetTotalLimitCPUOk

`func (o *DisplayedJob1) GetTotalLimitCPUOk() (*float32, bool)`

GetTotalLimitCPUOk returns a tuple with the TotalLimitCPU field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalLimitCPU

`func (o *DisplayedJob1) SetTotalLimitCPU(v float32)`

SetTotalLimitCPU sets TotalLimitCPU field to given value.

### HasTotalLimitCPU

`func (o *DisplayedJob1) HasTotalLimitCPU() bool`

HasTotalLimitCPU returns a boolean if a field has been set.

### GetTotalLimitMemory

`func (o *DisplayedJob1) GetTotalLimitMemory() float32`

GetTotalLimitMemory returns the TotalLimitMemory field if non-nil, zero value otherwise.

### GetTotalLimitMemoryOk

`func (o *DisplayedJob1) GetTotalLimitMemoryOk() (*float32, bool)`

GetTotalLimitMemoryOk returns a tuple with the TotalLimitMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalLimitMemory

`func (o *DisplayedJob1) SetTotalLimitMemory(v float32)`

SetTotalLimitMemory sets TotalLimitMemory field to given value.

### HasTotalLimitMemory

`func (o *DisplayedJob1) HasTotalLimitMemory() bool`

HasTotalLimitMemory returns a boolean if a field has been set.

### GetWorkloadKind

`func (o *DisplayedJob1) GetWorkloadKind() float32`

GetWorkloadKind returns the WorkloadKind field if non-nil, zero value otherwise.

### GetWorkloadKindOk

`func (o *DisplayedJob1) GetWorkloadKindOk() (*float32, bool)`

GetWorkloadKindOk returns a tuple with the WorkloadKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadKind

`func (o *DisplayedJob1) SetWorkloadKind(v float32)`

SetWorkloadKind sets WorkloadKind field to given value.

### HasWorkloadKind

`func (o *DisplayedJob1) HasWorkloadKind() bool`

HasWorkloadKind returns a boolean if a field has been set.

### GetCliCommand

`func (o *DisplayedJob1) GetCliCommand() string`

GetCliCommand returns the CliCommand field if non-nil, zero value otherwise.

### GetCliCommandOk

`func (o *DisplayedJob1) GetCliCommandOk() (*string, bool)`

GetCliCommandOk returns a tuple with the CliCommand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCliCommand

`func (o *DisplayedJob1) SetCliCommand(v string)`

SetCliCommand sets CliCommand field to given value.

### HasCliCommand

`func (o *DisplayedJob1) HasCliCommand() bool`

HasCliCommand returns a boolean if a field has been set.

### GetRequestedMigDevices

`func (o *DisplayedJob1) GetRequestedMigDevices() string`

GetRequestedMigDevices returns the RequestedMigDevices field if non-nil, zero value otherwise.

### GetRequestedMigDevicesOk

`func (o *DisplayedJob1) GetRequestedMigDevicesOk() (*string, bool)`

GetRequestedMigDevicesOk returns a tuple with the RequestedMigDevices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedMigDevices

`func (o *DisplayedJob1) SetRequestedMigDevices(v string)`

SetRequestedMigDevices sets RequestedMigDevices field to given value.

### HasRequestedMigDevices

`func (o *DisplayedJob1) HasRequestedMigDevices() bool`

HasRequestedMigDevices returns a boolean if a field has been set.

### GetDynamicData

`func (o *DisplayedJob1) GetDynamicData() map[string]interface{}`

GetDynamicData returns the DynamicData field if non-nil, zero value otherwise.

### GetDynamicDataOk

`func (o *DisplayedJob1) GetDynamicDataOk() (*map[string]interface{}, bool)`

GetDynamicDataOk returns a tuple with the DynamicData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDynamicData

`func (o *DisplayedJob1) SetDynamicData(v map[string]interface{})`

SetDynamicData sets DynamicData field to given value.

### HasDynamicData

`func (o *DisplayedJob1) HasDynamicData() bool`

HasDynamicData returns a boolean if a field has been set.

### GetExistsInCluster

`func (o *DisplayedJob1) GetExistsInCluster() bool`

GetExistsInCluster returns the ExistsInCluster field if non-nil, zero value otherwise.

### GetExistsInClusterOk

`func (o *DisplayedJob1) GetExistsInClusterOk() (*bool, bool)`

GetExistsInClusterOk returns a tuple with the ExistsInCluster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistsInCluster

`func (o *DisplayedJob1) SetExistsInCluster(v bool)`

SetExistsInCluster sets ExistsInCluster field to given value.

### HasExistsInCluster

`func (o *DisplayedJob1) HasExistsInCluster() bool`

HasExistsInCluster returns a boolean if a field has been set.

### GetIsJupyter

`func (o *DisplayedJob1) GetIsJupyter() bool`

GetIsJupyter returns the IsJupyter field if non-nil, zero value otherwise.

### GetIsJupyterOk

`func (o *DisplayedJob1) GetIsJupyterOk() (*bool, bool)`

GetIsJupyterOk returns a tuple with the IsJupyter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsJupyter

`func (o *DisplayedJob1) SetIsJupyter(v bool)`

SetIsJupyter sets IsJupyter field to given value.

### HasIsJupyter

`func (o *DisplayedJob1) HasIsJupyter() bool`

HasIsJupyter returns a boolean if a field has been set.

### GetJobUrl

`func (o *DisplayedJob1) GetJobUrl() string`

GetJobUrl returns the JobUrl field if non-nil, zero value otherwise.

### GetJobUrlOk

`func (o *DisplayedJob1) GetJobUrlOk() (*string, bool)`

GetJobUrlOk returns a tuple with the JobUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobUrl

`func (o *DisplayedJob1) SetJobUrl(v string)`

SetJobUrl sets JobUrl field to given value.

### HasJobUrl

`func (o *DisplayedJob1) HasJobUrl() bool`

HasJobUrl returns a boolean if a field has been set.

### GetNodePool

`func (o *DisplayedJob1) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *DisplayedJob1) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *DisplayedJob1) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *DisplayedJob1) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


