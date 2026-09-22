# MetricsCompositeResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Measurements** | [**[]MeasurementResponse**](MeasurementResponse.md) |  | 
**Histogram** | Pointer to [**[]HistogramSeries**](HistogramSeries.md) |  | [optional] 

## Methods

### NewMetricsCompositeResponse

`func NewMetricsCompositeResponse(measurements []MeasurementResponse, ) *MetricsCompositeResponse`

NewMetricsCompositeResponse instantiates a new MetricsCompositeResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricsCompositeResponseWithDefaults

`func NewMetricsCompositeResponseWithDefaults() *MetricsCompositeResponse`

NewMetricsCompositeResponseWithDefaults instantiates a new MetricsCompositeResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeasurements

`func (o *MetricsCompositeResponse) GetMeasurements() []MeasurementResponse`

GetMeasurements returns the Measurements field if non-nil, zero value otherwise.

### GetMeasurementsOk

`func (o *MetricsCompositeResponse) GetMeasurementsOk() (*[]MeasurementResponse, bool)`

GetMeasurementsOk returns a tuple with the Measurements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurements

`func (o *MetricsCompositeResponse) SetMeasurements(v []MeasurementResponse)`

SetMeasurements sets Measurements field to given value.


### GetHistogram

`func (o *MetricsCompositeResponse) GetHistogram() []HistogramSeries`

GetHistogram returns the Histogram field if non-nil, zero value otherwise.

### GetHistogramOk

`func (o *MetricsCompositeResponse) GetHistogramOk() (*[]HistogramSeries, bool)`

GetHistogramOk returns a tuple with the Histogram field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHistogram

`func (o *MetricsCompositeResponse) SetHistogram(v []HistogramSeries)`

SetHistogram sets Histogram field to given value.

### HasHistogram

`func (o *MetricsCompositeResponse) HasHistogram() bool`

HasHistogram returns a boolean if a field has been set.

### SetHistogramNil

`func (o *MetricsCompositeResponse) SetHistogramNil(b bool)`

 SetHistogramNil sets the value for Histogram to be an explicit nil

### UnsetHistogram
`func (o *MetricsCompositeResponse) UnsetHistogram()`

UnsetHistogram ensures that no value is present for Histogram, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


