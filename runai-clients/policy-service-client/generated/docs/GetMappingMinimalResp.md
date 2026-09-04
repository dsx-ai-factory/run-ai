# GetMappingMinimalResp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fields** | Pointer to [**map[string]map[string]map[string]string**](map.md) | Maps each field name to its spec selectors. | [optional] 
**SpecSelectors** | Pointer to **map[string]string** | Maps each spec selector name to its description. | [optional] 

## Methods

### NewGetMappingMinimalResp

`func NewGetMappingMinimalResp() *GetMappingMinimalResp`

NewGetMappingMinimalResp instantiates a new GetMappingMinimalResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetMappingMinimalRespWithDefaults

`func NewGetMappingMinimalRespWithDefaults() *GetMappingMinimalResp`

NewGetMappingMinimalRespWithDefaults instantiates a new GetMappingMinimalResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFields

`func (o *GetMappingMinimalResp) GetFields() map[string]map[string]map[string]string`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *GetMappingMinimalResp) GetFieldsOk() (*map[string]map[string]map[string]string, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *GetMappingMinimalResp) SetFields(v map[string]map[string]map[string]string)`

SetFields sets Fields field to given value.

### HasFields

`func (o *GetMappingMinimalResp) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetSpecSelectors

`func (o *GetMappingMinimalResp) GetSpecSelectors() map[string]string`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *GetMappingMinimalResp) GetSpecSelectorsOk() (*map[string]string, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *GetMappingMinimalResp) SetSpecSelectors(v map[string]string)`

SetSpecSelectors sets SpecSelectors field to given value.

### HasSpecSelectors

`func (o *GetMappingMinimalResp) HasSpecSelectors() bool`

HasSpecSelectors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


