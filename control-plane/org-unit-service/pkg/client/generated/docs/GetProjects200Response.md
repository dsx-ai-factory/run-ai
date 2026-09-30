# GetProjects200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Projects** | [**[]Project**](Project.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetProjects200Response

`func NewGetProjects200Response(projects []Project, ) *GetProjects200Response`

NewGetProjects200Response instantiates a new GetProjects200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetProjects200ResponseWithDefaults

`func NewGetProjects200ResponseWithDefaults() *GetProjects200Response`

NewGetProjects200ResponseWithDefaults instantiates a new GetProjects200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjects

`func (o *GetProjects200Response) GetProjects() []Project`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *GetProjects200Response) GetProjectsOk() (*[]Project, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *GetProjects200Response) SetProjects(v []Project)`

SetProjects sets Projects field to given value.


### GetNext

`func (o *GetProjects200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetProjects200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetProjects200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetProjects200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


