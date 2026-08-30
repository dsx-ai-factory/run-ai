# Result

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MetricLabels** | **map[string]string** |  | 
**Measurements** | [**[]Measurement**](Measurement.md) |  | 

## Methods

### NewResult

`func NewResult(metricLabels map[string]string, measurements []Measurement, ) *Result`

NewResult instantiates a new Result object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResultWithDefaults

`func NewResultWithDefaults() *Result`

NewResultWithDefaults instantiates a new Result object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetricLabels

`func (o *Result) GetMetricLabels() map[string]string`

GetMetricLabels returns the MetricLabels field if non-nil, zero value otherwise.

### GetMetricLabelsOk

`func (o *Result) GetMetricLabelsOk() (*map[string]string, bool)`

GetMetricLabelsOk returns a tuple with the MetricLabels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricLabels

`func (o *Result) SetMetricLabels(v map[string]string)`

SetMetricLabels sets MetricLabels field to given value.


### GetMeasurements

`func (o *Result) GetMeasurements() []Measurement`

GetMeasurements returns the Measurements field if non-nil, zero value otherwise.

### GetMeasurementsOk

`func (o *Result) GetMeasurementsOk() (*[]Measurement, bool)`

GetMeasurementsOk returns a tuple with the Measurements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurements

`func (o *Result) SetMeasurements(v []Measurement)`

SetMeasurements sets Measurements field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


