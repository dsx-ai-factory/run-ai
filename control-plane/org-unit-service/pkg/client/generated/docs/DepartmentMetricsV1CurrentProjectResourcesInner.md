# DepartmentMetricsV1CurrentProjectResourcesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Cpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Memory** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**ProjectName** | **string** |  | 
**NodepoolName** | **string** |  | 

## Methods

### NewDepartmentMetricsV1CurrentProjectResourcesInner

`func NewDepartmentMetricsV1CurrentProjectResourcesInner(numberOfPendingWorkloads int32, gpu MetricsResourceDataV1, cpu MetricsResourceDataV1, memory MetricsResourceDataV1, projectName string, nodepoolName string, ) *DepartmentMetricsV1CurrentProjectResourcesInner`

NewDepartmentMetricsV1CurrentProjectResourcesInner instantiates a new DepartmentMetricsV1CurrentProjectResourcesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentMetricsV1CurrentProjectResourcesInnerWithDefaults

`func NewDepartmentMetricsV1CurrentProjectResourcesInnerWithDefaults() *DepartmentMetricsV1CurrentProjectResourcesInner`

NewDepartmentMetricsV1CurrentProjectResourcesInnerWithDefaults instantiates a new DepartmentMetricsV1CurrentProjectResourcesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetGpu() MetricsResourceDataV1`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetGpuOk() (*MetricsResourceDataV1, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetGpu(v MetricsResourceDataV1)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetCpu() MetricsResourceDataV1`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetCpuOk() (*MetricsResourceDataV1, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetCpu(v MetricsResourceDataV1)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetMemory() MetricsResourceDataV1`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetMemoryOk() (*MetricsResourceDataV1, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetMemory(v MetricsResourceDataV1)`

SetMemory sets Memory field to given value.


### GetProjectName

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetNodepoolName

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetNodepoolName() string`

GetNodepoolName returns the NodepoolName field if non-nil, zero value otherwise.

### GetNodepoolNameOk

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) GetNodepoolNameOk() (*string, bool)`

GetNodepoolNameOk returns a tuple with the NodepoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepoolName

`func (o *DepartmentMetricsV1CurrentProjectResourcesInner) SetNodepoolName(v string)`

SetNodepoolName sets NodepoolName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


