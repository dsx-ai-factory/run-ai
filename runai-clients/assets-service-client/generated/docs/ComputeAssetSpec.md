# ComputeAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CpuCoreLimit** | Pointer to **NullableFloat64** | Limitations on the number of CPUs consumed by the workload (0.5, 1, .etc). The system guarantees that this workload will not be able to consume more than this amount of CPUs. | [optional] 
**CpuCoreRequest** | Pointer to **NullableFloat64** | CPU units to allocate for the created workload (0.5, 1, .etc). The workload will receive at least this amount of CPU. Note that the workload will not be scheduled unless the system can guarantee this amount of CPUs to the workload. | [optional] 
**CpuMemoryLimit** | Pointer to **NullableString** | Limitations on the CPU memory to allocate for this workload (1G, 20M, .etc). The system guarantees that this workload will not be able to consume more than this amount of memory. The workload will receive an error when trying to allocate more memory than this limit. | [optional] 
**CpuMemoryRequest** | Pointer to **NullableString** | The amount of CPU memory to allocate for this workload (1G, 20M, .etc). The workload will receive at least this amount of memory. Note that the workload will not be scheduled unless the system can guarantee this amount of memory to the workload | [optional] 
**ExtendedResources** | Pointer to [**[]ExtendedResource**](ExtendedResource.md) | Extended resources and their quantity. | [optional] 
**GpuDevicesRequest** | Pointer to **NullableInt32** | Requested number of GPU devices. Currently if more than one device is requested, it is not possible to provide values for gpuMemory or gpuPortion. | [optional] 
**GpuMemoryLimit** | Pointer to **NullableString** | Limitation on the memory consumed by the workload, per GPU device. The system guarantees The gpuMemoryLimit must be no less than gpuMemoryRequest. | [optional] 
**GpuMemoryRequest** | Pointer to **NullableString** | Required if and only if gpuRequestType is memory. States the GPU memory to allocate for the created workload, per GPU device. Note that the workload will not be scheduled unless the system can guarantee this amount of GPU memory to the workload. | [optional] 
**GpuPortionLimit** | Pointer to **NullableFloat64** | Limitations on the portion consumed by the workload, per GPU device. The system guarantees The gpuPotionLimit must be no less than the gpuPortionRequest. | [optional] 
**GpuPortionRequest** | Pointer to **NullableFloat64** | Required if and only if gpuRequestType is portion. States the portion of the GPU to allocate for the created workload, per GPU device, between 0 and 1. The default is no allocated GPUs. | [optional] 
**GpuRequestType** | Pointer to [**NullableGpuRequestType**](GpuRequestType.md) |  | [optional] 
**LargeShmRequest** | Pointer to **NullableBool** | A large /dev/shm device to mount into a container running the created workload. An shm is a shared file system mounted on RAM. | [optional] 

## Methods

### NewComputeAssetSpec

`func NewComputeAssetSpec() *ComputeAssetSpec`

NewComputeAssetSpec instantiates a new ComputeAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputeAssetSpecWithDefaults

`func NewComputeAssetSpecWithDefaults() *ComputeAssetSpec`

NewComputeAssetSpecWithDefaults instantiates a new ComputeAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreLimit

`func (o *ComputeAssetSpec) GetCpuCoreLimit() float64`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *ComputeAssetSpec) GetCpuCoreLimitOk() (*float64, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *ComputeAssetSpec) SetCpuCoreLimit(v float64)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *ComputeAssetSpec) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *ComputeAssetSpec) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *ComputeAssetSpec) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuCoreRequest

`func (o *ComputeAssetSpec) GetCpuCoreRequest() float64`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *ComputeAssetSpec) GetCpuCoreRequestOk() (*float64, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *ComputeAssetSpec) SetCpuCoreRequest(v float64)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *ComputeAssetSpec) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *ComputeAssetSpec) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *ComputeAssetSpec) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *ComputeAssetSpec) GetCpuMemoryLimit() string`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *ComputeAssetSpec) GetCpuMemoryLimitOk() (*string, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *ComputeAssetSpec) SetCpuMemoryLimit(v string)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *ComputeAssetSpec) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *ComputeAssetSpec) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *ComputeAssetSpec) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *ComputeAssetSpec) GetCpuMemoryRequest() string`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *ComputeAssetSpec) GetCpuMemoryRequestOk() (*string, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *ComputeAssetSpec) SetCpuMemoryRequest(v string)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *ComputeAssetSpec) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *ComputeAssetSpec) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *ComputeAssetSpec) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetExtendedResources

`func (o *ComputeAssetSpec) GetExtendedResources() []ExtendedResource`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *ComputeAssetSpec) GetExtendedResourcesOk() (*[]ExtendedResource, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *ComputeAssetSpec) SetExtendedResources(v []ExtendedResource)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *ComputeAssetSpec) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *ComputeAssetSpec) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *ComputeAssetSpec) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil
### GetGpuDevicesRequest

`func (o *ComputeAssetSpec) GetGpuDevicesRequest() int32`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *ComputeAssetSpec) GetGpuDevicesRequestOk() (*int32, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *ComputeAssetSpec) SetGpuDevicesRequest(v int32)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *ComputeAssetSpec) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *ComputeAssetSpec) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *ComputeAssetSpec) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *ComputeAssetSpec) GetGpuMemoryLimit() string`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *ComputeAssetSpec) GetGpuMemoryLimitOk() (*string, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *ComputeAssetSpec) SetGpuMemoryLimit(v string)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *ComputeAssetSpec) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *ComputeAssetSpec) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *ComputeAssetSpec) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil
### GetGpuMemoryRequest

`func (o *ComputeAssetSpec) GetGpuMemoryRequest() string`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *ComputeAssetSpec) GetGpuMemoryRequestOk() (*string, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *ComputeAssetSpec) SetGpuMemoryRequest(v string)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *ComputeAssetSpec) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *ComputeAssetSpec) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *ComputeAssetSpec) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuPortionLimit

`func (o *ComputeAssetSpec) GetGpuPortionLimit() float64`

GetGpuPortionLimit returns the GpuPortionLimit field if non-nil, zero value otherwise.

### GetGpuPortionLimitOk

`func (o *ComputeAssetSpec) GetGpuPortionLimitOk() (*float64, bool)`

GetGpuPortionLimitOk returns a tuple with the GpuPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionLimit

`func (o *ComputeAssetSpec) SetGpuPortionLimit(v float64)`

SetGpuPortionLimit sets GpuPortionLimit field to given value.

### HasGpuPortionLimit

`func (o *ComputeAssetSpec) HasGpuPortionLimit() bool`

HasGpuPortionLimit returns a boolean if a field has been set.

### SetGpuPortionLimitNil

`func (o *ComputeAssetSpec) SetGpuPortionLimitNil(b bool)`

 SetGpuPortionLimitNil sets the value for GpuPortionLimit to be an explicit nil

### UnsetGpuPortionLimit
`func (o *ComputeAssetSpec) UnsetGpuPortionLimit()`

UnsetGpuPortionLimit ensures that no value is present for GpuPortionLimit, not even an explicit nil
### GetGpuPortionRequest

`func (o *ComputeAssetSpec) GetGpuPortionRequest() float64`

GetGpuPortionRequest returns the GpuPortionRequest field if non-nil, zero value otherwise.

### GetGpuPortionRequestOk

`func (o *ComputeAssetSpec) GetGpuPortionRequestOk() (*float64, bool)`

GetGpuPortionRequestOk returns a tuple with the GpuPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPortionRequest

`func (o *ComputeAssetSpec) SetGpuPortionRequest(v float64)`

SetGpuPortionRequest sets GpuPortionRequest field to given value.

### HasGpuPortionRequest

`func (o *ComputeAssetSpec) HasGpuPortionRequest() bool`

HasGpuPortionRequest returns a boolean if a field has been set.

### SetGpuPortionRequestNil

`func (o *ComputeAssetSpec) SetGpuPortionRequestNil(b bool)`

 SetGpuPortionRequestNil sets the value for GpuPortionRequest to be an explicit nil

### UnsetGpuPortionRequest
`func (o *ComputeAssetSpec) UnsetGpuPortionRequest()`

UnsetGpuPortionRequest ensures that no value is present for GpuPortionRequest, not even an explicit nil
### GetGpuRequestType

`func (o *ComputeAssetSpec) GetGpuRequestType() GpuRequestType`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *ComputeAssetSpec) GetGpuRequestTypeOk() (*GpuRequestType, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *ComputeAssetSpec) SetGpuRequestType(v GpuRequestType)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *ComputeAssetSpec) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *ComputeAssetSpec) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *ComputeAssetSpec) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetLargeShmRequest

`func (o *ComputeAssetSpec) GetLargeShmRequest() bool`

GetLargeShmRequest returns the LargeShmRequest field if non-nil, zero value otherwise.

### GetLargeShmRequestOk

`func (o *ComputeAssetSpec) GetLargeShmRequestOk() (*bool, bool)`

GetLargeShmRequestOk returns a tuple with the LargeShmRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLargeShmRequest

`func (o *ComputeAssetSpec) SetLargeShmRequest(v bool)`

SetLargeShmRequest sets LargeShmRequest field to given value.

### HasLargeShmRequest

`func (o *ComputeAssetSpec) HasLargeShmRequest() bool`

HasLargeShmRequest returns a boolean if a field has been set.

### SetLargeShmRequestNil

`func (o *ComputeAssetSpec) SetLargeShmRequestNil(b bool)`

 SetLargeShmRequestNil sets the value for LargeShmRequest to be an explicit nil

### UnsetLargeShmRequest
`func (o *ComputeAssetSpec) UnsetLargeShmRequest()`

UnsetLargeShmRequest ensures that no value is present for LargeShmRequest, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


