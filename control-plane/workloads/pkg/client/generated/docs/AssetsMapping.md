# AssetsMapping

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The name of the asset. | [optional] 
**Kind** | Pointer to [**AssetKind**](AssetKind.md) |  | [optional] 
**Id** | Pointer to **string** | Unique identifier of the asset. | [optional] 

## Methods

### NewAssetsMapping

`func NewAssetsMapping() *AssetsMapping`

NewAssetsMapping instantiates a new AssetsMapping object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsMappingWithDefaults

`func NewAssetsMappingWithDefaults() *AssetsMapping`

NewAssetsMappingWithDefaults instantiates a new AssetsMapping object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AssetsMapping) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetsMapping) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetsMapping) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AssetsMapping) HasName() bool`

HasName returns a boolean if a field has been set.

### GetKind

`func (o *AssetsMapping) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetsMapping) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetsMapping) SetKind(v AssetKind)`

SetKind sets Kind field to given value.

### HasKind

`func (o *AssetsMapping) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetId

`func (o *AssetsMapping) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetsMapping) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetsMapping) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AssetsMapping) HasId() bool`

HasId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


