# NimServiceRulesAllOfCompute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CpuCoreLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuCoreRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**CpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuDevicesRequest** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**GpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuPortionLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuPortionRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuRequestType** | Pointer to [**NullableGpuRequestRules**](GpuRequestRules.md) |  | [optional] 

## Methods

### NewNimServiceRulesAllOfCompute

`func NewNimServiceRulesAllOfCompute() *NimServiceRulesAllOfCompute`

NewNimServiceRulesAllOfCompute instantiates a new NimServiceRulesAllOfCompute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceRulesAllOfComputeWithDefaults

`func NewNimServiceRulesAllOfComputeWithDefaults() *NimServiceRulesAllOfCompute`

NewNimServiceRulesAllOfComputeWithDefaults instantiates a new NimServiceRulesAllOfCompute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreLimit

`func (o *NimServiceRulesAllOfCompute) GetCpuCoreLimit() NumberRules`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *NimServiceRulesAllOfCompute) GetCpuCoreLimitOk() (*NumberRules, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *NimServiceRulesAllOfCompute) SetCpuCoreLimit(v NumberRules)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *NimServiceRulesAllOfCompute) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *NimServiceRulesAllOfCompute) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *NimServiceRulesAllOfCompute) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuCoreRequest

`func (o *NimServiceRulesAllOfCompute) GetCpuCoreRequest() NumberRules`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *NimServiceRulesAllOfCompute) GetCpuCoreRequestOk() (*NumberRules, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *NimServiceRulesAllOfCompute) SetCpuCoreRequest(v NumberRules)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *NimServiceRulesAllOfCompute) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *NimServiceRulesAllOfCompute) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *NimServiceRulesAllOfCompute) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) GetCpuMemoryLimit() QuantityRules`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *NimServiceRulesAllOfCompute) GetCpuMemoryLimitOk() (*QuantityRules, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) SetCpuMemoryLimit(v QuantityRules)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *NimServiceRulesAllOfCompute) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *NimServiceRulesAllOfCompute) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) GetCpuMemoryRequest() QuantityRules`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *NimServiceRulesAllOfCompute) GetCpuMemoryRequestOk() (*QuantityRules, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) SetCpuMemoryRequest(v QuantityRules)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *NimServiceRulesAllOfCompute) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *NimServiceRulesAllOfCompute) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetGpuDevicesRequest

`func (o *NimServiceRulesAllOfCompute) GetGpuDevicesRequest() IntegerRules`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *NimServiceRulesAllOfCompute) GetGpuDevicesRequestOk() (*IntegerRules, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *NimServiceRulesAllOfCompute) SetGpuDevicesRequest(v IntegerRules)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *NimServiceRulesAllOfCompute) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *NimServiceRulesAllOfCompute) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *NimServiceRulesAllOfCompute) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) GetGpuMemoryLimit() QuantityRules`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *NimServiceRulesAllOfCompute) GetGpuMemoryLimitOk() (*QuantityRules, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) SetGpuMemoryLimit(v QuantityRules)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *NimServiceRulesAllOfCompute) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *NimServiceRulesAllOfCompute) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *NimServiceRulesAllOfCompute) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil
### GetGpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) GetGpuMemoryRequest() QuantityRules`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *NimServiceRulesAllOfCompute) GetGpuMemoryRequestOk() (*QuantityRules, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) SetGpuMemoryRequest(v QuantityRules)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *NimServiceRulesAllOfCompute) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *NimServiceRulesAllOfCompute) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *NimServiceRulesAllOfCompute) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuPortionLimit

`func (o *NimServiceRulesAllOfCompute) GetGpuPortionLimit() NumberRules`

GetGpuPortionLimit returns the GpuPortionLimit field if non-nil, zero value otherwise.

### GetGpuPortionLimitOk

`func (o *NimServiceRulesAllOfCompute) GetGpuPortionLimitOk() (*NumberRules, bool)`

GetGpuPortionLimitOk returns a tuple with the GpuPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionLimit

`func (o *NimServiceRulesAllOfCompute) SetGpuPortionLimit(v NumberRules)`

SetGpuPortionLimit sets GpuPortionLimit field to given value.

### HasGpuPortionLimit

`func (o *NimServiceRulesAllOfCompute) HasGpuPortionLimit() bool`

HasGpuPortionLimit returns a boolean if a field has been set.

### SetGpuPortionLimitNil

`func (o *NimServiceRulesAllOfCompute) SetGpuPortionLimitNil(b bool)`

 SetGpuPortionLimitNil sets the value for GpuPortionLimit to be an explicit nil

### UnsetGpuPortionLimit
`func (o *NimServiceRulesAllOfCompute) UnsetGpuPortionLimit()`

UnsetGpuPortionLimit ensures that no value is present for GpuPortionLimit, not even an explicit nil
### GetGpuPortionRequest

`func (o *NimServiceRulesAllOfCompute) GetGpuPortionRequest() NumberRules`

GetGpuPortionRequest returns the GpuPortionRequest field if non-nil, zero value otherwise.

### GetGpuPortionRequestOk

`func (o *NimServiceRulesAllOfCompute) GetGpuPortionRequestOk() (*NumberRules, bool)`

GetGpuPortionRequestOk returns a tuple with the GpuPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionRequest

`func (o *NimServiceRulesAllOfCompute) SetGpuPortionRequest(v NumberRules)`

SetGpuPortionRequest sets GpuPortionRequest field to given value.

### HasGpuPortionRequest

`func (o *NimServiceRulesAllOfCompute) HasGpuPortionRequest() bool`

HasGpuPortionRequest returns a boolean if a field has been set.

### SetGpuPortionRequestNil

`func (o *NimServiceRulesAllOfCompute) SetGpuPortionRequestNil(b bool)`

 SetGpuPortionRequestNil sets the value for GpuPortionRequest to be an explicit nil

### UnsetGpuPortionRequest
`func (o *NimServiceRulesAllOfCompute) UnsetGpuPortionRequest()`

UnsetGpuPortionRequest ensures that no value is present for GpuPortionRequest, not even an explicit nil
### GetGpuRequestType

`func (o *NimServiceRulesAllOfCompute) GetGpuRequestType() GpuRequestRules`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *NimServiceRulesAllOfCompute) GetGpuRequestTypeOk() (*GpuRequestRules, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *NimServiceRulesAllOfCompute) SetGpuRequestType(v GpuRequestRules)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *NimServiceRulesAllOfCompute) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *NimServiceRulesAllOfCompute) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *NimServiceRulesAllOfCompute) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


