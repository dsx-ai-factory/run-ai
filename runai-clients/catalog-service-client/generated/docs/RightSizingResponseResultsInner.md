# RightSizingResponseResultsInner

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
**Reason** | Pointer to **string** | Explanation when sizing cannot be recommended. | [optional] 
**Recommendation** | Pointer to [**NullableRightSizingResponseResultsInnerRecommendation**](RightSizingResponseResultsInnerRecommendation.md) |  | [optional] 

## Methods

### NewRightSizingResponseResultsInner

`func NewRightSizingResponseResultsInner(gpuType string, status string, ) *RightSizingResponseResultsInner`

NewRightSizingResponseResultsInner instantiates a new RightSizingResponseResultsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRightSizingResponseResultsInnerWithDefaults

`func NewRightSizingResponseResultsInnerWithDefaults() *RightSizingResponseResultsInner`

NewRightSizingResponseResultsInnerWithDefaults instantiates a new RightSizingResponseResultsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuType

`func (o *RightSizingResponseResultsInner) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *RightSizingResponseResultsInner) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *RightSizingResponseResultsInner) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.


### GetGpuMemoryCapacityBytes

`func (o *RightSizingResponseResultsInner) GetGpuMemoryCapacityBytes() int64`

GetGpuMemoryCapacityBytes returns the GpuMemoryCapacityBytes field if non-nil, zero value otherwise.

### GetGpuMemoryCapacityBytesOk

`func (o *RightSizingResponseResultsInner) GetGpuMemoryCapacityBytesOk() (*int64, bool)`

GetGpuMemoryCapacityBytesOk returns a tuple with the GpuMemoryCapacityBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryCapacityBytes

`func (o *RightSizingResponseResultsInner) SetGpuMemoryCapacityBytes(v int64)`

SetGpuMemoryCapacityBytes sets GpuMemoryCapacityBytes field to given value.

### HasGpuMemoryCapacityBytes

`func (o *RightSizingResponseResultsInner) HasGpuMemoryCapacityBytes() bool`

HasGpuMemoryCapacityBytes returns a boolean if a field has been set.

### GetStatus

`func (o *RightSizingResponseResultsInner) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RightSizingResponseResultsInner) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RightSizingResponseResultsInner) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResponseResultsInner) GetMinimumGpuMemoryPerDeviceGb() float64`

GetMinimumGpuMemoryPerDeviceGb returns the MinimumGpuMemoryPerDeviceGb field if non-nil, zero value otherwise.

### GetMinimumGpuMemoryPerDeviceGbOk

`func (o *RightSizingResponseResultsInner) GetMinimumGpuMemoryPerDeviceGbOk() (*float64, bool)`

GetMinimumGpuMemoryPerDeviceGbOk returns a tuple with the MinimumGpuMemoryPerDeviceGb field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResponseResultsInner) SetMinimumGpuMemoryPerDeviceGb(v float64)`

SetMinimumGpuMemoryPerDeviceGb sets MinimumGpuMemoryPerDeviceGb field to given value.

### HasMinimumGpuMemoryPerDeviceGb

`func (o *RightSizingResponseResultsInner) HasMinimumGpuMemoryPerDeviceGb() bool`

HasMinimumGpuMemoryPerDeviceGb returns a boolean if a field has been set.

### GetProfileId

`func (o *RightSizingResponseResultsInner) GetProfileId() string`

GetProfileId returns the ProfileId field if non-nil, zero value otherwise.

### GetProfileIdOk

`func (o *RightSizingResponseResultsInner) GetProfileIdOk() (*string, bool)`

GetProfileIdOk returns a tuple with the ProfileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProfileId

`func (o *RightSizingResponseResultsInner) SetProfileId(v string)`

SetProfileId sets ProfileId field to given value.

### HasProfileId

`func (o *RightSizingResponseResultsInner) HasProfileId() bool`

HasProfileId returns a boolean if a field has been set.

### GetPrecision

`func (o *RightSizingResponseResultsInner) GetPrecision() string`

GetPrecision returns the Precision field if non-nil, zero value otherwise.

### GetPrecisionOk

`func (o *RightSizingResponseResultsInner) GetPrecisionOk() (*string, bool)`

GetPrecisionOk returns a tuple with the Precision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrecision

`func (o *RightSizingResponseResultsInner) SetPrecision(v string)`

SetPrecision sets Precision field to given value.

### HasPrecision

`func (o *RightSizingResponseResultsInner) HasPrecision() bool`

HasPrecision returns a boolean if a field has been set.

### GetInferenceBackend

`func (o *RightSizingResponseResultsInner) GetInferenceBackend() string`

GetInferenceBackend returns the InferenceBackend field if non-nil, zero value otherwise.

### GetInferenceBackendOk

`func (o *RightSizingResponseResultsInner) GetInferenceBackendOk() (*string, bool)`

GetInferenceBackendOk returns a tuple with the InferenceBackend field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceBackend

`func (o *RightSizingResponseResultsInner) SetInferenceBackend(v string)`

SetInferenceBackend sets InferenceBackend field to given value.

### HasInferenceBackend

`func (o *RightSizingResponseResultsInner) HasInferenceBackend() bool`

HasInferenceBackend returns a boolean if a field has been set.

### GetReason

`func (o *RightSizingResponseResultsInner) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *RightSizingResponseResultsInner) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *RightSizingResponseResultsInner) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *RightSizingResponseResultsInner) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRecommendation

`func (o *RightSizingResponseResultsInner) GetRecommendation() RightSizingResponseResultsInnerRecommendation`

GetRecommendation returns the Recommendation field if non-nil, zero value otherwise.

### GetRecommendationOk

`func (o *RightSizingResponseResultsInner) GetRecommendationOk() (*RightSizingResponseResultsInnerRecommendation, bool)`

GetRecommendationOk returns a tuple with the Recommendation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecommendation

`func (o *RightSizingResponseResultsInner) SetRecommendation(v RightSizingResponseResultsInnerRecommendation)`

SetRecommendation sets Recommendation field to given value.

### HasRecommendation

`func (o *RightSizingResponseResultsInner) HasRecommendation() bool`

HasRecommendation returns a boolean if a field has been set.

### SetRecommendationNil

`func (o *RightSizingResponseResultsInner) SetRecommendationNil(b bool)`

 SetRecommendationNil sets the value for Recommendation to be an explicit nil

### UnsetRecommendation
`func (o *RightSizingResponseResultsInner) UnsetRecommendation()`

UnsetRecommendation ensures that no value is present for Recommendation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


