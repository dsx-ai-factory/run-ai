# RightSizingResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuType** | **string** |  | 
**GpuMemoryCapacityBytes** | Pointer to **int64** | Reported per-device capacity, not free memory or a quota guarantee. | [optional] 
**Status** | **string** |  | 
**MinimumGpuMemoryPerDeviceGb** | Pointer to **float64** | Published NGC minimum for the source profile, before allocation rounding. | [optional] 
**ProfileId** | Pointer to **string** |  | [optional] 
**Precision** | Pointer to **string** |  | [optional] 
**InferenceBackend** | Pointer to **string** |  | [optional] 
**TensorParallelSize** | Pointer to **int64** | Explicit tensor parallel size from the selected published profile. | [optional] 
**PipelineParallelSize** | Pointer to **int64** | Explicit pipeline parallel size from the selected published profile. | [optional] 
**Reason** | Pointer to **string** | Explanation when sizing cannot be recommended. | [optional] 
**Recommendation** | Pointer to [**NullableRightSizingRecommendation**](RightSizingRecommendation.md) |  | [optional] 

## Methods

### NewRightSizingResult

`func NewRightSizingResult(gpuType string, status string, ) *RightSizingResult`

NewRightSizingResult instantiates a new RightSizingResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRightSizingResultWithDefaults

`func NewRightSizingResultWithDefaults() *RightSizingResult`

NewRightSizingResultWithDefaults instantiates a new RightSizingResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuType

`func (o *RightSizingResult) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *RightSizingResult) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *RightSizingResult) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.


### GetGpuMemoryCapacityBytes

`func (o *RightSizingResult) GetGpuMemoryCapacityBytes() int64`

GetGpuMemoryCapacityBytes returns the GpuMemoryCapacityBytes field if non-nil, zero value otherwise.

### GetGpuMemoryCapacityBytesOk

`func (o *RightSizingResult) GetGpuMemoryCapacityBytesOk() (*int64, bool)`

GetGpuMemoryCapacityBytesOk returns a tuple with the GpuMemoryCapacityBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryCapacityBytes

`func (o *RightSizingResult) SetGpuMemoryCapacityBytes(v int64)`

SetGpuMemoryCapacityBytes sets GpuMemoryCapacityBytes field to given value.

### HasGpuMemoryCapacityBytes

`func (o *RightSizingResult) HasGpuMemoryCapacityBytes() bool`

HasGpuMemoryCapacityBytes returns a boolean if a field has been set.

### GetStatus

`func (o *RightSizingResult) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RightSizingResult) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RightSizingResult) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResult) GetMinimumGpuMemoryPerDeviceGb() float64`

GetMinimumGpuMemoryPerDeviceGb returns the MinimumGpuMemoryPerDeviceGb field if non-nil, zero value otherwise.

### GetMinimumGpuMemoryPerDeviceGbOk

`func (o *RightSizingResult) GetMinimumGpuMemoryPerDeviceGbOk() (*float64, bool)`

GetMinimumGpuMemoryPerDeviceGbOk returns a tuple with the MinimumGpuMemoryPerDeviceGb field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResult) SetMinimumGpuMemoryPerDeviceGb(v float64)`

SetMinimumGpuMemoryPerDeviceGb sets MinimumGpuMemoryPerDeviceGb field to given value.

### HasMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResult) HasMinimumGpuMemoryPerDeviceGb() bool`

HasMinimumGpuMemoryPerDeviceGb returns a boolean if a field has been set.

### GetProfileId

`func (o *RightSizingResult) GetProfileId() string`

GetProfileId returns the ProfileId field if non-nil, zero value otherwise.

### GetProfileIdOk

`func (o *RightSizingResult) GetProfileIdOk() (*string, bool)`

GetProfileIdOk returns a tuple with the ProfileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProfileId

`func (o *RightSizingResult) SetProfileId(v string)`

SetProfileId sets ProfileId field to given value.

### HasProfileId

`func (o *RightSizingResult) HasProfileId() bool`

HasProfileId returns a boolean if a field has been set.

### GetPrecision

`func (o *RightSizingResult) GetPrecision() string`

GetPrecision returns the Precision field if non-nil, zero value otherwise.

### GetPrecisionOk

`func (o *RightSizingResult) GetPrecisionOk() (*string, bool)`

GetPrecisionOk returns a tuple with the Precision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrecision

`func (o *RightSizingResult) SetPrecision(v string)`

SetPrecision sets Precision field to given value.

### HasPrecision

`func (o *RightSizingResult) HasPrecision() bool`

HasPrecision returns a boolean if a field has been set.

### GetInferenceBackend

`func (o *RightSizingResult) GetInferenceBackend() string`

GetInferenceBackend returns the InferenceBackend field if non-nil, zero value otherwise.

### GetInferenceBackendOk

`func (o *RightSizingResult) GetInferenceBackendOk() (*string, bool)`

GetInferenceBackendOk returns a tuple with the InferenceBackend field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceBackend

`func (o *RightSizingResult) SetInferenceBackend(v string)`

SetInferenceBackend sets InferenceBackend field to given value.

### HasInferenceBackend

`func (o *RightSizingResult) HasInferenceBackend() bool`

HasInferenceBackend returns a boolean if a field has been set.

### GetTensorParallelSize

`func (o *RightSizingResult) GetTensorParallelSize() int64`

GetTensorParallelSize returns the TensorParallelSize field if non-nil, zero value otherwise.

### GetTensorParallelSizeOk

`func (o *RightSizingResult) GetTensorParallelSizeOk() (*int64, bool)`

GetTensorParallelSizeOk returns a tuple with the TensorParallelSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTensorParallelSize

`func (o *RightSizingResult) SetTensorParallelSize(v int64)`

SetTensorParallelSize sets TensorParallelSize field to given value.

### HasTensorParallelSize

`func (o *RightSizingResult) HasTensorParallelSize() bool`

HasTensorParallelSize returns a boolean if a field has been set.

### GetPipelineParallelSize

`func (o *RightSizingResult) GetPipelineParallelSize() int64`

GetPipelineParallelSize returns the PipelineParallelSize field if non-nil, zero value otherwise.

### GetPipelineParallelSizeOk

`func (o *RightSizingResult) GetPipelineParallelSizeOk() (*int64, bool)`

GetPipelineParallelSizeOk returns a tuple with the PipelineParallelSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipelineParallelSize

`func (o *RightSizingResult) SetPipelineParallelSize(v int64)`

SetPipelineParallelSize sets PipelineParallelSize field to given value.

### HasPipelineParallelSize

`func (o *RightSizingResult) HasPipelineParallelSize() bool`

HasPipelineParallelSize returns a boolean if a field has been set.

### GetReason

`func (o *RightSizingResult) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *RightSizingResult) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *RightSizingResult) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *RightSizingResult) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRecommendation

`func (o *RightSizingResult) GetRecommendation() RightSizingRecommendation`

GetRecommendation returns the Recommendation field if non-nil, zero value otherwise.

### GetRecommendationOk

`func (o *RightSizingResult) GetRecommendationOk() (*RightSizingRecommendation, bool)`

GetRecommendationOk returns a tuple with the Recommendation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecommendation

`func (o *RightSizingResult) SetRecommendation(v RightSizingRecommendation)`

SetRecommendation sets Recommendation field to given value.

### HasRecommendation

`func (o *RightSizingResult) HasRecommendation() bool`

HasRecommendation returns a boolean if a field has been set.

### SetRecommendationNil

`func (o *RightSizingResult) SetRecommendationNil(b bool)`

 SetRecommendationNil sets the value for Recommendation to be an explicit nil

### UnsetRecommendation
`func (o *RightSizingResult) UnsetRecommendation()`

UnsetRecommendation ensures that no value is present for Recommendation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


