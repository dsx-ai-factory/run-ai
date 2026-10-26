# GetProjectsMetadata200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Projects** | [**[]ProjectMetadata**](ProjectMetadata.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetProjectsMetadata200Response

`func NewGetProjectsMetadata200Response(projects []ProjectMetadata, ) *GetProjectsMetadata200Response`

NewGetProjectsMetadata200Response instantiates a new GetProjectsMetadata200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetProjectsMetadata200ResponseWithDefaults

`func NewGetProjectsMetadata200ResponseWithDefaults() *GetProjectsMetadata200Response`

NewGetProjectsMetadata200ResponseWithDefaults instantiates a new GetProjectsMetadata200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjects

`func (o *GetProjectsMetadata200Response) GetProjects() []ProjectMetadata`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *GetProjectsMetadata200Response) GetProjectsOk() (*[]ProjectMetadata, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *GetProjectsMetadata200Response) SetProjects(v []ProjectMetadata)`

SetProjects sets Projects field to given value.


### GetNext

`func (o *GetProjectsMetadata200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetProjectsMetadata200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetProjectsMetadata200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetProjectsMetadata200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


