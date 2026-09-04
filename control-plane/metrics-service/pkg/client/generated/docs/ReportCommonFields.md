# ReportCommonFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Range** | [**NullableRange**](Range.md) |  | 
**AdditionalFields** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewReportCommonFields

`func NewReportCommonFields(name string, range_ NullableRange, ) *ReportCommonFields`

NewReportCommonFields instantiates a new ReportCommonFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReportCommonFieldsWithDefaults

`func NewReportCommonFieldsWithDefaults() *ReportCommonFields`

NewReportCommonFieldsWithDefaults instantiates a new ReportCommonFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ReportCommonFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ReportCommonFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ReportCommonFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *ReportCommonFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ReportCommonFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ReportCommonFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ReportCommonFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetLabels

`func (o *ReportCommonFields) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ReportCommonFields) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ReportCommonFields) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ReportCommonFields) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetRange

`func (o *ReportCommonFields) GetRange() Range`

GetRange returns the Range field if non-nil, zero value otherwise.

### GetRangeOk

`func (o *ReportCommonFields) GetRangeOk() (*Range, bool)`

GetRangeOk returns a tuple with the Range field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRange

`func (o *ReportCommonFields) SetRange(v Range)`

SetRange sets Range field to given value.


### SetRangeNil

`func (o *ReportCommonFields) SetRangeNil(b bool)`

 SetRangeNil sets the value for Range to be an explicit nil

### UnsetRange
`func (o *ReportCommonFields) UnsetRange()`

UnsetRange ensures that no value is present for Range, not even an explicit nil
### GetAdditionalFields

`func (o *ReportCommonFields) GetAdditionalFields() map[string]interface{}`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *ReportCommonFields) GetAdditionalFieldsOk() (*map[string]interface{}, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *ReportCommonFields) SetAdditionalFields(v map[string]interface{})`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *ReportCommonFields) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


