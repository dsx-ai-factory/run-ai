# MetricInformationRange

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Start** | **time.Time** | timestamp from when to fetch data in UTC | 
**End** | **time.Time** | timestamp until when to fetch data in UTC | 
**NumOfSamples** | Pointer to **int32** | the number of samples to take in the time range from start to end. | [optional] [default to 20]

## Methods

### NewMetricInformationRange

`func NewMetricInformationRange(start time.Time, end time.Time, ) *MetricInformationRange`

NewMetricInformationRange instantiates a new MetricInformationRange object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricInformationRangeWithDefaults

`func NewMetricInformationRangeWithDefaults() *MetricInformationRange`

NewMetricInformationRangeWithDefaults instantiates a new MetricInformationRange object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStart

`func (o *MetricInformationRange) GetStart() time.Time`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *MetricInformationRange) GetStartOk() (*time.Time, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *MetricInformationRange) SetStart(v time.Time)`

SetStart sets Start field to given value.


### GetEnd

`func (o *MetricInformationRange) GetEnd() time.Time`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *MetricInformationRange) GetEndOk() (*time.Time, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *MetricInformationRange) SetEnd(v time.Time)`

SetEnd sets End field to given value.


### GetNumOfSamples

`func (o *MetricInformationRange) GetNumOfSamples() int32`

GetNumOfSamples returns the NumOfSamples field if non-nil, zero value otherwise.

### GetNumOfSamplesOk

`func (o *MetricInformationRange) GetNumOfSamplesOk() (*int32, bool)`

GetNumOfSamplesOk returns a tuple with the NumOfSamples field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumOfSamples

`func (o *MetricInformationRange) SetNumOfSamples(v int32)`

SetNumOfSamples sets NumOfSamples field to given value.

### HasNumOfSamples

`func (o *MetricInformationRange) HasNumOfSamples() bool`

HasNumOfSamples returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


