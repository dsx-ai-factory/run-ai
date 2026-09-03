# DatasourceUpdateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Pvc** | Pointer to [**NullablePVCAssetSpec**](PVCAssetSpec.md) |  | [optional] 

## Methods

### NewDatasourceUpdateSpec

`func NewDatasourceUpdateSpec() *DatasourceUpdateSpec`

NewDatasourceUpdateSpec instantiates a new DatasourceUpdateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceUpdateSpecWithDefaults

`func NewDatasourceUpdateSpecWithDefaults() *DatasourceUpdateSpec`

NewDatasourceUpdateSpecWithDefaults instantiates a new DatasourceUpdateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPvc

`func (o *DatasourceUpdateSpec) GetPvc() PVCAssetSpec`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DatasourceUpdateSpec) GetPvcOk() (*PVCAssetSpec, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DatasourceUpdateSpec) SetPvc(v PVCAssetSpec)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DatasourceUpdateSpec) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *DatasourceUpdateSpec) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *DatasourceUpdateSpec) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


