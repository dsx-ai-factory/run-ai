# Recommendation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllocationType** | [**RecommendationAllocationType**](RecommendationAllocationType.md) | Whether the allocation requests GPU memory (&#x60;gpu_memory&#x60;) or whole GPUs (&#x60;gpu_count&#x60;). | 
**GpuCount** | **int64** | Number of GPUs used by the allocation. Always 1 for &#x60;gpu_memory&#x60;; for &#x60;gpu_count&#x60;, equals the selected profile’s tensor parallel size multiplied by its pipeline parallel size. | 
**GpuMemoryPerDeviceBytes** | Pointer to **int64** | Recommended memory per GPU, in bytes. For &#x60;gpu_memory&#x60;, this is the published minimum rounded up to whole bytes. For &#x60;gpu_count&#x60;, this is the reported capacity of each GPU. | [optional] 

## Methods

### NewRecommendation

`func NewRecommendation(allocationType RecommendationAllocationType, gpuCount int64, ) *Recommendation`

NewRecommendation instantiates a new Recommendation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecommendationWithDefaults

`func NewRecommendationWithDefaults() *Recommendation`

NewRecommendationWithDefaults instantiates a new Recommendation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocationType

`func (o *Recommendation) GetAllocationType() RecommendationAllocationType`

GetAllocationType returns the AllocationType field if non-nil, zero value otherwise.

### GetAllocationTypeOk

`func (o *Recommendation) GetAllocationTypeOk() (*RecommendationAllocationType, bool)`

GetAllocationTypeOk returns a tuple with the AllocationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocationType

`func (o *Recommendation) SetAllocationType(v RecommendationAllocationType)`

SetAllocationType sets AllocationType field to given value.


### GetGpuCount

`func (o *Recommendation) GetGpuCount() int64`

GetGpuCount returns the GpuCount field if non-nil, zero value otherwise.

### GetGpuCountOk

`func (o *Recommendation) GetGpuCountOk() (*int64, bool)`

GetGpuCountOk returns a tuple with the GpuCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuCount

`func (o *Recommendation) SetGpuCount(v int64)`

SetGpuCount sets GpuCount field to given value.


### GetGpuMemoryPerDeviceBytes

`func (o *Recommendation) GetGpuMemoryPerDeviceBytes() int64`

GetGpuMemoryPerDeviceBytes returns the GpuMemoryPerDeviceBytes field if non-nil, zero value otherwise.

### GetGpuMemoryPerDeviceBytesOk

`func (o *Recommendation) GetGpuMemoryPerDeviceBytesOk() (*int64, bool)`

GetGpuMemoryPerDeviceBytesOk returns a tuple with the GpuMemoryPerDeviceBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryPerDeviceBytes

`func (o *Recommendation) SetGpuMemoryPerDeviceBytes(v int64)`

SetGpuMemoryPerDeviceBytes sets GpuMemoryPerDeviceBytes field to given value.

### HasGpuMemoryPerDeviceBytes

`func (o *Recommendation) HasGpuMemoryPerDeviceBytes() bool`

HasGpuMemoryPerDeviceBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


