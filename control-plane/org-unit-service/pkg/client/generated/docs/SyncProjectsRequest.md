# SyncProjectsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** |  | 
**Projects** | [**[]ProjectInSyncRequest**](ProjectInSyncRequest.md) |  | 

## Methods

### NewSyncProjectsRequest

`func NewSyncProjectsRequest(clusterId string, projects []ProjectInSyncRequest, ) *SyncProjectsRequest`

NewSyncProjectsRequest instantiates a new SyncProjectsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncProjectsRequestWithDefaults

`func NewSyncProjectsRequestWithDefaults() *SyncProjectsRequest`

NewSyncProjectsRequestWithDefaults instantiates a new SyncProjectsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SyncProjectsRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SyncProjectsRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SyncProjectsRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjects

`func (o *SyncProjectsRequest) GetProjects() []ProjectInSyncRequest`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *SyncProjectsRequest) GetProjectsOk() (*[]ProjectInSyncRequest, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *SyncProjectsRequest) SetProjects(v []ProjectInSyncRequest)`

SetProjects sets Projects field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


