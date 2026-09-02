# AccessKeysSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToDelete** | Pointer to **[]string** | Access keys that needs to be deleted from the cluster. | [optional] 
**ToAdd** | Pointer to [**[]AccessKeyToAddOrUpdate**](AccessKeyToAddOrUpdate.md) | Access keys that needs to be added to the cluster. | [optional] 
**ToUpdate** | Pointer to [**[]AccessKeyToAddOrUpdate**](AccessKeyToAddOrUpdate.md) | Access keys that needs to be updated at the cluster. | [optional] 

## Methods

### NewAccessKeysSyncResponse

`func NewAccessKeysSyncResponse() *AccessKeysSyncResponse`

NewAccessKeysSyncResponse instantiates a new AccessKeysSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeysSyncResponseWithDefaults

`func NewAccessKeysSyncResponseWithDefaults() *AccessKeysSyncResponse`

NewAccessKeysSyncResponseWithDefaults instantiates a new AccessKeysSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToDelete

`func (o *AccessKeysSyncResponse) GetToDelete() []string`

GetToDelete returns the ToDelete field if non-nil, zero value otherwise.

### GetToDeleteOk

`func (o *AccessKeysSyncResponse) GetToDeleteOk() (*[]string, bool)`

GetToDeleteOk returns a tuple with the ToDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToDelete

`func (o *AccessKeysSyncResponse) SetToDelete(v []string)`

SetToDelete sets ToDelete field to given value.

### HasToDelete

`func (o *AccessKeysSyncResponse) HasToDelete() bool`

HasToDelete returns a boolean if a field has been set.

### GetToAdd

`func (o *AccessKeysSyncResponse) GetToAdd() []AccessKeyToAddOrUpdate`

GetToAdd returns the ToAdd field if non-nil, zero value otherwise.

### GetToAddOk

`func (o *AccessKeysSyncResponse) GetToAddOk() (*[]AccessKeyToAddOrUpdate, bool)`

GetToAddOk returns a tuple with the ToAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToAdd

`func (o *AccessKeysSyncResponse) SetToAdd(v []AccessKeyToAddOrUpdate)`

SetToAdd sets ToAdd field to given value.

### HasToAdd

`func (o *AccessKeysSyncResponse) HasToAdd() bool`

HasToAdd returns a boolean if a field has been set.

### GetToUpdate

`func (o *AccessKeysSyncResponse) GetToUpdate() []AccessKeyToAddOrUpdate`

GetToUpdate returns the ToUpdate field if non-nil, zero value otherwise.

### GetToUpdateOk

`func (o *AccessKeysSyncResponse) GetToUpdateOk() (*[]AccessKeyToAddOrUpdate, bool)`

GetToUpdateOk returns a tuple with the ToUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToUpdate

`func (o *AccessKeysSyncResponse) SetToUpdate(v []AccessKeyToAddOrUpdate)`

SetToUpdate sets ToUpdate field to given value.

### HasToUpdate

`func (o *AccessKeysSyncResponse) HasToUpdate() bool`

HasToUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


