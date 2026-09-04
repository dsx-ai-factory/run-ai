# AssetDatasourceRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Name** | **string** | The name of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 
**Overrides** | Pointer to [**NullableDataSourceOverrides**](DataSourceOverrides.md) |  | [optional] 

## Methods

### NewAssetDatasourceRef

`func NewAssetDatasourceRef(id string, name string, kind AssetKind, ) *AssetDatasourceRef`

NewAssetDatasourceRef instantiates a new AssetDatasourceRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetDatasourceRefWithDefaults

`func NewAssetDatasourceRefWithDefaults() *AssetDatasourceRef`

NewAssetDatasourceRefWithDefaults instantiates a new AssetDatasourceRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssetDatasourceRef) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetDatasourceRef) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetDatasourceRef) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *AssetDatasourceRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetDatasourceRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetDatasourceRef) SetName(v string)`

SetName sets Name field to given value.


### GetKind

`func (o *AssetDatasourceRef) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetDatasourceRef) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetDatasourceRef) SetKind(v AssetKind)`

SetKind sets Kind field to given value.


### GetOverrides

`func (o *AssetDatasourceRef) GetOverrides() DataSourceOverrides`

GetOverrides returns the Overrides field if non-nil, zero value otherwise.

### GetOverridesOk

`func (o *AssetDatasourceRef) GetOverridesOk() (*DataSourceOverrides, bool)`

GetOverridesOk returns a tuple with the Overrides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverrides

`func (o *AssetDatasourceRef) SetOverrides(v DataSourceOverrides)`

SetOverrides sets Overrides field to given value.

### HasOverrides

`func (o *AssetDatasourceRef) HasOverrides() bool`

HasOverrides returns a boolean if a field has been set.

### SetOverridesNil

`func (o *AssetDatasourceRef) SetOverridesNil(b bool)`

 SetOverridesNil sets the value for Overrides to be an explicit nil

### UnsetOverrides
`func (o *AssetDatasourceRef) UnsetOverrides()`

UnsetOverrides ensures that no value is present for Overrides, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


