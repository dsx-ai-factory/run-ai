# TenantSettingValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | Pointer to **int32** | The tenant ID | [optional] 
**Value** | Pointer to [**TenantSettingValueValue**](TenantSettingValueValue.md) |  | [optional] 

## Methods

### NewTenantSettingValue

`func NewTenantSettingValue() *TenantSettingValue`

NewTenantSettingValue instantiates a new TenantSettingValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantSettingValueWithDefaults

`func NewTenantSettingValueWithDefaults() *TenantSettingValue`

NewTenantSettingValueWithDefaults instantiates a new TenantSettingValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *TenantSettingValue) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *TenantSettingValue) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *TenantSettingValue) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *TenantSettingValue) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetValue

`func (o *TenantSettingValue) GetValue() TenantSettingValueValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *TenantSettingValue) GetValueOk() (*TenantSettingValueValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *TenantSettingValue) SetValue(v TenantSettingValueValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *TenantSettingValue) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


