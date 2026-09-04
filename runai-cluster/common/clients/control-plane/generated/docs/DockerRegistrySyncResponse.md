# DockerRegistrySyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToDelete** | Pointer to **[]string** | Docker registries that needs to be deleted from the cluster. | [optional] 
**ToAdd** | Pointer to [**[]DockerRegistryToAddOrUpdate**](DockerRegistryToAddOrUpdate.md) | Docker registries that needs to be added to the cluster. | [optional] 
**ToUpdate** | Pointer to [**[]DockerRegistryToAddOrUpdate**](DockerRegistryToAddOrUpdate.md) | Docker registries that needs to be updated at the cluster. | [optional] 

## Methods

### NewDockerRegistrySyncResponse

`func NewDockerRegistrySyncResponse() *DockerRegistrySyncResponse`

NewDockerRegistrySyncResponse instantiates a new DockerRegistrySyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistrySyncResponseWithDefaults

`func NewDockerRegistrySyncResponseWithDefaults() *DockerRegistrySyncResponse`

NewDockerRegistrySyncResponseWithDefaults instantiates a new DockerRegistrySyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToDelete

`func (o *DockerRegistrySyncResponse) GetToDelete() []string`

GetToDelete returns the ToDelete field if non-nil, zero value otherwise.

### GetToDeleteOk

`func (o *DockerRegistrySyncResponse) GetToDeleteOk() (*[]string, bool)`

GetToDeleteOk returns a tuple with the ToDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToDelete

`func (o *DockerRegistrySyncResponse) SetToDelete(v []string)`

SetToDelete sets ToDelete field to given value.

### HasToDelete

`func (o *DockerRegistrySyncResponse) HasToDelete() bool`

HasToDelete returns a boolean if a field has been set.

### GetToAdd

`func (o *DockerRegistrySyncResponse) GetToAdd() []DockerRegistryToAddOrUpdate`

GetToAdd returns the ToAdd field if non-nil, zero value otherwise.

### GetToAddOk

`func (o *DockerRegistrySyncResponse) GetToAddOk() (*[]DockerRegistryToAddOrUpdate, bool)`

GetToAddOk returns a tuple with the ToAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToAdd

`func (o *DockerRegistrySyncResponse) SetToAdd(v []DockerRegistryToAddOrUpdate)`

SetToAdd sets ToAdd field to given value.

### HasToAdd

`func (o *DockerRegistrySyncResponse) HasToAdd() bool`

HasToAdd returns a boolean if a field has been set.

### GetToUpdate

`func (o *DockerRegistrySyncResponse) GetToUpdate() []DockerRegistryToAddOrUpdate`

GetToUpdate returns the ToUpdate field if non-nil, zero value otherwise.

### GetToUpdateOk

`func (o *DockerRegistrySyncResponse) GetToUpdateOk() (*[]DockerRegistryToAddOrUpdate, bool)`

GetToUpdateOk returns a tuple with the ToUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToUpdate

`func (o *DockerRegistrySyncResponse) SetToUpdate(v []DockerRegistryToAddOrUpdate)`

SetToUpdate sets ToUpdate field to given value.

### HasToUpdate

`func (o *DockerRegistrySyncResponse) HasToUpdate() bool`

HasToUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


