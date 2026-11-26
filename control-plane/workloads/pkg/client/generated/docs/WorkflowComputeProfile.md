# WorkflowComputeProfile

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CpuCoreLimit** | Pointer to **NullableFloat64** | Limitations on the number of CPUs consumed by the workload (0.5, 1, .etc). The system guarantees that this workload will not be able to consume more than this amount of CPUs. | [optional] 
**CpuCoreRequest** | Pointer to **NullableFloat64** | CPU units to allocate for the created workload (0.5, 1, .etc). The workload will receive at least this amount of CPU. Note that the workload will not be scheduled unless the system can guarantee this amount of CPUs to the workload. | [optional] 
**CpuMemoryLimit** | Pointer to **NullableString** | Limitations on the CPU memory to allocate for this workload (1G, 20M, .etc). The system guarantees that this workload will not be able to consume more than this amount of memory. The workload will receive an error when trying to allocate more memory than this limit. | [optional] 
**CpuMemoryRequest** | Pointer to **NullableString** | The amount of CPU memory to allocate for this workload (1G, 20M, .etc). The workload will receive at least this amount of memory. Note that the workload will not be scheduled unless the system can guarantee this amount of memory to the workload | [optional] 
**EphemeralStorageRequest** | Pointer to **NullableString** | The amount of local disk space (ephemeral storage) to allocate for the task (1G, 20M, etc.), including the space of its S3 inputs and outputs. The task will not be scheduled unless a node can guarantee this amount of disk space. | [optional] 
**ExtendedResources** | Pointer to [**[]ExtendedResource**](ExtendedResource.md) | Extended resources and their quantity. | [optional] 
**GpuComputeMode** | Pointer to [**NullableGpuComputeMode**](GpuComputeMode.md) |  | [optional] 
**GpuDevicesRequest** | Pointer to **NullableInt32** | Requested number of GPU devices. Currently if more than one device is requested, it is not possible to provide values for gpuMemory or gpuPortion. | [optional] 
**GpuMemoryLimit** | Pointer to **NullableString** | Limitation on the memory consumed by the workload, per GPU device. The system guarantees The gpuMemoryLimit must be no less than gpuMemoryRequest. | [optional] 
**GpuMemoryPortionLimit** | Pointer to **NullableFloat64** | Maximum GPU memory fraction consumed per GPU device. Must be no less than gpuMemoryPortionRequest. | [optional] 
**GpuMemoryPortionRequest** | Pointer to **NullableFloat64** | GPU memory fraction to allocate per GPU device, between 0 and 1. Required when gpuRequestType is portion. | [optional] 
**GpuMemoryRequest** | Pointer to **NullableString** | Required if and only if gpuRequestType is memory. States the GPU memory to allocate for the created workload, per GPU device. Note that the workload will not be scheduled unless the system can guarantee this amount of GPU memory to the workload. | [optional] 
**GpuRequestType** | Pointer to [**NullableGpuRequestType**](GpuRequestType.md) |  | [optional] 
**LargeShmRequest** | Pointer to **NullableBool** | A large /dev/shm device to mount into a container running the created workload. An shm is a shared file system mounted on RAM. | [optional] 
**Name** | Pointer to **NullableString** | The name of the profile, which tasks use to reference it. It must be unique among the profiles of its kind. (mandatory) | [optional] 

## Methods

### NewWorkflowComputeProfile

`func NewWorkflowComputeProfile() *WorkflowComputeProfile`

NewWorkflowComputeProfile instantiates a new WorkflowComputeProfile object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowComputeProfileWithDefaults

`func NewWorkflowComputeProfileWithDefaults() *WorkflowComputeProfile`

NewWorkflowComputeProfileWithDefaults instantiates a new WorkflowComputeProfile object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpuCoreLimit

`func (o *WorkflowComputeProfile) GetCpuCoreLimit() float64`

GetCpuCoreLimit returns the CpuCoreLimit field if non-nil, zero value otherwise.

### GetCpuCoreLimitOk

`func (o *WorkflowComputeProfile) GetCpuCoreLimitOk() (*float64, bool)`

GetCpuCoreLimitOk returns a tuple with the CpuCoreLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreLimit

`func (o *WorkflowComputeProfile) SetCpuCoreLimit(v float64)`

SetCpuCoreLimit sets CpuCoreLimit field to given value.

### HasCpuCoreLimit

`func (o *WorkflowComputeProfile) HasCpuCoreLimit() bool`

HasCpuCoreLimit returns a boolean if a field has been set.

### SetCpuCoreLimitNil

`func (o *WorkflowComputeProfile) SetCpuCoreLimitNil(b bool)`

 SetCpuCoreLimitNil sets the value for CpuCoreLimit to be an explicit nil

### UnsetCpuCoreLimit
`func (o *WorkflowComputeProfile) UnsetCpuCoreLimit()`

UnsetCpuCoreLimit ensures that no value is present for CpuCoreLimit, not even an explicit nil
### GetCpuCoreRequest

`func (o *WorkflowComputeProfile) GetCpuCoreRequest() float64`

GetCpuCoreRequest returns the CpuCoreRequest field if non-nil, zero value otherwise.

### GetCpuCoreRequestOk

`func (o *WorkflowComputeProfile) GetCpuCoreRequestOk() (*float64, bool)`

GetCpuCoreRequestOk returns a tuple with the CpuCoreRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuCoreRequest

`func (o *WorkflowComputeProfile) SetCpuCoreRequest(v float64)`

SetCpuCoreRequest sets CpuCoreRequest field to given value.

### HasCpuCoreRequest

`func (o *WorkflowComputeProfile) HasCpuCoreRequest() bool`

HasCpuCoreRequest returns a boolean if a field has been set.

### SetCpuCoreRequestNil

`func (o *WorkflowComputeProfile) SetCpuCoreRequestNil(b bool)`

 SetCpuCoreRequestNil sets the value for CpuCoreRequest to be an explicit nil

### UnsetCpuCoreRequest
`func (o *WorkflowComputeProfile) UnsetCpuCoreRequest()`

UnsetCpuCoreRequest ensures that no value is present for CpuCoreRequest, not even an explicit nil
### GetCpuMemoryLimit

`func (o *WorkflowComputeProfile) GetCpuMemoryLimit() string`

GetCpuMemoryLimit returns the CpuMemoryLimit field if non-nil, zero value otherwise.

### GetCpuMemoryLimitOk

`func (o *WorkflowComputeProfile) GetCpuMemoryLimitOk() (*string, bool)`

GetCpuMemoryLimitOk returns a tuple with the CpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryLimit

`func (o *WorkflowComputeProfile) SetCpuMemoryLimit(v string)`

SetCpuMemoryLimit sets CpuMemoryLimit field to given value.

### HasCpuMemoryLimit

`func (o *WorkflowComputeProfile) HasCpuMemoryLimit() bool`

HasCpuMemoryLimit returns a boolean if a field has been set.

### SetCpuMemoryLimitNil

`func (o *WorkflowComputeProfile) SetCpuMemoryLimitNil(b bool)`

 SetCpuMemoryLimitNil sets the value for CpuMemoryLimit to be an explicit nil

### UnsetCpuMemoryLimit
`func (o *WorkflowComputeProfile) UnsetCpuMemoryLimit()`

UnsetCpuMemoryLimit ensures that no value is present for CpuMemoryLimit, not even an explicit nil
### GetCpuMemoryRequest

`func (o *WorkflowComputeProfile) GetCpuMemoryRequest() string`

GetCpuMemoryRequest returns the CpuMemoryRequest field if non-nil, zero value otherwise.

### GetCpuMemoryRequestOk

`func (o *WorkflowComputeProfile) GetCpuMemoryRequestOk() (*string, bool)`

GetCpuMemoryRequestOk returns a tuple with the CpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemoryRequest

`func (o *WorkflowComputeProfile) SetCpuMemoryRequest(v string)`

SetCpuMemoryRequest sets CpuMemoryRequest field to given value.

### HasCpuMemoryRequest

`func (o *WorkflowComputeProfile) HasCpuMemoryRequest() bool`

HasCpuMemoryRequest returns a boolean if a field has been set.

### SetCpuMemoryRequestNil

`func (o *WorkflowComputeProfile) SetCpuMemoryRequestNil(b bool)`

 SetCpuMemoryRequestNil sets the value for CpuMemoryRequest to be an explicit nil

### UnsetCpuMemoryRequest
`func (o *WorkflowComputeProfile) UnsetCpuMemoryRequest()`

UnsetCpuMemoryRequest ensures that no value is present for CpuMemoryRequest, not even an explicit nil
### GetEphemeralStorageRequest

`func (o *WorkflowComputeProfile) GetEphemeralStorageRequest() string`

GetEphemeralStorageRequest returns the EphemeralStorageRequest field if non-nil, zero value otherwise.

### GetEphemeralStorageRequestOk

`func (o *WorkflowComputeProfile) GetEphemeralStorageRequestOk() (*string, bool)`

GetEphemeralStorageRequestOk returns a tuple with the EphemeralStorageRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEphemeralStorageRequest

`func (o *WorkflowComputeProfile) SetEphemeralStorageRequest(v string)`

SetEphemeralStorageRequest sets EphemeralStorageRequest field to given value.

### HasEphemeralStorageRequest

`func (o *WorkflowComputeProfile) HasEphemeralStorageRequest() bool`

HasEphemeralStorageRequest returns a boolean if a field has been set.

### SetEphemeralStorageRequestNil

`func (o *WorkflowComputeProfile) SetEphemeralStorageRequestNil(b bool)`

 SetEphemeralStorageRequestNil sets the value for EphemeralStorageRequest to be an explicit nil

### UnsetEphemeralStorageRequest
`func (o *WorkflowComputeProfile) UnsetEphemeralStorageRequest()`

UnsetEphemeralStorageRequest ensures that no value is present for EphemeralStorageRequest, not even an explicit nil
### GetExtendedResources

`func (o *WorkflowComputeProfile) GetExtendedResources() []ExtendedResource`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *WorkflowComputeProfile) GetExtendedResourcesOk() (*[]ExtendedResource, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *WorkflowComputeProfile) SetExtendedResources(v []ExtendedResource)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *WorkflowComputeProfile) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *WorkflowComputeProfile) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *WorkflowComputeProfile) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil
### GetGpuComputeMode

`func (o *WorkflowComputeProfile) GetGpuComputeMode() GpuComputeMode`

GetGpuComputeMode returns the GpuComputeMode field if non-nil, zero value otherwise.

### GetGpuComputeModeOk

`func (o *WorkflowComputeProfile) GetGpuComputeModeOk() (*GpuComputeMode, bool)`

GetGpuComputeModeOk returns a tuple with the GpuComputeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuComputeMode

`func (o *WorkflowComputeProfile) SetGpuComputeMode(v GpuComputeMode)`

SetGpuComputeMode sets GpuComputeMode field to given value.

### HasGpuComputeMode

`func (o *WorkflowComputeProfile) HasGpuComputeMode() bool`

HasGpuComputeMode returns a boolean if a field has been set.

### SetGpuComputeModeNil

`func (o *WorkflowComputeProfile) SetGpuComputeModeNil(b bool)`

 SetGpuComputeModeNil sets the value for GpuComputeMode to be an explicit nil

### UnsetGpuComputeMode
`func (o *WorkflowComputeProfile) UnsetGpuComputeMode()`

UnsetGpuComputeMode ensures that no value is present for GpuComputeMode, not even an explicit nil
### GetGpuDevicesRequest

`func (o *WorkflowComputeProfile) GetGpuDevicesRequest() int32`

GetGpuDevicesRequest returns the GpuDevicesRequest field if non-nil, zero value otherwise.

### GetGpuDevicesRequestOk

`func (o *WorkflowComputeProfile) GetGpuDevicesRequestOk() (*int32, bool)`

GetGpuDevicesRequestOk returns a tuple with the GpuDevicesRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuDevicesRequest

`func (o *WorkflowComputeProfile) SetGpuDevicesRequest(v int32)`

SetGpuDevicesRequest sets GpuDevicesRequest field to given value.

### HasGpuDevicesRequest

`func (o *WorkflowComputeProfile) HasGpuDevicesRequest() bool`

HasGpuDevicesRequest returns a boolean if a field has been set.

### SetGpuDevicesRequestNil

`func (o *WorkflowComputeProfile) SetGpuDevicesRequestNil(b bool)`

 SetGpuDevicesRequestNil sets the value for GpuDevicesRequest to be an explicit nil

### UnsetGpuDevicesRequest
`func (o *WorkflowComputeProfile) UnsetGpuDevicesRequest()`

UnsetGpuDevicesRequest ensures that no value is present for GpuDevicesRequest, not even an explicit nil
### GetGpuMemoryLimit

`func (o *WorkflowComputeProfile) GetGpuMemoryLimit() string`

GetGpuMemoryLimit returns the GpuMemoryLimit field if non-nil, zero value otherwise.

### GetGpuMemoryLimitOk

`func (o *WorkflowComputeProfile) GetGpuMemoryLimitOk() (*string, bool)`

GetGpuMemoryLimitOk returns a tuple with the GpuMemoryLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryLimit

`func (o *WorkflowComputeProfile) SetGpuMemoryLimit(v string)`

SetGpuMemoryLimit sets GpuMemoryLimit field to given value.

### HasGpuMemoryLimit

`func (o *WorkflowComputeProfile) HasGpuMemoryLimit() bool`

HasGpuMemoryLimit returns a boolean if a field has been set.

### SetGpuMemoryLimitNil

`func (o *WorkflowComputeProfile) SetGpuMemoryLimitNil(b bool)`

 SetGpuMemoryLimitNil sets the value for GpuMemoryLimit to be an explicit nil

### UnsetGpuMemoryLimit
`func (o *WorkflowComputeProfile) UnsetGpuMemoryLimit()`

UnsetGpuMemoryLimit ensures that no value is present for GpuMemoryLimit, not even an explicit nil
### GetGpuMemoryPortionLimit

`func (o *WorkflowComputeProfile) GetGpuMemoryPortionLimit() float64`

GetGpuMemoryPortionLimit returns the GpuMemoryPortionLimit field if non-nil, zero value otherwise.

### GetGpuMemoryPortionLimitOk

`func (o *WorkflowComputeProfile) GetGpuMemoryPortionLimitOk() (*float64, bool)`

GetGpuMemoryPortionLimitOk returns a tuple with the GpuMemoryPortionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryPortionLimit

`func (o *WorkflowComputeProfile) SetGpuMemoryPortionLimit(v float64)`

SetGpuMemoryPortionLimit sets GpuMemoryPortionLimit field to given value.

### HasGpuMemoryPortionLimit

`func (o *WorkflowComputeProfile) HasGpuMemoryPortionLimit() bool`

HasGpuMemoryPortionLimit returns a boolean if a field has been set.

### SetGpuMemoryPortionLimitNil

`func (o *WorkflowComputeProfile) SetGpuMemoryPortionLimitNil(b bool)`

 SetGpuMemoryPortionLimitNil sets the value for GpuMemoryPortionLimit to be an explicit nil

### UnsetGpuMemoryPortionLimit
`func (o *WorkflowComputeProfile) UnsetGpuMemoryPortionLimit()`

UnsetGpuMemoryPortionLimit ensures that no value is present for GpuMemoryPortionLimit, not even an explicit nil
### GetGpuMemoryPortionRequest

`func (o *WorkflowComputeProfile) GetGpuMemoryPortionRequest() float64`

GetGpuMemoryPortionRequest returns the GpuMemoryPortionRequest field if non-nil, zero value otherwise.

### GetGpuMemoryPortionRequestOk

`func (o *WorkflowComputeProfile) GetGpuMemoryPortionRequestOk() (*float64, bool)`

GetGpuMemoryPortionRequestOk returns a tuple with the GpuMemoryPortionRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryPortionRequest

`func (o *WorkflowComputeProfile) SetGpuMemoryPortionRequest(v float64)`

SetGpuMemoryPortionRequest sets GpuMemoryPortionRequest field to given value.

### HasGpuMemoryPortionRequest

`func (o *WorkflowComputeProfile) HasGpuMemoryPortionRequest() bool`

HasGpuMemoryPortionRequest returns a boolean if a field has been set.

### SetGpuMemoryPortionRequestNil

`func (o *WorkflowComputeProfile) SetGpuMemoryPortionRequestNil(b bool)`

 SetGpuMemoryPortionRequestNil sets the value for GpuMemoryPortionRequest to be an explicit nil

### UnsetGpuMemoryPortionRequest
`func (o *WorkflowComputeProfile) UnsetGpuMemoryPortionRequest()`

UnsetGpuMemoryPortionRequest ensures that no value is present for GpuMemoryPortionRequest, not even an explicit nil
### GetGpuMemoryRequest

`func (o *WorkflowComputeProfile) GetGpuMemoryRequest() string`

GetGpuMemoryRequest returns the GpuMemoryRequest field if non-nil, zero value otherwise.

### GetGpuMemoryRequestOk

`func (o *WorkflowComputeProfile) GetGpuMemoryRequestOk() (*string, bool)`

GetGpuMemoryRequestOk returns a tuple with the GpuMemoryRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryRequest

`func (o *WorkflowComputeProfile) SetGpuMemoryRequest(v string)`

SetGpuMemoryRequest sets GpuMemoryRequest field to given value.

### HasGpuMemoryRequest

`func (o *WorkflowComputeProfile) HasGpuMemoryRequest() bool`

HasGpuMemoryRequest returns a boolean if a field has been set.

### SetGpuMemoryRequestNil

`func (o *WorkflowComputeProfile) SetGpuMemoryRequestNil(b bool)`

 SetGpuMemoryRequestNil sets the value for GpuMemoryRequest to be an explicit nil

### UnsetGpuMemoryRequest
`func (o *WorkflowComputeProfile) UnsetGpuMemoryRequest()`

UnsetGpuMemoryRequest ensures that no value is present for GpuMemoryRequest, not even an explicit nil
### GetGpuRequestType

`func (o *WorkflowComputeProfile) GetGpuRequestType() GpuRequestType`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *WorkflowComputeProfile) GetGpuRequestTypeOk() (*GpuRequestType, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *WorkflowComputeProfile) SetGpuRequestType(v GpuRequestType)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *WorkflowComputeProfile) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *WorkflowComputeProfile) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *WorkflowComputeProfile) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetLargeShmRequest

`func (o *WorkflowComputeProfile) GetLargeShmRequest() bool`

GetLargeShmRequest returns the LargeShmRequest field if non-nil, zero value otherwise.

### GetLargeShmRequestOk

`func (o *WorkflowComputeProfile) GetLargeShmRequestOk() (*bool, bool)`

GetLargeShmRequestOk returns a tuple with the LargeShmRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLargeShmRequest

`func (o *WorkflowComputeProfile) SetLargeShmRequest(v bool)`

SetLargeShmRequest sets LargeShmRequest field to given value.

### HasLargeShmRequest

`func (o *WorkflowComputeProfile) HasLargeShmRequest() bool`

HasLargeShmRequest returns a boolean if a field has been set.

### SetLargeShmRequestNil

`func (o *WorkflowComputeProfile) SetLargeShmRequestNil(b bool)`

 SetLargeShmRequestNil sets the value for LargeShmRequest to be an explicit nil

### UnsetLargeShmRequest
`func (o *WorkflowComputeProfile) UnsetLargeShmRequest()`

UnsetLargeShmRequest ensures that no value is present for LargeShmRequest, not even an explicit nil
### GetName

`func (o *WorkflowComputeProfile) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowComputeProfile) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowComputeProfile) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowComputeProfile) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowComputeProfile) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowComputeProfile) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


