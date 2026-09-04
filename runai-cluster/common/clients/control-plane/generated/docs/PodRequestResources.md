# PodRequestResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuRequestType** | Pointer to [**NullableGpuRequestType**](GpuRequestType.md) |  | [optional] 
**Gpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**GpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 
**Cpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**CpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 
**MigProfile** | Pointer to [**NullableMigProfile**](MigProfile.md) |  | [optional] 
**ExtendedResources** | Pointer to [**[]WorkloadsExtendedResource**](WorkloadsExtendedResource.md) | Set of extended resources with their quantity | [optional] 

## Methods

### NewPodRequestResources

`func NewPodRequestResources() *PodRequestResources`

NewPodRequestResources instantiates a new PodRequestResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodRequestResourcesWithDefaults

`func NewPodRequestResourcesWithDefaults() *PodRequestResources`

NewPodRequestResourcesWithDefaults instantiates a new PodRequestResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuRequestType

`func (o *PodRequestResources) GetGpuRequestType() GpuRequestType`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *PodRequestResources) GetGpuRequestTypeOk() (*GpuRequestType, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *PodRequestResources) SetGpuRequestType(v GpuRequestType)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *PodRequestResources) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *PodRequestResources) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *PodRequestResources) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetGpu

`func (o *PodRequestResources) GetGpu() RequestResourceCores`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *PodRequestResources) GetGpuOk() (*RequestResourceCores, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *PodRequestResources) SetGpu(v RequestResourceCores)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *PodRequestResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *PodRequestResources) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *PodRequestResources) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetGpuMemory

`func (o *PodRequestResources) GetGpuMemory() RequestResourceQuantity`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *PodRequestResources) GetGpuMemoryOk() (*RequestResourceQuantity, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *PodRequestResources) SetGpuMemory(v RequestResourceQuantity)`

SetGpuMemory sets GpuMemory field to given value.

### HasGpuMemory

`func (o *PodRequestResources) HasGpuMemory() bool`

HasGpuMemory returns a boolean if a field has been set.

### SetGpuMemoryNil

`func (o *PodRequestResources) SetGpuMemoryNil(b bool)`

 SetGpuMemoryNil sets the value for GpuMemory to be an explicit nil

### UnsetGpuMemory
`func (o *PodRequestResources) UnsetGpuMemory()`

UnsetGpuMemory ensures that no value is present for GpuMemory, not even an explicit nil
### GetCpu

`func (o *PodRequestResources) GetCpu() RequestResourceCores`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *PodRequestResources) GetCpuOk() (*RequestResourceCores, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *PodRequestResources) SetCpu(v RequestResourceCores)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *PodRequestResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *PodRequestResources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *PodRequestResources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetCpuMemory

`func (o *PodRequestResources) GetCpuMemory() RequestResourceQuantity`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *PodRequestResources) GetCpuMemoryOk() (*RequestResourceQuantity, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *PodRequestResources) SetCpuMemory(v RequestResourceQuantity)`

SetCpuMemory sets CpuMemory field to given value.

### HasCpuMemory

`func (o *PodRequestResources) HasCpuMemory() bool`

HasCpuMemory returns a boolean if a field has been set.

### SetCpuMemoryNil

`func (o *PodRequestResources) SetCpuMemoryNil(b bool)`

 SetCpuMemoryNil sets the value for CpuMemory to be an explicit nil

### UnsetCpuMemory
`func (o *PodRequestResources) UnsetCpuMemory()`

UnsetCpuMemory ensures that no value is present for CpuMemory, not even an explicit nil
### GetMigProfile

`func (o *PodRequestResources) GetMigProfile() MigProfile`

GetMigProfile returns the MigProfile field if non-nil, zero value otherwise.

### GetMigProfileOk

`func (o *PodRequestResources) GetMigProfileOk() (*MigProfile, bool)`

GetMigProfileOk returns a tuple with the MigProfile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMigProfile

`func (o *PodRequestResources) SetMigProfile(v MigProfile)`

SetMigProfile sets MigProfile field to given value.

### HasMigProfile

`func (o *PodRequestResources) HasMigProfile() bool`

HasMigProfile returns a boolean if a field has been set.

### SetMigProfileNil

`func (o *PodRequestResources) SetMigProfileNil(b bool)`

 SetMigProfileNil sets the value for MigProfile to be an explicit nil

### UnsetMigProfile
`func (o *PodRequestResources) UnsetMigProfile()`

UnsetMigProfile ensures that no value is present for MigProfile, not even an explicit nil
### GetExtendedResources

`func (o *PodRequestResources) GetExtendedResources() []WorkloadsExtendedResource`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *PodRequestResources) GetExtendedResourcesOk() (*[]WorkloadsExtendedResource, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *PodRequestResources) SetExtendedResources(v []WorkloadsExtendedResource)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *PodRequestResources) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *PodRequestResources) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *PodRequestResources) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


