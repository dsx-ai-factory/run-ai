# ClusterCurrentProjectResourcesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**ResourceData**](ResourceData.md) |  | 
**Cpu** | [**ResourceData**](ResourceData.md) |  | 
**Memory** | [**ResourceData**](ResourceData.md) |  | 
**ProjectName** | **string** |  | 
**DepartmentName** | **string** |  | 
**NodepoolName** | **string** |  | 

## Methods

### NewClusterCurrentProjectResourcesInner

`func NewClusterCurrentProjectResourcesInner(numberOfPendingWorkloads int32, gpu ResourceData, cpu ResourceData, memory ResourceData, projectName string, departmentName string, nodepoolName string, ) *ClusterCurrentProjectResourcesInner`

NewClusterCurrentProjectResourcesInner instantiates a new ClusterCurrentProjectResourcesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCurrentProjectResourcesInnerWithDefaults

`func NewClusterCurrentProjectResourcesInnerWithDefaults() *ClusterCurrentProjectResourcesInner`

NewClusterCurrentProjectResourcesInnerWithDefaults instantiates a new ClusterCurrentProjectResourcesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *ClusterCurrentProjectResourcesInner) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *ClusterCurrentProjectResourcesInner) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *ClusterCurrentProjectResourcesInner) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *ClusterCurrentProjectResourcesInner) GetGpu() ResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ClusterCurrentProjectResourcesInner) GetGpuOk() (*ResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ClusterCurrentProjectResourcesInner) SetGpu(v ResourceData)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *ClusterCurrentProjectResourcesInner) GetCpu() ResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ClusterCurrentProjectResourcesInner) GetCpuOk() (*ResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ClusterCurrentProjectResourcesInner) SetCpu(v ResourceData)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *ClusterCurrentProjectResourcesInner) GetMemory() ResourceData`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ClusterCurrentProjectResourcesInner) GetMemoryOk() (*ResourceData, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ClusterCurrentProjectResourcesInner) SetMemory(v ResourceData)`

SetMemory sets Memory field to given value.


### GetProjectName

`func (o *ClusterCurrentProjectResourcesInner) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *ClusterCurrentProjectResourcesInner) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *ClusterCurrentProjectResourcesInner) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetDepartmentName

`func (o *ClusterCurrentProjectResourcesInner) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *ClusterCurrentProjectResourcesInner) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *ClusterCurrentProjectResourcesInner) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.


### GetNodepoolName

`func (o *ClusterCurrentProjectResourcesInner) GetNodepoolName() string`

GetNodepoolName returns the NodepoolName field if non-nil, zero value otherwise.

### GetNodepoolNameOk

`func (o *ClusterCurrentProjectResourcesInner) GetNodepoolNameOk() (*string, bool)`

GetNodepoolNameOk returns a tuple with the NodepoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepoolName

`func (o *ClusterCurrentProjectResourcesInner) SetNodepoolName(v string)`

SetNodepoolName sets NodepoolName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


