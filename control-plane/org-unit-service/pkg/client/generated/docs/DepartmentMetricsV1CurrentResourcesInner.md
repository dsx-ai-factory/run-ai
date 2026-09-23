# DepartmentMetricsV1CurrentResourcesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Cpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Memory** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**NodepoolName** | **string** |  | 

## Methods

### NewDepartmentMetricsV1CurrentResourcesInner

`func NewDepartmentMetricsV1CurrentResourcesInner(numberOfPendingWorkloads int32, gpu MetricsResourceDataV1, cpu MetricsResourceDataV1, memory MetricsResourceDataV1, nodepoolName string, ) *DepartmentMetricsV1CurrentResourcesInner`

NewDepartmentMetricsV1CurrentResourcesInner instantiates a new DepartmentMetricsV1CurrentResourcesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentMetricsV1CurrentResourcesInnerWithDefaults

`func NewDepartmentMetricsV1CurrentResourcesInnerWithDefaults() *DepartmentMetricsV1CurrentResourcesInner`

NewDepartmentMetricsV1CurrentResourcesInnerWithDefaults instantiates a new DepartmentMetricsV1CurrentResourcesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *DepartmentMetricsV1CurrentResourcesInner) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetGpu() MetricsResourceDataV1`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetGpuOk() (*MetricsResourceDataV1, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *DepartmentMetricsV1CurrentResourcesInner) SetGpu(v MetricsResourceDataV1)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetCpu() MetricsResourceDataV1`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetCpuOk() (*MetricsResourceDataV1, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *DepartmentMetricsV1CurrentResourcesInner) SetCpu(v MetricsResourceDataV1)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetMemory() MetricsResourceDataV1`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetMemoryOk() (*MetricsResourceDataV1, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *DepartmentMetricsV1CurrentResourcesInner) SetMemory(v MetricsResourceDataV1)`

SetMemory sets Memory field to given value.


### GetNodepoolName

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetNodepoolName() string`

GetNodepoolName returns the NodepoolName field if non-nil, zero value otherwise.

### GetNodepoolNameOk

`func (o *DepartmentMetricsV1CurrentResourcesInner) GetNodepoolNameOk() (*string, bool)`

GetNodepoolNameOk returns a tuple with the NodepoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepoolName

`func (o *DepartmentMetricsV1CurrentResourcesInner) SetNodepoolName(v string)`

SetNodepoolName sets NodepoolName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


