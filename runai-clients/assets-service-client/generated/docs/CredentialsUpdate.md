# CredentialsUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMetaForSync**](AssetMetaForSync.md) |  | 
**UpdateCount** | **int64** | the value of an update counter of a data in the system. | 
**IsDeleted** | Pointer to **NullableBool** | the asset has been deleted | [optional] 
**Spec** | Pointer to [**CredentialsUpdateSpec**](CredentialsUpdateSpec.md) |  | [optional] 

## Methods

### NewCredentialsUpdate

`func NewCredentialsUpdate(meta AssetMetaForSync, updateCount int64, ) *CredentialsUpdate`

NewCredentialsUpdate instantiates a new CredentialsUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsUpdateWithDefaults

`func NewCredentialsUpdateWithDefaults() *CredentialsUpdate`

NewCredentialsUpdateWithDefaults instantiates a new CredentialsUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *CredentialsUpdate) GetMeta() AssetMetaForSync`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *CredentialsUpdate) GetMetaOk() (*AssetMetaForSync, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *CredentialsUpdate) SetMeta(v AssetMetaForSync)`

SetMeta sets Meta field to given value.


### GetUpdateCount

`func (o *CredentialsUpdate) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *CredentialsUpdate) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *CredentialsUpdate) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.


### GetIsDeleted

`func (o *CredentialsUpdate) GetIsDeleted() bool`

GetIsDeleted returns the IsDeleted field if non-nil, zero value otherwise.

### GetIsDeletedOk

`func (o *CredentialsUpdate) GetIsDeletedOk() (*bool, bool)`

GetIsDeletedOk returns a tuple with the IsDeleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeleted

`func (o *CredentialsUpdate) SetIsDeleted(v bool)`

SetIsDeleted sets IsDeleted field to given value.

### HasIsDeleted

`func (o *CredentialsUpdate) HasIsDeleted() bool`

HasIsDeleted returns a boolean if a field has been set.

### SetIsDeletedNil

`func (o *CredentialsUpdate) SetIsDeletedNil(b bool)`

 SetIsDeletedNil sets the value for IsDeleted to be an explicit nil

### UnsetIsDeleted
`func (o *CredentialsUpdate) UnsetIsDeleted()`

UnsetIsDeleted ensures that no value is present for IsDeleted, not even an explicit nil
### GetSpec

`func (o *CredentialsUpdate) GetSpec() CredentialsUpdateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *CredentialsUpdate) GetSpecOk() (*CredentialsUpdateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *CredentialsUpdate) SetSpec(v CredentialsUpdateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *CredentialsUpdate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


