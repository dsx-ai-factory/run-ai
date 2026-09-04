# MetricInformationForReport

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**DisplayName** | **string** |  | 
**Filters** | Pointer to **map[string][]string** |  | [optional] 
**Quantile** | Pointer to **NullableFloat32** | the quantile value to calculate | [optional] 

## Methods

### NewMetricInformationForReport

`func NewMetricInformationForReport(name string, displayName string, ) *MetricInformationForReport`

NewMetricInformationForReport instantiates a new MetricInformationForReport object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricInformationForReportWithDefaults

`func NewMetricInformationForReportWithDefaults() *MetricInformationForReport`

NewMetricInformationForReportWithDefaults instantiates a new MetricInformationForReport object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *MetricInformationForReport) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *MetricInformationForReport) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *MetricInformationForReport) SetName(v string)`

SetName sets Name field to given value.


### GetDisplayName

`func (o *MetricInformationForReport) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *MetricInformationForReport) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *MetricInformationForReport) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.


### GetFilters

`func (o *MetricInformationForReport) GetFilters() map[string][]string`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *MetricInformationForReport) GetFiltersOk() (*map[string][]string, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *MetricInformationForReport) SetFilters(v map[string][]string)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *MetricInformationForReport) HasFilters() bool`

HasFilters returns a boolean if a field has been set.

### GetQuantile

`func (o *MetricInformationForReport) GetQuantile() float32`

GetQuantile returns the Quantile field if non-nil, zero value otherwise.

### GetQuantileOk

`func (o *MetricInformationForReport) GetQuantileOk() (*float32, bool)`

GetQuantileOk returns a tuple with the Quantile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuantile

`func (o *MetricInformationForReport) SetQuantile(v float32)`

SetQuantile sets Quantile field to given value.

### HasQuantile

`func (o *MetricInformationForReport) HasQuantile() bool`

HasQuantile returns a boolean if a field has been set.

### SetQuantileNil

`func (o *MetricInformationForReport) SetQuantileNil(b bool)`

 SetQuantileNil sets the value for Quantile to be an explicit nil

### UnsetQuantile
`func (o *MetricInformationForReport) UnsetQuantile()`

UnsetQuantile ensures that no value is present for Quantile, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


