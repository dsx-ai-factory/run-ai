# AssetFlatResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AssetsMeta** | Pointer to [**[]AssetMeta**](AssetMeta.md) |  | [optional] 
**Worker** | Pointer to [**SupersetSpec**](SupersetSpec.md) |  | [optional] 
**Master** | Pointer to [**NullableMasterSpec**](MasterSpec.md) |  | [optional] 

## Methods

### NewAssetFlatResponse

`func NewAssetFlatResponse() *AssetFlatResponse`

NewAssetFlatResponse instantiates a new AssetFlatResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetFlatResponseWithDefaults

`func NewAssetFlatResponseWithDefaults() *AssetFlatResponse`

NewAssetFlatResponseWithDefaults instantiates a new AssetFlatResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssetsMeta

`func (o *AssetFlatResponse) GetAssetsMeta() []AssetMeta`

GetAssetsMeta returns the AssetsMeta field if non-nil, zero value otherwise.

### GetAssetsMetaOk

`func (o *AssetFlatResponse) GetAssetsMetaOk() (*[]AssetMeta, bool)`

GetAssetsMetaOk returns a tuple with the AssetsMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetsMeta

`func (o *AssetFlatResponse) SetAssetsMeta(v []AssetMeta)`

SetAssetsMeta sets AssetsMeta field to given value.

### HasAssetsMeta

`func (o *AssetFlatResponse) HasAssetsMeta() bool`

HasAssetsMeta returns a boolean if a field has been set.

### GetWorker

`func (o *AssetFlatResponse) GetWorker() SupersetSpec`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *AssetFlatResponse) GetWorkerOk() (*SupersetSpec, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *AssetFlatResponse) SetWorker(v SupersetSpec)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *AssetFlatResponse) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### GetMaster

`func (o *AssetFlatResponse) GetMaster() MasterSpec`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *AssetFlatResponse) GetMasterOk() (*MasterSpec, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *AssetFlatResponse) SetMaster(v MasterSpec)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *AssetFlatResponse) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *AssetFlatResponse) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *AssetFlatResponse) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


