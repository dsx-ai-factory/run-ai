# CredentialsListResponseEntry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**CredentialsListResponseAssetSpec**](CredentialsListResponseAssetSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 

## Methods

### NewCredentialsListResponseEntry

`func NewCredentialsListResponseEntry(meta AssetMeta, spec CredentialsListResponseAssetSpec, ) *CredentialsListResponseEntry`

NewCredentialsListResponseEntry instantiates a new CredentialsListResponseEntry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsListResponseEntryWithDefaults

`func NewCredentialsListResponseEntryWithDefaults() *CredentialsListResponseEntry`

NewCredentialsListResponseEntryWithDefaults instantiates a new CredentialsListResponseEntry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *CredentialsListResponseEntry) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *CredentialsListResponseEntry) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *CredentialsListResponseEntry) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *CredentialsListResponseEntry) GetSpec() CredentialsListResponseAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *CredentialsListResponseEntry) GetSpecOk() (*CredentialsListResponseAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *CredentialsListResponseEntry) SetSpec(v CredentialsListResponseAssetSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *CredentialsListResponseEntry) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *CredentialsListResponseEntry) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *CredentialsListResponseEntry) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *CredentialsListResponseEntry) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *CredentialsListResponseEntry) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *CredentialsListResponseEntry) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


