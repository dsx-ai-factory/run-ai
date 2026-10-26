# WorkloadsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectId** | **string** |  | 
**ClusterId** | Pointer to **string** |  | [optional] 
**Workloads** | [**[]WorkloadSyncFields**](WorkloadSyncFields.md) |  | 
**From** | Pointer to **string** | the first workload id to sync. if empty, all workloads before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last workload to sync. if empty, all workloads after the last one sent will be deleted. | [optional] 
**IdsToIgnore** | Pointer to **[]string** | workload ids to ignore when deleting. | [optional] 

## Methods

### NewWorkloadsSyncRequest

`func NewWorkloadsSyncRequest(projectId string, workloads []WorkloadSyncFields, ) *WorkloadsSyncRequest`

NewWorkloadsSyncRequest instantiates a new WorkloadsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsSyncRequestWithDefaults

`func NewWorkloadsSyncRequestWithDefaults() *WorkloadsSyncRequest`

NewWorkloadsSyncRequestWithDefaults instantiates a new WorkloadsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectId

`func (o *WorkloadsSyncRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadsSyncRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadsSyncRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkloadsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *WorkloadsSyncRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetWorkloads

`func (o *WorkloadsSyncRequest) GetWorkloads() []WorkloadSyncFields`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *WorkloadsSyncRequest) GetWorkloadsOk() (*[]WorkloadSyncFields, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *WorkloadsSyncRequest) SetWorkloads(v []WorkloadSyncFields)`

SetWorkloads sets Workloads field to given value.


### GetFrom

`func (o *WorkloadsSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *WorkloadsSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *WorkloadsSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *WorkloadsSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *WorkloadsSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *WorkloadsSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *WorkloadsSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *WorkloadsSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.

### GetIdsToIgnore

`func (o *WorkloadsSyncRequest) GetIdsToIgnore() []string`

GetIdsToIgnore returns the IdsToIgnore field if non-nil, zero value otherwise.

### GetIdsToIgnoreOk

`func (o *WorkloadsSyncRequest) GetIdsToIgnoreOk() (*[]string, bool)`

GetIdsToIgnoreOk returns a tuple with the IdsToIgnore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdsToIgnore

`func (o *WorkloadsSyncRequest) SetIdsToIgnore(v []string)`

SetIdsToIgnore sets IdsToIgnore field to given value.

### HasIdsToIgnore

`func (o *WorkloadsSyncRequest) HasIdsToIgnore() bool`

HasIdsToIgnore returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


