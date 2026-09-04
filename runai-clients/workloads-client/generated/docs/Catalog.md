# Catalog

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsStaticData** | **bool** |  | 
**ResultTotal** | Pointer to **int32** |  | [optional] 
**ResultPageTotal** | Pointer to **int32** |  | [optional] 
**Params** | Pointer to [**CatalogParams**](CatalogParams.md) |  | [optional] 
**Resources** | Pointer to [**[]CatalogResourcesInner**](CatalogResourcesInner.md) |  | [optional] 

## Methods

### NewCatalog

`func NewCatalog(isStaticData bool, ) *Catalog`

NewCatalog instantiates a new Catalog object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCatalogWithDefaults

`func NewCatalogWithDefaults() *Catalog`

NewCatalogWithDefaults instantiates a new Catalog object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsStaticData

`func (o *Catalog) GetIsStaticData() bool`

GetIsStaticData returns the IsStaticData field if non-nil, zero value otherwise.

### GetIsStaticDataOk

`func (o *Catalog) GetIsStaticDataOk() (*bool, bool)`

GetIsStaticDataOk returns a tuple with the IsStaticData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsStaticData

`func (o *Catalog) SetIsStaticData(v bool)`

SetIsStaticData sets IsStaticData field to given value.


### GetResultTotal

`func (o *Catalog) GetResultTotal() int32`

GetResultTotal returns the ResultTotal field if non-nil, zero value otherwise.

### GetResultTotalOk

`func (o *Catalog) GetResultTotalOk() (*int32, bool)`

GetResultTotalOk returns a tuple with the ResultTotal field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResultTotal

`func (o *Catalog) SetResultTotal(v int32)`

SetResultTotal sets ResultTotal field to given value.

### HasResultTotal

`func (o *Catalog) HasResultTotal() bool`

HasResultTotal returns a boolean if a field has been set.

### GetResultPageTotal

`func (o *Catalog) GetResultPageTotal() int32`

GetResultPageTotal returns the ResultPageTotal field if non-nil, zero value otherwise.

### GetResultPageTotalOk

`func (o *Catalog) GetResultPageTotalOk() (*int32, bool)`

GetResultPageTotalOk returns a tuple with the ResultPageTotal field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResultPageTotal

`func (o *Catalog) SetResultPageTotal(v int32)`

SetResultPageTotal sets ResultPageTotal field to given value.

### HasResultPageTotal

`func (o *Catalog) HasResultPageTotal() bool`

HasResultPageTotal returns a boolean if a field has been set.

### GetParams

`func (o *Catalog) GetParams() CatalogParams`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *Catalog) GetParamsOk() (*CatalogParams, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *Catalog) SetParams(v CatalogParams)`

SetParams sets Params field to given value.

### HasParams

`func (o *Catalog) HasParams() bool`

HasParams returns a boolean if a field has been set.

### GetResources

`func (o *Catalog) GetResources() []CatalogResourcesInner`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Catalog) GetResourcesOk() (*[]CatalogResourcesInner, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Catalog) SetResources(v []CatalogResourcesInner)`

SetResources sets Resources field to given value.

### HasResources

`func (o *Catalog) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


