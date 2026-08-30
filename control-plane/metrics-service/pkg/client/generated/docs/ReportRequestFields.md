# ReportRequestFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Range** | [**NullableRange**](Range.md) |  | 
**AdditionalFields** | Pointer to **map[string]interface{}** |  | [optional] 
**Metrics** | [**[]MetricInformationForReport**](MetricInformationForReport.md) |  | 

## Methods

### NewReportRequestFields

`func NewReportRequestFields(name string, range_ NullableRange, metrics []MetricInformationForReport, ) *ReportRequestFields`

NewReportRequestFields instantiates a new ReportRequestFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReportRequestFieldsWithDefaults

`func NewReportRequestFieldsWithDefaults() *ReportRequestFields`

NewReportRequestFieldsWithDefaults instantiates a new ReportRequestFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ReportRequestFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ReportRequestFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ReportRequestFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *ReportRequestFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ReportRequestFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ReportRequestFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ReportRequestFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetLabels

`func (o *ReportRequestFields) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ReportRequestFields) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ReportRequestFields) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ReportRequestFields) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetRange

`func (o *ReportRequestFields) GetRange() Range`

GetRange returns the Range field if non-nil, zero value otherwise.

### GetRangeOk

`func (o *ReportRequestFields) GetRangeOk() (*Range, bool)`

GetRangeOk returns a tuple with the Range field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRange

`func (o *ReportRequestFields) SetRange(v Range)`

SetRange sets Range field to given value.


### SetRangeNil

`func (o *ReportRequestFields) SetRangeNil(b bool)`

 SetRangeNil sets the value for Range to be an explicit nil

### UnsetRange
`func (o *ReportRequestFields) UnsetRange()`

UnsetRange ensures that no value is present for Range, not even an explicit nil
### GetAdditionalFields

`func (o *ReportRequestFields) GetAdditionalFields() map[string]interface{}`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *ReportRequestFields) GetAdditionalFieldsOk() (*map[string]interface{}, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *ReportRequestFields) SetAdditionalFields(v map[string]interface{})`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *ReportRequestFields) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.

### GetMetrics

`func (o *ReportRequestFields) GetMetrics() []MetricInformationForReport`

GetMetrics returns the Metrics field if non-nil, zero value otherwise.

### GetMetricsOk

`func (o *ReportRequestFields) GetMetricsOk() (*[]MetricInformationForReport, bool)`

GetMetricsOk returns a tuple with the Metrics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetrics

`func (o *ReportRequestFields) SetMetrics(v []MetricInformationForReport)`

SetMetrics sets Metrics field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


