# JobResourcesData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | [**JobResourceData**](JobResourceData.md) |  | 
**GpuMemory** | [**JobResourceData**](JobResourceData.md) |  | 
**Cpu** | [**JobResourceData**](JobResourceData.md) |  | 
**CpuMemory** | [**JobResourceData**](JobResourceData.md) |  | 
**Advanced** | Pointer to [**JobAdvancedData**](JobAdvancedData.md) |  | [optional] 

## Methods

### NewJobResourcesData

`func NewJobResourcesData(gpu JobResourceData, gpuMemory JobResourceData, cpu JobResourceData, cpuMemory JobResourceData, ) *JobResourcesData`

NewJobResourcesData instantiates a new JobResourcesData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobResourcesDataWithDefaults

`func NewJobResourcesDataWithDefaults() *JobResourcesData`

NewJobResourcesDataWithDefaults instantiates a new JobResourcesData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *JobResourcesData) GetGpu() JobResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *JobResourcesData) GetGpuOk() (*JobResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *JobResourcesData) SetGpu(v JobResourceData)`

SetGpu sets Gpu field to given value.


### GetGpuMemory

`func (o *JobResourcesData) GetGpuMemory() JobResourceData`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *JobResourcesData) GetGpuMemoryOk() (*JobResourceData, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *JobResourcesData) SetGpuMemory(v JobResourceData)`

SetGpuMemory sets GpuMemory field to given value.


### GetCpu

`func (o *JobResourcesData) GetCpu() JobResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *JobResourcesData) GetCpuOk() (*JobResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *JobResourcesData) SetCpu(v JobResourceData)`

SetCpu sets Cpu field to given value.


### GetCpuMemory

`func (o *JobResourcesData) GetCpuMemory() JobResourceData`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *JobResourcesData) GetCpuMemoryOk() (*JobResourceData, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *JobResourcesData) SetCpuMemory(v JobResourceData)`

SetCpuMemory sets CpuMemory field to given value.


### GetAdvanced

`func (o *JobResourcesData) GetAdvanced() JobAdvancedData`

GetAdvanced returns the Advanced field if non-nil, zero value otherwise.

### GetAdvancedOk

`func (o *JobResourcesData) GetAdvancedOk() (*JobAdvancedData, bool)`

GetAdvancedOk returns a tuple with the Advanced field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdvanced

`func (o *JobResourcesData) SetAdvanced(v JobAdvancedData)`

SetAdvanced sets Advanced field to given value.

### HasAdvanced

`func (o *JobResourcesData) HasAdvanced() bool`

HasAdvanced returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


