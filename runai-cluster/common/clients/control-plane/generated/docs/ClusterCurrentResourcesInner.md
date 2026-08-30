# ClusterCurrentResourcesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**ResourceData**](ResourceData.md) |  | 
**Cpu** | [**ResourceData**](ResourceData.md) |  | 
**Memory** | [**ResourceData**](ResourceData.md) |  | 
**NodepoolName** | **string** |  | 

## Methods

### NewClusterCurrentResourcesInner

`func NewClusterCurrentResourcesInner(numberOfPendingWorkloads int32, gpu ResourceData, cpu ResourceData, memory ResourceData, nodepoolName string, ) *ClusterCurrentResourcesInner`

NewClusterCurrentResourcesInner instantiates a new ClusterCurrentResourcesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCurrentResourcesInnerWithDefaults

`func NewClusterCurrentResourcesInnerWithDefaults() *ClusterCurrentResourcesInner`

NewClusterCurrentResourcesInnerWithDefaults instantiates a new ClusterCurrentResourcesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *ClusterCurrentResourcesInner) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *ClusterCurrentResourcesInner) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *ClusterCurrentResourcesInner) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *ClusterCurrentResourcesInner) GetGpu() ResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ClusterCurrentResourcesInner) GetGpuOk() (*ResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ClusterCurrentResourcesInner) SetGpu(v ResourceData)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *ClusterCurrentResourcesInner) GetCpu() ResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ClusterCurrentResourcesInner) GetCpuOk() (*ResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ClusterCurrentResourcesInner) SetCpu(v ResourceData)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *ClusterCurrentResourcesInner) GetMemory() ResourceData`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ClusterCurrentResourcesInner) GetMemoryOk() (*ResourceData, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ClusterCurrentResourcesInner) SetMemory(v ResourceData)`

SetMemory sets Memory field to given value.


### GetNodepoolName

`func (o *ClusterCurrentResourcesInner) GetNodepoolName() string`

GetNodepoolName returns the NodepoolName field if non-nil, zero value otherwise.

### GetNodepoolNameOk

`func (o *ClusterCurrentResourcesInner) GetNodepoolNameOk() (*string, bool)`

GetNodepoolNameOk returns a tuple with the NodepoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepoolName

`func (o *ClusterCurrentResourcesInner) SetNodepoolName(v string)`

SetNodepoolName sets NodepoolName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


