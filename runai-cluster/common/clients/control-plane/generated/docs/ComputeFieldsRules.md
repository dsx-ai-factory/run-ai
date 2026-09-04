# ComputeFieldsRules

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
**ExtendedResources** | Pointer to [**NullableExtendedResourcesRules**](ExtendedResourcesRules.md) |  | [optional] 

## Methods

### NewComputeFieldsRules

`func NewComputeFieldsRules() *ComputeFieldsRules`

NewComputeFieldsRules instantiates a new ComputeFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputeFieldsRulesWithDefaults

`func NewComputeFieldsRulesWithDefaults() *ComputeFieldsRules`

NewComputeFieldsRulesWithDefaults instantiates a new ComputeFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreRequest

`func (o *ComputeFieldsRules) GetCpuCoreRequest() NumberRules`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *ComputeFieldsRules) GetCpuCoreRequestOk() (*NumberRules, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *ComputeFieldsRules) SetCpuCoreRequest(v NumberRules)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *ComputeFieldsRules) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *ComputeFieldsRules) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *ComputeFieldsRules) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuCoreLimit

`func (o *ComputeFieldsRules) GetCpuCoreLimit() NumberRules`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *ComputeFieldsRules) GetCpuCoreLimitOk() (*NumberRules, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *ComputeFieldsRules) SetCpuCoreLimit(v NumberRules)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *ComputeFieldsRules) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *ComputeFieldsRules) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *ComputeFieldsRules) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *ComputeFieldsRules) GetCpuMemoryRequest() QuantityRules`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *ComputeFieldsRules) GetCpuMemoryRequestOk() (*QuantityRules, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *ComputeFieldsRules) SetCpuMemoryRequest(v QuantityRules)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *ComputeFieldsRules) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *ComputeFieldsRules) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *ComputeFieldsRules) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *ComputeFieldsRules) GetCpuMemoryLimit() QuantityRules`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *ComputeFieldsRules) GetCpuMemoryLimitOk() (*QuantityRules, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *ComputeFieldsRules) SetCpuMemoryLimit(v QuantityRules)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *ComputeFieldsRules) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *ComputeFieldsRules) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *ComputeFieldsRules) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetLargeShmRequest

`func (o *ComputeFieldsRules) GetLargeShmRequest() BooleanRules`

GetLargeShmRequest returns the LargeShmRequest field if non-nil, zero value otherwise.

### GetLargeShmRequestOk

`func (o *ComputeFieldsRules) GetLargeShmRequestOk() (*BooleanRules, bool)`

GetLargeShmRequestOk returns a tuple with the LargeShmRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLargeShmRequest

`func (o *ComputeFieldsRules) SetLargeShmRequest(v BooleanRules)`

SetLargeShmRequest sets LargeShmRequest field to given value.

### HasLargeShmRequest

`func (o *ComputeFieldsRules) HasLargeShmRequest() bool`

HasLargeShmRequest returns a boolean if a field has been set.

### SetLargeShmRequestNil

`func (o *ComputeFieldsRules) SetLargeShmRequestNil(b bool)`

 SetLargeShmRequestNil sets the value for LargeShmRequest to be an explicit nil

### UnsetLargeShmRequest
`func (o *ComputeFieldsRules) UnsetLargeShmRequest()`

UnsetLargeShmRequest ensures that no value is present for LargeShmRequest, not even an explicit nil
### GetGpuRequestType

`func (o *ComputeFieldsRules) GetGpuRequestType() GpuRequestRules`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *ComputeFieldsRules) GetGpuRequestTypeOk() (*GpuRequestRules, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *ComputeFieldsRules) SetGpuRequestType(v GpuRequestRules)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *ComputeFieldsRules) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *ComputeFieldsRules) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *ComputeFieldsRules) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetMigProfile

`func (o *ComputeFieldsRules) GetMigProfile() MigProfileRules`

GetMigProfile returns the MigProfile field if non-nil, zero value otherwise.

### GetMigProfileOk

`func (o *ComputeFieldsRules) GetMigProfileOk() (*MigProfileRules, bool)`

GetMigProfileOk returns a tuple with the MigProfile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMigProfile

`func (o *ComputeFieldsRules) SetMigProfile(v MigProfileRules)`

SetMigProfile sets MigProfile field to given value.

### HasMigProfile

`func (o *ComputeFieldsRules) HasMigProfile() bool`

HasMigProfile returns a boolean if a field has been set.

### SetMigProfileNil

`func (o *ComputeFieldsRules) SetMigProfileNil(b bool)`

 SetMigProfileNil sets the value for MigProfile to be an explicit nil

### UnsetMigProfile
`func (o *ComputeFieldsRules) UnsetMigProfile()`

UnsetMigProfile ensures that no value is present for MigProfile, not even an explicit nil
### GetGpuDevicesRequest

`func (o *ComputeFieldsRules) GetGpuDevicesRequest() IntegerRules`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *ComputeFieldsRules) GetGpuDevicesRequestOk() (*IntegerRules, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *ComputeFieldsRules) SetGpuDevicesRequest(v IntegerRules)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *ComputeFieldsRules) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *ComputeFieldsRules) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *ComputeFieldsRules) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuPortionRequest

`func (o *ComputeFieldsRules) GetGpuPortionRequest() NumberRules`

GetGpuPortionRequest returns the GpuPortionRequest field if non-nil, zero value otherwise.

### GetGpuPortionRequestOk

`func (o *ComputeFieldsRules) GetGpuPortionRequestOk() (*NumberRules, bool)`

GetGpuPortionRequestOk returns a tuple with the GpuPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionRequest

`func (o *ComputeFieldsRules) SetGpuPortionRequest(v NumberRules)`

SetGpuPortionRequest sets GpuPortionRequest field to given value.

### HasGpuPortionRequest

`func (o *ComputeFieldsRules) HasGpuPortionRequest() bool`

HasGpuPortionRequest returns a boolean if a field has been set.

### SetGpuPortionRequestNil

`func (o *ComputeFieldsRules) SetGpuPortionRequestNil(b bool)`

 SetGpuPortionRequestNil sets the value for GpuPortionRequest to be an explicit nil

### UnsetGpuPortionRequest
`func (o *ComputeFieldsRules) UnsetGpuPortionRequest()`

UnsetGpuPortionRequest ensures that no value is present for GpuPortionRequest, not even an explicit nil
### GetGpuPortionLimit

`func (o *ComputeFieldsRules) GetGpuPortionLimit() NumberRules`

GetGpuPortionLimit returns the GpuPortionLimit field if non-nil, zero value otherwise.

### GetGpuPortionLimitOk

`func (o *ComputeFieldsRules) GetGpuPortionLimitOk() (*NumberRules, bool)`

GetGpuPortionLimitOk returns a tuple with the GpuPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionLimit

`func (o *ComputeFieldsRules) SetGpuPortionLimit(v NumberRules)`

SetGpuPortionLimit sets GpuPortionLimit field to given value.

### HasGpuPortionLimit

`func (o *ComputeFieldsRules) HasGpuPortionLimit() bool`

HasGpuPortionLimit returns a boolean if a field has been set.

### SetGpuPortionLimitNil

`func (o *ComputeFieldsRules) SetGpuPortionLimitNil(b bool)`

 SetGpuPortionLimitNil sets the value for GpuPortionLimit to be an explicit nil

### UnsetGpuPortionLimit
`func (o *ComputeFieldsRules) UnsetGpuPortionLimit()`

UnsetGpuPortionLimit ensures that no value is present for GpuPortionLimit, not even an explicit nil
### GetGpuMemoryRequest

`func (o *ComputeFieldsRules) GetGpuMemoryRequest() QuantityRules`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *ComputeFieldsRules) GetGpuMemoryRequestOk() (*QuantityRules, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *ComputeFieldsRules) SetGpuMemoryRequest(v QuantityRules)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *ComputeFieldsRules) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *ComputeFieldsRules) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *ComputeFieldsRules) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *ComputeFieldsRules) GetGpuMemoryLimit() QuantityRules`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *ComputeFieldsRules) GetGpuMemoryLimitOk() (*QuantityRules, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *ComputeFieldsRules) SetGpuMemoryLimit(v QuantityRules)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *ComputeFieldsRules) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *ComputeFieldsRules) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *ComputeFieldsRules) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil
### GetExtendedResources

`func (o *ComputeFieldsRules) GetExtendedResources() ExtendedResourcesRules`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *ComputeFieldsRules) GetExtendedResourcesOk() (*ExtendedResourcesRules, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *ComputeFieldsRules) SetExtendedResources(v ExtendedResourcesRules)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *ComputeFieldsRules) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *ComputeFieldsRules) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *ComputeFieldsRules) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


