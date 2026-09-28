# SupersetRulesAllOfCompute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CpuCoreLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuCoreRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**CpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**CpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**ExtendedResources** | Pointer to [**NullableExtendedResourcesRules**](ExtendedResourcesRules.md) |  | [optional] 
**GpuDevicesRequest** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**GpuMemoryLimit** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuMemoryRequest** | Pointer to [**NullableQuantityRules**](QuantityRules.md) |  | [optional] 
**GpuPortionLimit** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuPortionRequest** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**GpuRequestType** | Pointer to [**NullableGpuRequestRules**](GpuRequestRules.md) |  | [optional] 
**LargeShmRequest** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewSupersetRulesAllOfCompute

`func NewSupersetRulesAllOfCompute() *SupersetRulesAllOfCompute`

NewSupersetRulesAllOfCompute instantiates a new SupersetRulesAllOfCompute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetRulesAllOfComputeWithDefaults

`func NewSupersetRulesAllOfComputeWithDefaults() *SupersetRulesAllOfCompute`

NewSupersetRulesAllOfComputeWithDefaults instantiates a new SupersetRulesAllOfCompute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreLimit

`func (o *SupersetRulesAllOfCompute) GetCpuCoreLimit() NumberRules`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *SupersetRulesAllOfCompute) GetCpuCoreLimitOk() (*NumberRules, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *SupersetRulesAllOfCompute) SetCpuCoreLimit(v NumberRules)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *SupersetRulesAllOfCompute) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *SupersetRulesAllOfCompute) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *SupersetRulesAllOfCompute) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuCoreRequest

`func (o *SupersetRulesAllOfCompute) GetCpuCoreRequest() NumberRules`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *SupersetRulesAllOfCompute) GetCpuCoreRequestOk() (*NumberRules, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *SupersetRulesAllOfCompute) SetCpuCoreRequest(v NumberRules)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *SupersetRulesAllOfCompute) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *SupersetRulesAllOfCompute) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *SupersetRulesAllOfCompute) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) GetCpuMemoryLimit() QuantityRules`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *SupersetRulesAllOfCompute) GetCpuMemoryLimitOk() (*QuantityRules, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) SetCpuMemoryLimit(v QuantityRules)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *SupersetRulesAllOfCompute) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *SupersetRulesAllOfCompute) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) GetCpuMemoryRequest() QuantityRules`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *SupersetRulesAllOfCompute) GetCpuMemoryRequestOk() (*QuantityRules, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) SetCpuMemoryRequest(v QuantityRules)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *SupersetRulesAllOfCompute) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *SupersetRulesAllOfCompute) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetExtendedResources

`func (o *SupersetRulesAllOfCompute) GetExtendedResources() ExtendedResourcesRules`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *SupersetRulesAllOfCompute) GetExtendedResourcesOk() (*ExtendedResourcesRules, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *SupersetRulesAllOfCompute) SetExtendedResources(v ExtendedResourcesRules)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *SupersetRulesAllOfCompute) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *SupersetRulesAllOfCompute) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *SupersetRulesAllOfCompute) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil
### GetGpuDevicesRequest

`func (o *SupersetRulesAllOfCompute) GetGpuDevicesRequest() IntegerRules`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *SupersetRulesAllOfCompute) GetGpuDevicesRequestOk() (*IntegerRules, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *SupersetRulesAllOfCompute) SetGpuDevicesRequest(v IntegerRules)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *SupersetRulesAllOfCompute) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *SupersetRulesAllOfCompute) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *SupersetRulesAllOfCompute) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) GetGpuMemoryLimit() QuantityRules`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *SupersetRulesAllOfCompute) GetGpuMemoryLimitOk() (*QuantityRules, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) SetGpuMemoryLimit(v QuantityRules)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *SupersetRulesAllOfCompute) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *SupersetRulesAllOfCompute) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *SupersetRulesAllOfCompute) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil
### GetGpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) GetGpuMemoryRequest() QuantityRules`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *SupersetRulesAllOfCompute) GetGpuMemoryRequestOk() (*QuantityRules, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) SetGpuMemoryRequest(v QuantityRules)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *SupersetRulesAllOfCompute) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *SupersetRulesAllOfCompute) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *SupersetRulesAllOfCompute) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuPortionLimit

`func (o *SupersetRulesAllOfCompute) GetGpuPortionLimit() NumberRules`

GetGpuPortionLimit returns the GpuPortionLimit field if non-nil, zero value otherwise.

### GetGpuPortionLimitOk

`func (o *SupersetRulesAllOfCompute) GetGpuPortionLimitOk() (*NumberRules, bool)`

GetGpuPortionLimitOk returns a tuple with the GpuPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionLimit

`func (o *SupersetRulesAllOfCompute) SetGpuPortionLimit(v NumberRules)`

SetGpuPortionLimit sets GpuPortionLimit field to given value.

### HasGpuPortionLimit

`func (o *SupersetRulesAllOfCompute) HasGpuPortionLimit() bool`

HasGpuPortionLimit returns a boolean if a field has been set.

### SetGpuPortionLimitNil

`func (o *SupersetRulesAllOfCompute) SetGpuPortionLimitNil(b bool)`

 SetGpuPortionLimitNil sets the value for GpuPortionLimit to be an explicit nil

### UnsetGpuPortionLimit
`func (o *SupersetRulesAllOfCompute) UnsetGpuPortionLimit()`

UnsetGpuPortionLimit ensures that no value is present for GpuPortionLimit, not even an explicit nil
### GetGpuPortionRequest

`func (o *SupersetRulesAllOfCompute) GetGpuPortionRequest() NumberRules`

GetGpuPortionRequest returns the GpuPortionRequest field if non-nil, zero value otherwise.

### GetGpuPortionRequestOk

`func (o *SupersetRulesAllOfCompute) GetGpuPortionRequestOk() (*NumberRules, bool)`

GetGpuPortionRequestOk returns a tuple with the GpuPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionRequest

`func (o *SupersetRulesAllOfCompute) SetGpuPortionRequest(v NumberRules)`

SetGpuPortionRequest sets GpuPortionRequest field to given value.

### HasGpuPortionRequest

`func (o *SupersetRulesAllOfCompute) HasGpuPortionRequest() bool`

HasGpuPortionRequest returns a boolean if a field has been set.

### SetGpuPortionRequestNil

`func (o *SupersetRulesAllOfCompute) SetGpuPortionRequestNil(b bool)`

 SetGpuPortionRequestNil sets the value for GpuPortionRequest to be an explicit nil

### UnsetGpuPortionRequest
`func (o *SupersetRulesAllOfCompute) UnsetGpuPortionRequest()`

UnsetGpuPortionRequest ensures that no value is present for GpuPortionRequest, not even an explicit nil
### GetGpuRequestType

`func (o *SupersetRulesAllOfCompute) GetGpuRequestType() GpuRequestRules`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *SupersetRulesAllOfCompute) GetGpuRequestTypeOk() (*GpuRequestRules, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *SupersetRulesAllOfCompute) SetGpuRequestType(v GpuRequestRules)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *SupersetRulesAllOfCompute) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *SupersetRulesAllOfCompute) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *SupersetRulesAllOfCompute) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetLargeShmRequest

`func (o *SupersetRulesAllOfCompute) GetLargeShmRequest() BooleanRules`

GetLargeShmRequest returns the LargeShmRequest field if non-nil, zero value otherwise.

### GetLargeShmRequestOk

`func (o *SupersetRulesAllOfCompute) GetLargeShmRequestOk() (*BooleanRules, bool)`

GetLargeShmRequestOk returns a tuple with the LargeShmRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLargeShmRequest

`func (o *SupersetRulesAllOfCompute) SetLargeShmRequest(v BooleanRules)`

SetLargeShmRequest sets LargeShmRequest field to given value.

### HasLargeShmRequest

`func (o *SupersetRulesAllOfCompute) HasLargeShmRequest() bool`

HasLargeShmRequest returns a boolean if a field has been set.

### SetLargeShmRequestNil

`func (o *SupersetRulesAllOfCompute) SetLargeShmRequestNil(b bool)`

 SetLargeShmRequestNil sets the value for LargeShmRequest to be an explicit nil

### UnsetLargeShmRequest
`func (o *SupersetRulesAllOfCompute) UnsetLargeShmRequest()`

UnsetLargeShmRequest ensures that no value is present for LargeShmRequest, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


