# HistogramSeries

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | specifies what data returned | 
**Values** | [**[]HistogramValue**](HistogramValue.md) |  | 

## Methods

### NewHistogramSeries

`func NewHistogramSeries(type_ string, values []HistogramValue, ) *HistogramSeries`

NewHistogramSeries instantiates a new HistogramSeries object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistogramSeriesWithDefaults

`func NewHistogramSeriesWithDefaults() *HistogramSeries`

NewHistogramSeriesWithDefaults instantiates a new HistogramSeries object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *HistogramSeries) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *HistogramSeries) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *HistogramSeries) SetType(v string)`

SetType sets Type field to given value.


### GetValues

`func (o *HistogramSeries) GetValues() []HistogramValue`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *HistogramSeries) GetValuesOk() (*[]HistogramValue, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *HistogramSeries) SetValues(v []HistogramValue)`

SetValues sets Values field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


