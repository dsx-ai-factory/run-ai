# Department1CurrentProjectResourcesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**ResourceData**](ResourceData.md) |  | 
**Cpu** | [**ResourceData**](ResourceData.md) |  | 
**Memory** | [**ResourceData**](ResourceData.md) |  | 
**ProjectName** | **string** |  | 
**NodepoolName** | **string** |  | 

## Methods

### NewDepartment1CurrentProjectResourcesInner

`func NewDepartment1CurrentProjectResourcesInner(numberOfPendingWorkloads int32, gpu ResourceData, cpu ResourceData, memory ResourceData, projectName string, nodepoolName string, ) *Department1CurrentProjectResourcesInner`

NewDepartment1CurrentProjectResourcesInner instantiates a new Department1CurrentProjectResourcesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartment1CurrentProjectResourcesInnerWithDefaults

`func NewDepartment1CurrentProjectResourcesInnerWithDefaults() *Department1CurrentProjectResourcesInner`

NewDepartment1CurrentProjectResourcesInnerWithDefaults instantiates a new Department1CurrentProjectResourcesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *Department1CurrentProjectResourcesInner) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *Department1CurrentProjectResourcesInner) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *Department1CurrentProjectResourcesInner) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *Department1CurrentProjectResourcesInner) GetGpu() ResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *Department1CurrentProjectResourcesInner) GetGpuOk() (*ResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *Department1CurrentProjectResourcesInner) SetGpu(v ResourceData)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *Department1CurrentProjectResourcesInner) GetCpu() ResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *Department1CurrentProjectResourcesInner) GetCpuOk() (*ResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *Department1CurrentProjectResourcesInner) SetCpu(v ResourceData)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *Department1CurrentProjectResourcesInner) GetMemory() ResourceData`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *Department1CurrentProjectResourcesInner) GetMemoryOk() (*ResourceData, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *Department1CurrentProjectResourcesInner) SetMemory(v ResourceData)`

SetMemory sets Memory field to given value.


### GetProjectName

`func (o *Department1CurrentProjectResourcesInner) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *Department1CurrentProjectResourcesInner) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *Department1CurrentProjectResourcesInner) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetNodepoolName

`func (o *Department1CurrentProjectResourcesInner) GetNodepoolName() string`

GetNodepoolName returns the NodepoolName field if non-nil, zero value otherwise.

### GetNodepoolNameOk

`func (o *Department1CurrentProjectResourcesInner) GetNodepoolNameOk() (*string, bool)`

GetNodepoolNameOk returns a tuple with the NodepoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepoolName

`func (o *Department1CurrentProjectResourcesInner) SetNodepoolName(v string)`

SetNodepoolName sets NodepoolName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


