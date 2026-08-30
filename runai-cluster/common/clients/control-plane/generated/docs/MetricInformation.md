# MetricInformation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Filters** | Pointer to **map[string][]string** |  | [optional] 
**ResponseLabels** | Pointer to **[]string** |  | [optional] 
**Range** | Pointer to [**NullableMetricInformationRange**](MetricInformationRange.md) |  | [optional] 

## Methods

### NewMetricInformation

`func NewMetricInformation(name string, ) *MetricInformation`

NewMetricInformation instantiates a new MetricInformation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricInformationWithDefaults

`func NewMetricInformationWithDefaults() *MetricInformation`

NewMetricInformationWithDefaults instantiates a new MetricInformation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *MetricInformation) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *MetricInformation) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *MetricInformation) SetName(v string)`

SetName sets Name field to given value.


### GetFilters

`func (o *MetricInformation) GetFilters() map[string][]string`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *MetricInformation) GetFiltersOk() (*map[string][]string, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *MetricInformation) SetFilters(v map[string][]string)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *MetricInformation) HasFilters() bool`

HasFilters returns a boolean if a field has been set.

### GetResponseLabels

`func (o *MetricInformation) GetResponseLabels() []string`

GetResponseLabels returns the ResponseLabels field if non-nil, zero value otherwise.

### GetResponseLabelsOk

`func (o *MetricInformation) GetResponseLabelsOk() (*[]string, bool)`

GetResponseLabelsOk returns a tuple with the ResponseLabels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseLabels

`func (o *MetricInformation) SetResponseLabels(v []string)`

SetResponseLabels sets ResponseLabels field to given value.

### HasResponseLabels

`func (o *MetricInformation) HasResponseLabels() bool`

HasResponseLabels returns a boolean if a field has been set.

### GetRange

`func (o *MetricInformation) GetRange() MetricInformationRange`

GetRange returns the Range field if non-nil, zero value otherwise.

### GetRangeOk

`func (o *MetricInformation) GetRangeOk() (*MetricInformationRange, bool)`

GetRangeOk returns a tuple with the Range field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRange

`func (o *MetricInformation) SetRange(v MetricInformationRange)`

SetRange sets Range field to given value.

### HasRange

`func (o *MetricInformation) HasRange() bool`

HasRange returns a boolean if a field has been set.

### SetRangeNil

`func (o *MetricInformation) SetRangeNil(b bool)`

 SetRangeNil sets the value for Range to be an explicit nil

### UnsetRange
`func (o *MetricInformation) UnsetRange()`

UnsetRange ensures that no value is present for Range, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


