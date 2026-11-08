# RecommendationResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuType** | **string** | Exact GPU type reported by the cluster inventory. Different memory capacities for the same GPU type produce separate results. | 
**GpuMemoryCapacityBytes** | Pointer to **int64** | Reported memory capacity of each GPU, in bytes. Omitted when capacity is unknown. Does not represent free memory or a quota guarantee. | [optional] 
**Status** | [**RecommendationStatus**](RecommendationStatus.md) | Whether an allocation was recommended for this GPU type and memory capacity. | 
**MinimumGpuMemoryPerDeviceGb** | Pointer to **float64** | Published minimum memory per GPU for the selected NGC profile, in decimal GB (1 GB &#x3D; 1,000,000,000 bytes), before allocation rounding. Omitted when no profile is selected. | [optional] 
**ProfileId** | Pointer to **string** | Identifier of the selected published NGC execution profile. Omitted when no profile is selected. | [optional] 
**Precision** | Pointer to **string** | Model precision reported by the selected NGC profile. Omitted when no profile is selected or the profile does not report precision. | [optional] 
**InferenceBackend** | Pointer to **string** | Inference backend reported by the selected NGC profile. Omitted when no profile is selected or the profile does not report a backend. | [optional] 
**TensorParallelSize** | Pointer to **int64** | Number of GPUs participating in tensor parallelism in the selected NGC profile. Omitted when no profile is selected. | [optional] 
**PipelineParallelSize** | Pointer to **int64** | Number of pipeline parallel stages in the selected NGC profile. Omitted when no profile is selected. | [optional] 
**Reason** | Pointer to **string** | Explanation of why no allocation could be recommended. Omitted for a recommended allocation. | [optional] 
**Recommendation** | Pointer to [**NullableRecommendation**](Recommendation.md) | Recommended allocation for this result. Absent or null when no allocation could be recommended. | [optional] 

## Methods

### NewRecommendationResult

`func NewRecommendationResult(gpuType string, status RecommendationStatus, ) *RecommendationResult`

NewRecommendationResult instantiates a new RecommendationResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecommendationResultWithDefaults

`func NewRecommendationResultWithDefaults() *RecommendationResult`

NewRecommendationResultWithDefaults instantiates a new RecommendationResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuType

`func (o *RecommendationResult) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *RecommendationResult) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *RecommendationResult) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.


### GetGpuMemoryCapacityBytes

`func (o *RecommendationResult) GetGpuMemoryCapacityBytes() int64`

GetGpuMemoryCapacityBytes returns the GpuMemoryCapacityBytes field if non-nil, zero value otherwise.

### GetGpuMemoryCapacityBytesOk

`func (o *RecommendationResult) GetGpuMemoryCapacityBytesOk() (*int64, bool)`

GetGpuMemoryCapacityBytesOk returns a tuple with the GpuMemoryCapacityBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryCapacityBytes

`func (o *RecommendationResult) SetGpuMemoryCapacityBytes(v int64)`

SetGpuMemoryCapacityBytes sets GpuMemoryCapacityBytes field to given value.

### HasGpuMemoryCapacityBytes

`func (o *RecommendationResult) HasGpuMemoryCapacityBytes() bool`

HasGpuMemoryCapacityBytes returns a boolean if a field has been set.

### GetStatus

`func (o *RecommendationResult) GetStatus() RecommendationStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RecommendationResult) GetStatusOk() (*RecommendationStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RecommendationResult) SetStatus(v RecommendationStatus)`

SetStatus sets Status field to given value.


### GetMinimumGpuMemoryPerDeviceGb

`func (o *RecommendationResult) GetMinimumGpuMemoryPerDeviceGb() float64`

GetMinimumGpuMemoryPerDeviceGb returns the MinimumGpuMemoryPerDeviceGb field if non-nil, zero value otherwise.

### GetMinimumGpuMemoryPerDeviceGbOk

`func (o *RecommendationResult) GetMinimumGpuMemoryPerDeviceGbOk() (*float64, bool)`

GetMinimumGpuMemoryPerDeviceGbOk returns a tuple with the MinimumGpuMemoryPerDeviceGb field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumGpuMemoryPerDeviceGb

`func (o *RecommendationResult) SetMinimumGpuMemoryPerDeviceGb(v float64)`

SetMinimumGpuMemoryPerDeviceGb sets MinimumGpuMemoryPerDeviceGb field to given value.

### HasMinimumGpuMemoryPerDeviceGb

`func (o *RecommendationResult) HasMinimumGpuMemoryPerDeviceGb() bool`

HasMinimumGpuMemoryPerDeviceGb returns a boolean if a field has been set.

### GetProfileId

`func (o *RecommendationResult) GetProfileId() string`

GetProfileId returns the ProfileId field if non-nil, zero value otherwise.

### GetProfileIdOk

`func (o *RecommendationResult) GetProfileIdOk() (*string, bool)`

GetProfileIdOk returns a tuple with the ProfileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProfileId

`func (o *RecommendationResult) SetProfileId(v string)`

SetProfileId sets ProfileId field to given value.

### HasProfileId

`func (o *RecommendationResult) HasProfileId() bool`

HasProfileId returns a boolean if a field has been set.

### GetPrecision

`func (o *RecommendationResult) GetPrecision() string`

GetPrecision returns the Precision field if non-nil, zero value otherwise.

### GetPrecisionOk

`func (o *RecommendationResult) GetPrecisionOk() (*string, bool)`

GetPrecisionOk returns a tuple with the Precision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrecision

`func (o *RecommendationResult) SetPrecision(v string)`

SetPrecision sets Precision field to given value.

### HasPrecision

`func (o *RecommendationResult) HasPrecision() bool`

HasPrecision returns a boolean if a field has been set.

### GetInferenceBackend

`func (o *RecommendationResult) GetInferenceBackend() string`

GetInferenceBackend returns the InferenceBackend field if non-nil, zero value otherwise.

### GetInferenceBackendOk

`func (o *RecommendationResult) GetInferenceBackendOk() (*string, bool)`

GetInferenceBackendOk returns a tuple with the InferenceBackend field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceBackend

`func (o *RecommendationResult) SetInferenceBackend(v string)`

SetInferenceBackend sets InferenceBackend field to given value.

### HasInferenceBackend

`func (o *RecommendationResult) HasInferenceBackend() bool`

HasInferenceBackend returns a boolean if a field has been set.

### GetTensorParallelSize

`func (o *RecommendationResult) GetTensorParallelSize() int64`

GetTensorParallelSize returns the TensorParallelSize field if non-nil, zero value otherwise.

### GetTensorParallelSizeOk

`func (o *RecommendationResult) GetTensorParallelSizeOk() (*int64, bool)`

GetTensorParallelSizeOk returns a tuple with the TensorParallelSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTensorParallelSize

`func (o *RecommendationResult) SetTensorParallelSize(v int64)`

SetTensorParallelSize sets TensorParallelSize field to given value.

### HasTensorParallelSize

`func (o *RecommendationResult) HasTensorParallelSize() bool`

HasTensorParallelSize returns a boolean if a field has been set.

### GetPipelineParallelSize

`func (o *RecommendationResult) GetPipelineParallelSize() int64`

GetPipelineParallelSize returns the PipelineParallelSize field if non-nil, zero value otherwise.

### GetPipelineParallelSizeOk

`func (o *RecommendationResult) GetPipelineParallelSizeOk() (*int64, bool)`

GetPipelineParallelSizeOk returns a tuple with the PipelineParallelSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipelineParallelSize

`func (o *RecommendationResult) SetPipelineParallelSize(v int64)`

SetPipelineParallelSize sets PipelineParallelSize field to given value.

### HasPipelineParallelSize

`func (o *RecommendationResult) HasPipelineParallelSize() bool`

HasPipelineParallelSize returns a boolean if a field has been set.

### GetReason

`func (o *RecommendationResult) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *RecommendationResult) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *RecommendationResult) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *RecommendationResult) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRecommendation

`func (o *RecommendationResult) GetRecommendation() Recommendation`

GetRecommendation returns the Recommendation field if non-nil, zero value otherwise.

### GetRecommendationOk

`func (o *RecommendationResult) GetRecommendationOk() (*Recommendation, bool)`

GetRecommendationOk returns a tuple with the Recommendation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecommendation

`func (o *RecommendationResult) SetRecommendation(v Recommendation)`

SetRecommendation sets Recommendation field to given value.

### HasRecommendation

`func (o *RecommendationResult) HasRecommendation() bool`

HasRecommendation returns a boolean if a field has been set.

### SetRecommendationNil

`func (o *RecommendationResult) SetRecommendationNil(b bool)`

 SetRecommendationNil sets the value for Recommendation to be an explicit nil

### UnsetRecommendation
`func (o *RecommendationResult) UnsetRecommendation()`

UnsetRecommendation ensures that no value is present for Recommendation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


