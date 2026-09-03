# PasswordsSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToDelete** | Pointer to **[]string** | Passwords that need to be deleted from the cluster. | [optional] 
**ToAdd** | Pointer to [**[]PasswordToAddOrUpdate**](PasswordToAddOrUpdate.md) | Passwords that needs to be added to the cluster. | [optional] 
**ToUpdate** | Pointer to [**[]PasswordToAddOrUpdate**](PasswordToAddOrUpdate.md) | Passwords that needs to be updated at the cluster. | [optional] 

## Methods

### NewPasswordsSyncResponse

`func NewPasswordsSyncResponse() *PasswordsSyncResponse`

NewPasswordsSyncResponse instantiates a new PasswordsSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordsSyncResponseWithDefaults

`func NewPasswordsSyncResponseWithDefaults() *PasswordsSyncResponse`

NewPasswordsSyncResponseWithDefaults instantiates a new PasswordsSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToDelete

`func (o *PasswordsSyncResponse) GetToDelete() []string`

GetToDelete returns the ToDelete field if non-nil, zero value otherwise.

### GetToDeleteOk

`func (o *PasswordsSyncResponse) GetToDeleteOk() (*[]string, bool)`

GetToDeleteOk returns a tuple with the ToDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToDelete

`func (o *PasswordsSyncResponse) SetToDelete(v []string)`

SetToDelete sets ToDelete field to given value.

### HasToDelete

`func (o *PasswordsSyncResponse) HasToDelete() bool`

HasToDelete returns a boolean if a field has been set.

### GetToAdd

`func (o *PasswordsSyncResponse) GetToAdd() []PasswordToAddOrUpdate`

GetToAdd returns the ToAdd field if non-nil, zero value otherwise.

### GetToAddOk

`func (o *PasswordsSyncResponse) GetToAddOk() (*[]PasswordToAddOrUpdate, bool)`

GetToAddOk returns a tuple with the ToAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToAdd

`func (o *PasswordsSyncResponse) SetToAdd(v []PasswordToAddOrUpdate)`

SetToAdd sets ToAdd field to given value.

### HasToAdd

`func (o *PasswordsSyncResponse) HasToAdd() bool`

HasToAdd returns a boolean if a field has been set.

### GetToUpdate

`func (o *PasswordsSyncResponse) GetToUpdate() []PasswordToAddOrUpdate`

GetToUpdate returns the ToUpdate field if non-nil, zero value otherwise.

### GetToUpdateOk

`func (o *PasswordsSyncResponse) GetToUpdateOk() (*[]PasswordToAddOrUpdate, bool)`

GetToUpdateOk returns a tuple with the ToUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToUpdate

`func (o *PasswordsSyncResponse) SetToUpdate(v []PasswordToAddOrUpdate)`

SetToUpdate sets ToUpdate field to given value.

### HasToUpdate

`func (o *PasswordsSyncResponse) HasToUpdate() bool`

HasToUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


