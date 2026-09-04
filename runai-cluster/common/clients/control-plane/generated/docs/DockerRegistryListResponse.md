# DockerRegistryListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]DockerRegistry**](DockerRegistry.md) |  | 

## Methods

### NewDockerRegistryListResponse

`func NewDockerRegistryListResponse(entries []DockerRegistry, ) *DockerRegistryListResponse`

NewDockerRegistryListResponse instantiates a new DockerRegistryListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistryListResponseWithDefaults

`func NewDockerRegistryListResponseWithDefaults() *DockerRegistryListResponse`

NewDockerRegistryListResponseWithDefaults instantiates a new DockerRegistryListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *DockerRegistryListResponse) GetEntries() []DockerRegistry`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *DockerRegistryListResponse) GetEntriesOk() (*[]DockerRegistry, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *DockerRegistryListResponse) SetEntries(v []DockerRegistry)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


