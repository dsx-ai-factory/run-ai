# GenericSecretSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToDelete** | Pointer to **[]string** | Generic secrets that needs to be deleted from the cluster. | [optional] 
**ToAdd** | Pointer to [**[]GenericSecretToAddOrUpdate**](GenericSecretToAddOrUpdate.md) | Generic secrets that needs to be added to the cluster. | [optional] 
**ToUpdate** | Pointer to [**[]GenericSecretToAddOrUpdate**](GenericSecretToAddOrUpdate.md) | Generic secrets that needs to be updated at the cluster. | [optional] 

## Methods

### NewGenericSecretSyncResponse

`func NewGenericSecretSyncResponse() *GenericSecretSyncResponse`

NewGenericSecretSyncResponse instantiates a new GenericSecretSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretSyncResponseWithDefaults

`func NewGenericSecretSyncResponseWithDefaults() *GenericSecretSyncResponse`

NewGenericSecretSyncResponseWithDefaults instantiates a new GenericSecretSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToDelete

`func (o *GenericSecretSyncResponse) GetToDelete() []string`

GetToDelete returns the ToDelete field if non-nil, zero value otherwise.

### GetToDeleteOk

`func (o *GenericSecretSyncResponse) GetToDeleteOk() (*[]string, bool)`

GetToDeleteOk returns a tuple with the ToDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToDelete

`func (o *GenericSecretSyncResponse) SetToDelete(v []string)`

SetToDelete sets ToDelete field to given value.

### HasToDelete

`func (o *GenericSecretSyncResponse) HasToDelete() bool`

HasToDelete returns a boolean if a field has been set.

### GetToAdd

`func (o *GenericSecretSyncResponse) GetToAdd() []GenericSecretToAddOrUpdate`

GetToAdd returns the ToAdd field if non-nil, zero value otherwise.

### GetToAddOk

`func (o *GenericSecretSyncResponse) GetToAddOk() (*[]GenericSecretToAddOrUpdate, bool)`

GetToAddOk returns a tuple with the ToAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToAdd

`func (o *GenericSecretSyncResponse) SetToAdd(v []GenericSecretToAddOrUpdate)`

SetToAdd sets ToAdd field to given value.

### HasToAdd

`func (o *GenericSecretSyncResponse) HasToAdd() bool`

HasToAdd returns a boolean if a field has been set.

### GetToUpdate

`func (o *GenericSecretSyncResponse) GetToUpdate() []GenericSecretToAddOrUpdate`

GetToUpdate returns the ToUpdate field if non-nil, zero value otherwise.

### GetToUpdateOk

`func (o *GenericSecretSyncResponse) GetToUpdateOk() (*[]GenericSecretToAddOrUpdate, bool)`

GetToUpdateOk returns a tuple with the ToUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToUpdate

`func (o *GenericSecretSyncResponse) SetToUpdate(v []GenericSecretToAddOrUpdate)`

SetToUpdate sets ToUpdate field to given value.

### HasToUpdate

`func (o *GenericSecretSyncResponse) HasToUpdate() bool`

HasToUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


