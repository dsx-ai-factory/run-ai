# TenantSettingCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** |  | [optional] 
**Value** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewTenantSettingCreationRequest

`func NewTenantSettingCreationRequest() *TenantSettingCreationRequest`

NewTenantSettingCreationRequest instantiates a new TenantSettingCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantSettingCreationRequestWithDefaults

`func NewTenantSettingCreationRequestWithDefaults() *TenantSettingCreationRequest`

NewTenantSettingCreationRequestWithDefaults instantiates a new TenantSettingCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *TenantSettingCreationRequest) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *TenantSettingCreationRequest) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *TenantSettingCreationRequest) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *TenantSettingCreationRequest) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetValue

`func (o *TenantSettingCreationRequest) GetValue() interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *TenantSettingCreationRequest) GetValueOk() (*interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *TenantSettingCreationRequest) SetValue(v interface{})`

SetValue sets Value field to given value.

### HasValue

`func (o *TenantSettingCreationRequest) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *TenantSettingCreationRequest) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *TenantSettingCreationRequest) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


