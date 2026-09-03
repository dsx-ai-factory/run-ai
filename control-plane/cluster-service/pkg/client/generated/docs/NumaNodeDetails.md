# NumaNodeDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The NUMA node zone name reported by NodeResourceTopology. | 
**TotalCpuCores** | **float64** | Total CPU in cores available for allocation, taken from the allocatable value reported by NodeResourceTopology. | 
**AllocatedCpuCores** | **float64** | Allocated CPU in cores, calculated as allocatable minus available. | 
**TotalCpuMemoryBytes** | **int64** | Total CPU memory in bytes available for allocation, taken from the allocatable value reported by NodeResourceTopology. | 
**AllocatedCpuMemoryBytes** | **int64** | Allocated memory in bytes, calculated as allocatable minus available. | 
**TotalGpus** | **int32** | Total GPUs available for allocation, taken from the allocatable value reported by NodeResourceTopology. | 
**AllocatedGpus** | **int32** | Allocated GPUs, calculated as allocatable minus available. | 

## Methods

### NewNumaNodeDetails

`func NewNumaNodeDetails(name string, totalCpuCores float64, allocatedCpuCores float64, totalCpuMemoryBytes int64, allocatedCpuMemoryBytes int64, totalGpus int32, allocatedGpus int32, ) *NumaNodeDetails`

NewNumaNodeDetails instantiates a new NumaNodeDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNumaNodeDetailsWithDefaults

`func NewNumaNodeDetailsWithDefaults() *NumaNodeDetails`

NewNumaNodeDetailsWithDefaults instantiates a new NumaNodeDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NumaNodeDetails) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NumaNodeDetails) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NumaNodeDetails) SetName(v string)`

SetName sets Name field to given value.


### GetTotalCpuCores

`func (o *NumaNodeDetails) GetTotalCpuCores() float64`

GetTotalCpuCores returns the TotalCpuCores field if non-nil, zero value otherwise.

### GetTotalCpuCoresOk

`func (o *NumaNodeDetails) GetTotalCpuCoresOk() (*float64, bool)`

GetTotalCpuCoresOk returns a tuple with the TotalCpuCores field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCpuCores

`func (o *NumaNodeDetails) SetTotalCpuCores(v float64)`

SetTotalCpuCores sets TotalCpuCores field to given value.


### GetAllocatedCpuCores

`func (o *NumaNodeDetails) GetAllocatedCpuCores() float64`

GetAllocatedCpuCores returns the AllocatedCpuCores field if non-nil, zero value otherwise.

### GetAllocatedCpuCoresOk

`func (o *NumaNodeDetails) GetAllocatedCpuCoresOk() (*float64, bool)`

GetAllocatedCpuCoresOk returns a tuple with the AllocatedCpuCores field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedCpuCores

`func (o *NumaNodeDetails) SetAllocatedCpuCores(v float64)`

SetAllocatedCpuCores sets AllocatedCpuCores field to given value.


### GetTotalCpuMemoryBytes

`func (o *NumaNodeDetails) GetTotalCpuMemoryBytes() int64`

GetTotalCpuMemoryBytes returns the TotalCpuMemoryBytes field if non-nil, zero value otherwise.

### GetTotalCpuMemoryBytesOk

`func (o *NumaNodeDetails) GetTotalCpuMemoryBytesOk() (*int64, bool)`

GetTotalCpuMemoryBytesOk returns a tuple with the TotalCpuMemoryBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCpuMemoryBytes

`func (o *NumaNodeDetails) SetTotalCpuMemoryBytes(v int64)`

SetTotalCpuMemoryBytes sets TotalCpuMemoryBytes field to given value.


### GetAllocatedCpuMemoryBytes

`func (o *NumaNodeDetails) GetAllocatedCpuMemoryBytes() int64`

GetAllocatedCpuMemoryBytes returns the AllocatedCpuMemoryBytes field if non-nil, zero value otherwise.

### GetAllocatedCpuMemoryBytesOk

`func (o *NumaNodeDetails) GetAllocatedCpuMemoryBytesOk() (*int64, bool)`

GetAllocatedCpuMemoryBytesOk returns a tuple with the AllocatedCpuMemoryBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedCpuMemoryBytes

`func (o *NumaNodeDetails) SetAllocatedCpuMemoryBytes(v int64)`

SetAllocatedCpuMemoryBytes sets AllocatedCpuMemoryBytes field to given value.


### GetTotalGpus

`func (o *NumaNodeDetails) GetTotalGpus() int32`

GetTotalGpus returns the TotalGpus field if non-nil, zero value otherwise.

### GetTotalGpusOk

`func (o *NumaNodeDetails) GetTotalGpusOk() (*int32, bool)`

GetTotalGpusOk returns a tuple with the TotalGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalGpus

`func (o *NumaNodeDetails) SetTotalGpus(v int32)`

SetTotalGpus sets TotalGpus field to given value.


### GetAllocatedGpus

`func (o *NumaNodeDetails) GetAllocatedGpus() int32`

GetAllocatedGpus returns the AllocatedGpus field if non-nil, zero value otherwise.

### GetAllocatedGpusOk

`func (o *NumaNodeDetails) GetAllocatedGpusOk() (*int32, bool)`

GetAllocatedGpusOk returns a tuple with the AllocatedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedGpus

`func (o *NumaNodeDetails) SetAllocatedGpus(v int32)`

SetAllocatedGpus sets AllocatedGpus field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


