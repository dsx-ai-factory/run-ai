# TenantUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | Pointer to **NullableString** | An optional alias for the tenant. Must be unique across all tenant names and aliases. | [optional] 

## Methods

### NewTenantUpdateRequest

`func NewTenantUpdateRequest() *TenantUpdateRequest`

NewTenantUpdateRequest instantiates a new TenantUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantUpdateRequestWithDefaults

`func NewTenantUpdateRequestWithDefaults() *TenantUpdateRequest`

NewTenantUpdateRequestWithDefaults instantiates a new TenantUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *TenantUpdateRequest) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *TenantUpdateRequest) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *TenantUpdateRequest) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *TenantUpdateRequest) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### SetAliasNil

`func (o *TenantUpdateRequest) SetAliasNil(b bool)`

 SetAliasNil sets the value for Alias to be an explicit nil

### UnsetAlias
`func (o *TenantUpdateRequest) UnsetAlias()`

UnsetAlias ensures that no value is present for Alias, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


