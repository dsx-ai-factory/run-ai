# PVCsSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToDelete** | Pointer to **[]string** | PVCs that need to be deleted from the cluster. | [optional] 
**ToAdd** | Pointer to [**[]PVCToAddOrUpdate**](PVCToAddOrUpdate.md) | PVCs that needs to be added to the cluster. | [optional] 
**ToUpdate** | Pointer to [**[]PVCToAddOrUpdate**](PVCToAddOrUpdate.md) | PVCs that needs to be updated at the cluster. | [optional] 

## Methods

### NewPVCsSyncResponse

`func NewPVCsSyncResponse() *PVCsSyncResponse`

NewPVCsSyncResponse instantiates a new PVCsSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCsSyncResponseWithDefaults

`func NewPVCsSyncResponseWithDefaults() *PVCsSyncResponse`

NewPVCsSyncResponseWithDefaults instantiates a new PVCsSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToDelete

`func (o *PVCsSyncResponse) GetToDelete() []string`

GetToDelete returns the ToDelete field if non-nil, zero value otherwise.

### GetToDeleteOk

`func (o *PVCsSyncResponse) GetToDeleteOk() (*[]string, bool)`

GetToDeleteOk returns a tuple with the ToDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToDelete

`func (o *PVCsSyncResponse) SetToDelete(v []string)`

SetToDelete sets ToDelete field to given value.

### HasToDelete

`func (o *PVCsSyncResponse) HasToDelete() bool`

HasToDelete returns a boolean if a field has been set.

### GetToAdd

`func (o *PVCsSyncResponse) GetToAdd() []PVCToAddOrUpdate`

GetToAdd returns the ToAdd field if non-nil, zero value otherwise.

### GetToAddOk

`func (o *PVCsSyncResponse) GetToAddOk() (*[]PVCToAddOrUpdate, bool)`

GetToAddOk returns a tuple with the ToAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToAdd

`func (o *PVCsSyncResponse) SetToAdd(v []PVCToAddOrUpdate)`

SetToAdd sets ToAdd field to given value.

### HasToAdd

`func (o *PVCsSyncResponse) HasToAdd() bool`

HasToAdd returns a boolean if a field has been set.

### GetToUpdate

`func (o *PVCsSyncResponse) GetToUpdate() []PVCToAddOrUpdate`

GetToUpdate returns the ToUpdate field if non-nil, zero value otherwise.

### GetToUpdateOk

`func (o *PVCsSyncResponse) GetToUpdateOk() (*[]PVCToAddOrUpdate, bool)`

GetToUpdateOk returns a tuple with the ToUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToUpdate

`func (o *PVCsSyncResponse) SetToUpdate(v []PVCToAddOrUpdate)`

SetToUpdate sets ToUpdate field to given value.

### HasToUpdate

`func (o *PVCsSyncResponse) HasToUpdate() bool`

HasToUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


