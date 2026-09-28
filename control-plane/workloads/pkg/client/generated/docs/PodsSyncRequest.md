# PodsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectId** | **string** |  | 
**ClusterId** | Pointer to **string** |  | [optional] 
**Pods** | [**[]PodSyncFields**](PodSyncFields.md) |  | 
**From** | Pointer to **string** | the first pod id to sync. if empty, all pods before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last pod to sync. if empty, all pods after the last one sent will be deleted. | [optional] 
**IdsToIgnore** | Pointer to **[]string** | pod ids to ignore when deleting. | [optional] 

## Methods

### NewPodsSyncRequest

`func NewPodsSyncRequest(projectId string, pods []PodSyncFields, ) *PodsSyncRequest`

NewPodsSyncRequest instantiates a new PodsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodsSyncRequestWithDefaults

`func NewPodsSyncRequestWithDefaults() *PodsSyncRequest`

NewPodsSyncRequestWithDefaults instantiates a new PodsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectId

`func (o *PodsSyncRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *PodsSyncRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *PodsSyncRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *PodsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PodsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PodsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *PodsSyncRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetPods

`func (o *PodsSyncRequest) GetPods() []PodSyncFields`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *PodsSyncRequest) GetPodsOk() (*[]PodSyncFields, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *PodsSyncRequest) SetPods(v []PodSyncFields)`

SetPods sets Pods field to given value.


### GetFrom

`func (o *PodsSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *PodsSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *PodsSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *PodsSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *PodsSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *PodsSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *PodsSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *PodsSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.

### GetIdsToIgnore

`func (o *PodsSyncRequest) GetIdsToIgnore() []string`

GetIdsToIgnore returns the IdsToIgnore field if non-nil, zero value otherwise.

### GetIdsToIgnoreOk

`func (o *PodsSyncRequest) GetIdsToIgnoreOk() (*[]string, bool)`

GetIdsToIgnoreOk returns a tuple with the IdsToIgnore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdsToIgnore

`func (o *PodsSyncRequest) SetIdsToIgnore(v []string)`

SetIdsToIgnore sets IdsToIgnore field to given value.

### HasIdsToIgnore

`func (o *PodsSyncRequest) HasIdsToIgnore() bool`

HasIdsToIgnore returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


