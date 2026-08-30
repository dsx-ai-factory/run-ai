# ComputeFlatFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CpuCoreRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuCoreLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**CpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**LargeShmRequest** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**GpuRequestType** | Pointer to [**NullableGpuRequestRules**](GpuRequestRules.md) |  | [optional] 
**MigProfile** | Pointer to [**NullableMigProfileRules**](MigProfileRules.md) |  | [optional] 
**GpuDevicesRequest** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**GpuPortionRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuPortionLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 

## Methods

### NewComputeFlatFieldsRules

`func NewComputeFlatFieldsRules() *ComputeFlatFieldsRules`

NewComputeFlatFieldsRules instantiates a new ComputeFlatFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputeFlatFieldsRulesWithDefaults

`func NewComputeFlatFieldsRulesWithDefaults() *ComputeFlatFieldsRules`

NewComputeFlatFieldsRulesWithDefaults instantiates a new ComputeFlatFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreRequest

`func (o *ComputeFlatFieldsRules) GetCpuCoreRequest() NumberRules`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *ComputeFlatFieldsRules) GetCpuCoreRequestOk() (*NumberRules, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *ComputeFlatFieldsRules) SetCpuCoreRequest(v NumberRules)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *ComputeFlatFieldsRules) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *ComputeFlatFieldsRules) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *ComputeFlatFieldsRules) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuCoreLimit

`func (o *ComputeFlatFieldsRules) GetCpuCoreLimit() NumberRules`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *ComputeFlatFieldsRules) GetCpuCoreLimitOk() (*NumberRules, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *ComputeFlatFieldsRules) SetCpuCoreLimit(v NumberRules)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *ComputeFlatFieldsRules) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *ComputeFlatFieldsRules) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *ComputeFlatFieldsRules) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *ComputeFlatFieldsRules) GetCpuMemoryRequest() QuantityRules`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *ComputeFlatFieldsRules) GetCpuMemoryRequestOk() (*QuantityRules, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *ComputeFlatFieldsRules) SetCpuMemoryRequest(v QuantityRules)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *ComputeFlatFieldsRules) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *ComputeFlatFieldsRules) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *ComputeFlatFieldsRules) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *ComputeFlatFieldsRules) GetCpuMemoryLimit() QuantityRules`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *ComputeFlatFieldsRules) GetCpuMemoryLimitOk() (*QuantityRules, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *ComputeFlatFieldsRules) SetCpuMemoryLimit(v QuantityRules)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *ComputeFlatFieldsRules) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *ComputeFlatFieldsRules) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *ComputeFlatFieldsRules) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetLargeShmRequest

`func (o *ComputeFlatFieldsRules) GetLargeShmRequest() BooleanRules`

GetLargeShmRequest returns the LargeShmRequest field if non-nil, zero value otherwise.

### GetLargeShmRequestOk

`func (o *ComputeFlatFieldsRules) GetLargeShmRequestOk() (*BooleanRules, bool)`

GetLargeShmRequestOk returns a tuple with the LargeShmRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLargeShmRequest

`func (o *ComputeFlatFieldsRules) SetLargeShmRequest(v BooleanRules)`

SetLargeShmRequest sets LargeShmRequest field to given value.

### HasLargeShmRequest

`func (o *ComputeFlatFieldsRules) HasLargeShmRequest() bool`

HasLargeShmRequest returns a boolean if a field has been set.

### SetLargeShmRequestNil

`func (o *ComputeFlatFieldsRules) SetLargeShmRequestNil(b bool)`

 SetLargeShmRequestNil sets the value for LargeShmRequest to be an explicit nil

### UnsetLargeShmRequest
`func (o *ComputeFlatFieldsRules) UnsetLargeShmRequest()`

UnsetLargeShmRequest ensures that no value is present for LargeShmRequest, not even an explicit nil
### GetGpuRequestType

`func (o *ComputeFlatFieldsRules) GetGpuRequestType() GpuRequestRules`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *ComputeFlatFieldsRules) GetGpuRequestTypeOk() (*GpuRequestRules, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *ComputeFlatFieldsRules) SetGpuRequestType(v GpuRequestRules)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *ComputeFlatFieldsRules) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *ComputeFlatFieldsRules) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *ComputeFlatFieldsRules) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetMigProfile

`func (o *ComputeFlatFieldsRules) GetMigProfile() MigProfileRules`

GetMigProfile returns the MigProfile field if non-nil, zero value otherwise.

### GetMigProfileOk

`func (o *ComputeFlatFieldsRules) GetMigProfileOk() (*MigProfileRules, bool)`

GetMigProfileOk returns a tuple with the MigProfile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMigProfile

`func (o *ComputeFlatFieldsRules) SetMigProfile(v MigProfileRules)`

SetMigProfile sets MigProfile field to given value.

### HasMigProfile

`func (o *ComputeFlatFieldsRules) HasMigProfile() bool`

HasMigProfile returns a boolean if a field has been set.

### SetMigProfileNil

`func (o *ComputeFlatFieldsRules) SetMigProfileNil(b bool)`

 SetMigProfileNil sets the value for MigProfile to be an explicit nil

### UnsetMigProfile
`func (o *ComputeFlatFieldsRules) UnsetMigProfile()`

UnsetMigProfile ensures that no value is present for MigProfile, not even an explicit nil
### GetGpuDevicesRequest

`func (o *ComputeFlatFieldsRules) GetGpuDevicesRequest() IntegerRules`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *ComputeFlatFieldsRules) GetGpuDevicesRequestOk() (*IntegerRules, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *ComputeFlatFieldsRules) SetGpuDevicesRequest(v IntegerRules)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *ComputeFlatFieldsRules) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *ComputeFlatFieldsRules) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *ComputeFlatFieldsRules) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuPortionRequest

`func (o *ComputeFlatFieldsRules) GetGpuPortionRequest() NumberRules`

GetGpuPortionRequest returns the GpuPortionRequest field if non-nil, zero value otherwise.

### GetGpuPortionRequestOk

`func (o *ComputeFlatFieldsRules) GetGpuPortionRequestOk() (*NumberRules, bool)`

GetGpuPortionRequestOk returns a tuple with the GpuPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionRequest

`func (o *ComputeFlatFieldsRules) SetGpuPortionRequest(v NumberRules)`

SetGpuPortionRequest sets GpuPortionRequest field to given value.

### HasGpuPortionRequest

`func (o *ComputeFlatFieldsRules) HasGpuPortionRequest() bool`

HasGpuPortionRequest returns a boolean if a field has been set.

### SetGpuPortionRequestNil

`func (o *ComputeFlatFieldsRules) SetGpuPortionRequestNil(b bool)`

 SetGpuPortionRequestNil sets the value for GpuPortionRequest to be an explicit nil

### UnsetGpuPortionRequest
`func (o *ComputeFlatFieldsRules) UnsetGpuPortionRequest()`

UnsetGpuPortionRequest ensures that no value is present for GpuPortionRequest, not even an explicit nil
### GetGpuPortionLimit

`func (o *ComputeFlatFieldsRules) GetGpuPortionLimit() NumberRules`

GetGpuPortionLimit returns the GpuPortionLimit field if non-nil, zero value otherwise.

### GetGpuPortionLimitOk

`func (o *ComputeFlatFieldsRules) GetGpuPortionLimitOk() (*NumberRules, bool)`

GetGpuPortionLimitOk returns a tuple with the GpuPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionLimit

`func (o *ComputeFlatFieldsRules) SetGpuPortionLimit(v NumberRules)`

SetGpuPortionLimit sets GpuPortionLimit field to given value.

### HasGpuPortionLimit

`func (o *ComputeFlatFieldsRules) HasGpuPortionLimit() bool`

HasGpuPortionLimit returns a boolean if a field has been set.

### SetGpuPortionLimitNil

`func (o *ComputeFlatFieldsRules) SetGpuPortionLimitNil(b bool)`

 SetGpuPortionLimitNil sets the value for GpuPortionLimit to be an explicit nil

### UnsetGpuPortionLimit
`func (o *ComputeFlatFieldsRules) UnsetGpuPortionLimit()`

UnsetGpuPortionLimit ensures that no value is present for GpuPortionLimit, not even an explicit nil
### GetGpuMemoryRequest

`func (o *ComputeFlatFieldsRules) GetGpuMemoryRequest() QuantityRules`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *ComputeFlatFieldsRules) GetGpuMemoryRequestOk() (*QuantityRules, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *ComputeFlatFieldsRules) SetGpuMemoryRequest(v QuantityRules)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *ComputeFlatFieldsRules) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *ComputeFlatFieldsRules) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *ComputeFlatFieldsRules) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *ComputeFlatFieldsRules) GetGpuMemoryLimit() QuantityRules`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *ComputeFlatFieldsRules) GetGpuMemoryLimitOk() (*QuantityRules, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *ComputeFlatFieldsRules) SetGpuMemoryLimit(v QuantityRules)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *ComputeFlatFieldsRules) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *ComputeFlatFieldsRules) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *ComputeFlatFieldsRules) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


