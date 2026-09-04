# Pod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PodId** | **string** | Identifier of the pod running the job. | 
**JobId** | **string** | Unique identifier of the job. | 
**PodGroupId** | Pointer to **string** | This had been used as jobId. Remained here for backward compatibility | [optional] 
**ClusterUuid** | **string** | Unique identifier of the cluster. | 
**PodName** | **string** | The name of the pod running the job. | 
**ImageName** | **string** | The name of the image executed by the pod. | 
**NodeId** | Pointer to **string** | Unique identifier of the node. | [optional] 
**Phase** | **string** |  | 
**Status** | Pointer to **string** |  | [optional] 
**Created** | **int64** | Creation time of the pod. | 
**Completed** | **int64** | Completion time of the pod. | 
**Started** | Pointer to **int64** | The time when the pod started executing. | [optional] 
**LastUpdated** | **int64** | Last time the pod details were updated. | 
**DynamicData** | Pointer to **map[string]interface{}** |  | [optional] 
**ExistsInCluster** | Pointer to **bool** |  | [optional] 
**ResourceRequest** | Pointer to **map[string]float32** |  | [optional] 
**ResourceAllocation** | Pointer to **map[string]float32** |  | [optional] 
**NodePool** | Pointer to **string** | The node pool of the pod. | [optional] 
**Namespace** | Pointer to **string** | The namespace of the pod. | [optional] 

## Methods

### NewPod

`func NewPod(podId string, jobId string, clusterUuid string, podName string, imageName string, phase string, created int64, completed int64, lastUpdated int64, ) *Pod`

NewPod instantiates a new Pod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodWithDefaults

`func NewPodWithDefaults() *Pod`

NewPodWithDefaults instantiates a new Pod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPodId

`func (o *Pod) GetPodId() string`

GetPodId returns the PodId field if non-nil, zero value otherwise.

### GetPodIdOk

`func (o *Pod) GetPodIdOk() (*string, bool)`

GetPodIdOk returns a tuple with the PodId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodId

`func (o *Pod) SetPodId(v string)`

SetPodId sets PodId field to given value.


### GetJobId

`func (o *Pod) GetJobId() string`

GetJobId returns the JobId field if non-nil, zero value otherwise.

### GetJobIdOk

`func (o *Pod) GetJobIdOk() (*string, bool)`

GetJobIdOk returns a tuple with the JobId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobId

`func (o *Pod) SetJobId(v string)`

SetJobId sets JobId field to given value.


### GetPodGroupId

`func (o *Pod) GetPodGroupId() string`

GetPodGroupId returns the PodGroupId field if non-nil, zero value otherwise.

### GetPodGroupIdOk

`func (o *Pod) GetPodGroupIdOk() (*string, bool)`

GetPodGroupIdOk returns a tuple with the PodGroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupId

`func (o *Pod) SetPodGroupId(v string)`

SetPodGroupId sets PodGroupId field to given value.

### HasPodGroupId

`func (o *Pod) HasPodGroupId() bool`

HasPodGroupId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Pod) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Pod) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Pod) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetPodName

`func (o *Pod) GetPodName() string`

GetPodName returns the PodName field if non-nil, zero value otherwise.

### GetPodNameOk

`func (o *Pod) GetPodNameOk() (*string, bool)`

GetPodNameOk returns a tuple with the PodName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodName

`func (o *Pod) SetPodName(v string)`

SetPodName sets PodName field to given value.


### GetImageName

`func (o *Pod) GetImageName() string`

GetImageName returns the ImageName field if non-nil, zero value otherwise.

### GetImageNameOk

`func (o *Pod) GetImageNameOk() (*string, bool)`

GetImageNameOk returns a tuple with the ImageName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageName

`func (o *Pod) SetImageName(v string)`

SetImageName sets ImageName field to given value.


### GetNodeId

`func (o *Pod) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *Pod) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *Pod) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *Pod) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetPhase

`func (o *Pod) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Pod) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Pod) SetPhase(v string)`

SetPhase sets Phase field to given value.


### GetStatus

`func (o *Pod) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Pod) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Pod) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Pod) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreated

`func (o *Pod) GetCreated() int64`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *Pod) GetCreatedOk() (*int64, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *Pod) SetCreated(v int64)`

SetCreated sets Created field to given value.


### GetCompleted

`func (o *Pod) GetCompleted() int64`

GetCompleted returns the Completed field if non-nil, zero value otherwise.

### GetCompletedOk

`func (o *Pod) GetCompletedOk() (*int64, bool)`

GetCompletedOk returns a tuple with the Completed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompleted

`func (o *Pod) SetCompleted(v int64)`

SetCompleted sets Completed field to given value.


### GetStarted

`func (o *Pod) GetStarted() int64`

GetStarted returns the Started field if non-nil, zero value otherwise.

### GetStartedOk

`func (o *Pod) GetStartedOk() (*int64, bool)`

GetStartedOk returns a tuple with the Started field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStarted

`func (o *Pod) SetStarted(v int64)`

SetStarted sets Started field to given value.

### HasStarted

`func (o *Pod) HasStarted() bool`

HasStarted returns a boolean if a field has been set.

### GetLastUpdated

`func (o *Pod) GetLastUpdated() int64`

GetLastUpdated returns the LastUpdated field if non-nil, zero value otherwise.

### GetLastUpdatedOk

`func (o *Pod) GetLastUpdatedOk() (*int64, bool)`

GetLastUpdatedOk returns a tuple with the LastUpdated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUpdated

`func (o *Pod) SetLastUpdated(v int64)`

SetLastUpdated sets LastUpdated field to given value.


### GetDynamicData

`func (o *Pod) GetDynamicData() map[string]interface{}`

GetDynamicData returns the DynamicData field if non-nil, zero value otherwise.

### GetDynamicDataOk

`func (o *Pod) GetDynamicDataOk() (*map[string]interface{}, bool)`

GetDynamicDataOk returns a tuple with the DynamicData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDynamicData

`func (o *Pod) SetDynamicData(v map[string]interface{})`

SetDynamicData sets DynamicData field to given value.

### HasDynamicData

`func (o *Pod) HasDynamicData() bool`

HasDynamicData returns a boolean if a field has been set.

### GetExistsInCluster

`func (o *Pod) GetExistsInCluster() bool`

GetExistsInCluster returns the ExistsInCluster field if non-nil, zero value otherwise.

### GetExistsInClusterOk

`func (o *Pod) GetExistsInClusterOk() (*bool, bool)`

GetExistsInClusterOk returns a tuple with the ExistsInCluster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistsInCluster

`func (o *Pod) SetExistsInCluster(v bool)`

SetExistsInCluster sets ExistsInCluster field to given value.

### HasExistsInCluster

`func (o *Pod) HasExistsInCluster() bool`

HasExistsInCluster returns a boolean if a field has been set.

### GetResourceRequest

`func (o *Pod) GetResourceRequest() map[string]float32`

GetResourceRequest returns the ResourceRequest field if non-nil, zero value otherwise.

### GetResourceRequestOk

`func (o *Pod) GetResourceRequestOk() (*map[string]float32, bool)`

GetResourceRequestOk returns a tuple with the ResourceRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceRequest

`func (o *Pod) SetResourceRequest(v map[string]float32)`

SetResourceRequest sets ResourceRequest field to given value.

### HasResourceRequest

`func (o *Pod) HasResourceRequest() bool`

HasResourceRequest returns a boolean if a field has been set.

### GetResourceAllocation

`func (o *Pod) GetResourceAllocation() map[string]float32`

GetResourceAllocation returns the ResourceAllocation field if non-nil, zero value otherwise.

### GetResourceAllocationOk

`func (o *Pod) GetResourceAllocationOk() (*map[string]float32, bool)`

GetResourceAllocationOk returns a tuple with the ResourceAllocation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceAllocation

`func (o *Pod) SetResourceAllocation(v map[string]float32)`

SetResourceAllocation sets ResourceAllocation field to given value.

### HasResourceAllocation

`func (o *Pod) HasResourceAllocation() bool`

HasResourceAllocation returns a boolean if a field has been set.

### GetNodePool

`func (o *Pod) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *Pod) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *Pod) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *Pod) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.

### GetNamespace

`func (o *Pod) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Pod) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Pod) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Pod) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


