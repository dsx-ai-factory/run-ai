# RightSizingResponseResultsInnerRecommendation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllocationType** | **string** |  | 
**GpuCount** | **int64** |  | 
**GpuMemoryPerDeviceBytes** | Pointer to **int64** |  | [optional] 
**Estimated** | **bool** | Whether GPU count is a naive estimate. False describes allocation arithmetic only; it does not verify runtime or hardware compatibility. | 
**Warning** | Pointer to **string** | Conditional sizing prerequisites, unverified hardware/runtime compatibility, and limitations of an estimated count. Read before using the allocation. | [optional] 

## Methods

### NewRightSizingResponseResultsInnerRecommendation

`func NewRightSizingResponseResultsInnerRecommendation(allocationType string, gpuCount int64, estimated bool, ) *RightSizingResponseResultsInnerRecommendation`

NewRightSizingResponseResultsInnerRecommendation instantiates a new RightSizingResponseResultsInnerRecommendation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRightSizingResponseResultsInnerRecommendationWithDefaults

`func NewRightSizingResponseResultsInnerRecommendationWithDefaults() *RightSizingResponseResultsInnerRecommendation`

NewRightSizingResponseResultsInnerRecommendationWithDefaults instantiates a new RightSizingResponseResultsInnerRecommendation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocationType

`func (o *RightSizingResponseResultsInnerRecommendation) GetAllocationType() string`

GetAllocationType returns the AllocationType field if non-nil, zero value otherwise.

### GetAllocationTypeOk

`func (o *RightSizingResponseResultsInnerRecommendation) GetAllocationTypeOk() (*string, bool)`

GetAllocationTypeOk returns a tuple with the AllocationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocationType

`func (o *RightSizingResponseResultsInnerRecommendation) SetAllocationType(v string)`

SetAllocationType sets AllocationType field to given value.


### GetGpuCount

`func (o *RightSizingResponseResultsInnerRecommendation) GetGpuCount() int64`

GetGpuCount returns the GpuCount field if non-nil, zero value otherwise.

### GetGpuCountOk

`func (o *RightSizingResponseResultsInnerRecommendation) GetGpuCountOk() (*int64, bool)`

GetGpuCountOk returns a tuple with the GpuCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuCount

`func (o *RightSizingResponseResultsInnerRecommendation) SetGpuCount(v int64)`

SetGpuCount sets GpuCount field to given value.


### GetGpuMemoryPerDeviceBytes

`func (o *RightSizingResponseResultsInnerRecommendation) GetGpuMemoryPerDeviceBytes() int64`

GetGpuMemoryPerDeviceBytes returns the GpuMemoryPerDeviceBytes field if non-nil, zero value otherwise.

### GetGpuMemoryPerDeviceBytesOk

`func (o *RightSizingResponseResultsInnerRecommendation) GetGpuMemoryPerDeviceBytesOk() (*int64, bool)`

GetGpuMemoryPerDeviceBytesOk returns a tuple with the GpuMemoryPerDeviceBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryPerDeviceBytes

`func (o *RightSizingResponseResultsInnerRecommendation) SetGpuMemoryPerDeviceBytes(v int64)`

SetGpuMemoryPerDeviceBytes sets GpuMemoryPerDeviceBytes field to given value.

### HasGpuMemoryPerDeviceBytes

`func (o *RightSizingResponseResultsInnerRecommendation) HasGpuMemoryPerDeviceBytes() bool`

HasGpuMemoryPerDeviceBytes returns a boolean if a field has been set.

### GetEstimated

`func (o *RightSizingResponseResultsInnerRecommendation) GetEstimated() bool`

GetEstimated returns the Estimated field if non-nil, zero value otherwise.

### GetEstimatedOk

`func (o *RightSizingResponseResultsInnerRecommendation) GetEstimatedOk() (*bool, bool)`

GetEstimatedOk returns a tuple with the Estimated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEstimated

`func (o *RightSizingResponseResultsInnerRecommendation) SetEstimated(v bool)`

SetEstimated sets Estimated field to given value.


### GetWarning

`func (o *RightSizingResponseResultsInnerRecommendation) GetWarning() string`

GetWarning returns the Warning field if non-nil, zero value otherwise.

### GetWarningOk

`func (o *RightSizingResponseResultsInnerRecommendation) GetWarningOk() (*string, bool)`

GetWarningOk returns a tuple with the Warning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWarning

`func (o *RightSizingResponseResultsInnerRecommendation) SetWarning(v string)`

SetWarning sets Warning field to given value.

### HasWarning

`func (o *RightSizingResponseResultsInnerRecommendation) HasWarning() bool`

HasWarning returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


