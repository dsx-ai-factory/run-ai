# WorkloadPodsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectId** | **string** |  | 
**ClusterId** | Pointer to **string** |  | [optional] 
**WorkloadId** | **string** |  | 
**Pods** | [**[]PodSyncFields**](PodSyncFields.md) |  | 
**IdsToIgnore** | Pointer to **[]string** | pod ids to ignore when deleting. | [optional] 

## Methods

### NewWorkloadPodsSyncRequest

`func NewWorkloadPodsSyncRequest(projectId string, workloadId string, pods []PodSyncFields, ) *WorkloadPodsSyncRequest`

NewWorkloadPodsSyncRequest instantiates a new WorkloadPodsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadPodsSyncRequestWithDefaults

`func NewWorkloadPodsSyncRequestWithDefaults() *WorkloadPodsSyncRequest`

NewWorkloadPodsSyncRequestWithDefaults instantiates a new WorkloadPodsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectId

`func (o *WorkloadPodsSyncRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadPodsSyncRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadPodsSyncRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkloadPodsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadPodsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadPodsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *WorkloadPodsSyncRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetWorkloadId

`func (o *WorkloadPodsSyncRequest) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *WorkloadPodsSyncRequest) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *WorkloadPodsSyncRequest) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetPods

`func (o *WorkloadPodsSyncRequest) GetPods() []PodSyncFields`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *WorkloadPodsSyncRequest) GetPodsOk() (*[]PodSyncFields, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *WorkloadPodsSyncRequest) SetPods(v []PodSyncFields)`

SetPods sets Pods field to given value.


### GetIdsToIgnore

`func (o *WorkloadPodsSyncRequest) GetIdsToIgnore() []string`

GetIdsToIgnore returns the IdsToIgnore field if non-nil, zero value otherwise.

### GetIdsToIgnoreOk

`func (o *WorkloadPodsSyncRequest) GetIdsToIgnoreOk() (*[]string, bool)`

GetIdsToIgnoreOk returns a tuple with the IdsToIgnore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdsToIgnore

`func (o *WorkloadPodsSyncRequest) SetIdsToIgnore(v []string)`

SetIdsToIgnore sets IdsToIgnore field to given value.

### HasIdsToIgnore

`func (o *WorkloadPodsSyncRequest) HasIdsToIgnore() bool`

HasIdsToIgnore returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


