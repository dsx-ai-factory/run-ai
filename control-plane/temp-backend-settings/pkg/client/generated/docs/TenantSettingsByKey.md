# TenantSettingsByKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** |  | [optional] 
**Values** | Pointer to [**[]TenantSettingValue**](TenantSettingValue.md) |  | [optional] 

## Methods

### NewTenantSettingsByKey

`func NewTenantSettingsByKey() *TenantSettingsByKey`

NewTenantSettingsByKey instantiates a new TenantSettingsByKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantSettingsByKeyWithDefaults

`func NewTenantSettingsByKeyWithDefaults() *TenantSettingsByKey`

NewTenantSettingsByKeyWithDefaults instantiates a new TenantSettingsByKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *TenantSettingsByKey) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *TenantSettingsByKey) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *TenantSettingsByKey) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *TenantSettingsByKey) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetValues

`func (o *TenantSettingsByKey) GetValues() []TenantSettingValue`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *TenantSettingsByKey) GetValuesOk() (*[]TenantSettingValue, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *TenantSettingsByKey) SetValues(v []TenantSettingValue)`

SetValues sets Values field to given value.

### HasValues

`func (o *TenantSettingsByKey) HasValues() bool`

HasValues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


