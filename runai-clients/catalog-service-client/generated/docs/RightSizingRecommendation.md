# RightSizingRecommendation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllocationType** | [**RightSizingAllocationType**](RightSizingAllocationType.md) |  | 
**GpuCount** | **int64** |  | 
**GpuMemoryPerDeviceBytes** | Pointer to **int64** |  | [optional] 

## Methods

### NewRightSizingRecommendation

`func NewRightSizingRecommendation(allocationType RightSizingAllocationType, gpuCount int64, ) *RightSizingRecommendation`

NewRightSizingRecommendation instantiates a new RightSizingRecommendation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRightSizingRecommendationWithDefaults

`func NewRightSizingRecommendationWithDefaults() *RightSizingRecommendation`

NewRightSizingRecommendationWithDefaults instantiates a new RightSizingRecommendation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocationType

`func (o *RightSizingRecommendation) GetAllocationType() RightSizingAllocationType`

GetAllocationType returns the AllocationType field if non-nil, zero value otherwise.

### GetAllocationTypeOk

`func (o *RightSizingRecommendation) GetAllocationTypeOk() (*RightSizingAllocationType, bool)`

GetAllocationTypeOk returns a tuple with the AllocationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocationType

`func (o *RightSizingRecommendation) SetAllocationType(v RightSizingAllocationType)`

SetAllocationType sets AllocationType field to given value.


### GetGpuCount

`func (o *RightSizingRecommendation) GetGpuCount() int64`

GetGpuCount returns the GpuCount field if non-nil, zero value otherwise.

### GetGpuCountOk

`func (o *RightSizingRecommendation) GetGpuCountOk() (*int64, bool)`

GetGpuCountOk returns a tuple with the GpuCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuCount

`func (o *RightSizingRecommendation) SetGpuCount(v int64)`

SetGpuCount sets GpuCount field to given value.


### GetGpuMemoryPerDeviceBytes

`func (o *RightSizingRecommendation) GetGpuMemoryPerDeviceBytes() int64`

GetGpuMemoryPerDeviceBytes returns the GpuMemoryPerDeviceBytes field if non-nil, zero value otherwise.

### GetGpuMemoryPerDeviceBytesOk

`func (o *RightSizingRecommendation) GetGpuMemoryPerDeviceBytesOk() (*int64, bool)`

GetGpuMemoryPerDeviceBytesOk returns a tuple with the GpuMemoryPerDeviceBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryPerDeviceBytes

`func (o *RightSizingRecommendation) SetGpuMemoryPerDeviceBytes(v int64)`

SetGpuMemoryPerDeviceBytes sets GpuMemoryPerDeviceBytes field to given value.

### HasGpuMemoryPerDeviceBytes

`func (o *RightSizingRecommendation) HasGpuMemoryPerDeviceBytes() bool`

HasGpuMemoryPerDeviceBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


