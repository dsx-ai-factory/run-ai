# GetSyncProjects200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Projects** | [**[]ProjectSync**](ProjectSync.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetSyncProjects200Response

`func NewGetSyncProjects200Response(projects []ProjectSync, ) *GetSyncProjects200Response`

NewGetSyncProjects200Response instantiates a new GetSyncProjects200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetSyncProjects200ResponseWithDefaults

`func NewGetSyncProjects200ResponseWithDefaults() *GetSyncProjects200Response`

NewGetSyncProjects200ResponseWithDefaults instantiates a new GetSyncProjects200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjects

`func (o *GetSyncProjects200Response) GetProjects() []ProjectSync`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *GetSyncProjects200Response) GetProjectsOk() (*[]ProjectSync, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *GetSyncProjects200Response) SetProjects(v []ProjectSync)`

SetProjects sets Projects field to given value.


### GetNext

`func (o *GetSyncProjects200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetSyncProjects200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetSyncProjects200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetSyncProjects200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


