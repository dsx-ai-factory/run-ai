# AssetIdAndKind

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 
**Overrides** | Pointer to [**NullableDataSourceOverrides**](DataSourceOverrides.md) |  | [optional] 

## Methods

### NewAssetIdAndKind

`func NewAssetIdAndKind(id string, kind AssetKind, ) *AssetIdAndKind`

NewAssetIdAndKind instantiates a new AssetIdAndKind object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetIdAndKindWithDefaults

`func NewAssetIdAndKindWithDefaults() *AssetIdAndKind`

NewAssetIdAndKindWithDefaults instantiates a new AssetIdAndKind object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssetIdAndKind) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetIdAndKind) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetIdAndKind) SetId(v string)`

SetId sets Id field to given value.


### GetKind

`func (o *AssetIdAndKind) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetIdAndKind) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetIdAndKind) SetKind(v AssetKind)`

SetKind sets Kind field to given value.


### GetOverrides

`func (o *AssetIdAndKind) GetOverrides() DataSourceOverrides`

GetOverrides returns the Overrides field if non-nil, zero value otherwise.

### GetOverridesOk

`func (o *AssetIdAndKind) GetOverridesOk() (*DataSourceOverrides, bool)`

GetOverridesOk returns a tuple with the Overrides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverrides

`func (o *AssetIdAndKind) SetOverrides(v DataSourceOverrides)`

SetOverrides sets Overrides field to given value.

### HasOverrides

`func (o *AssetIdAndKind) HasOverrides() bool`

HasOverrides returns a boolean if a field has been set.

### SetOverridesNil

`func (o *AssetIdAndKind) SetOverridesNil(b bool)`

 SetOverridesNil sets the value for Overrides to be an explicit nil

### UnsetOverrides
`func (o *AssetIdAndKind) UnsetOverrides()`

UnsetOverrides ensures that no value is present for Overrides, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


