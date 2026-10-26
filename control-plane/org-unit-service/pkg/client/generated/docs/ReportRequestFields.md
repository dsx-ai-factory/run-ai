# ReportRequestFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**Start** | **time.Time** | timestamp from when to fetch data in UTC | 
**End** | **time.Time** | timestamp until when to fetch data in UTC | 
**GroupBy** | Pointer to [**NullableReportGroupByOptions**](ReportGroupByOptions.md) |  | [optional] 
**FilterBy** | Pointer to **[]string** | Filter results by a parameter. Use the format field-name &#x3D;&#x3D; value. | [optional] [default to []]

## Methods

### NewReportRequestFields

`func NewReportRequestFields(name string, start time.Time, end time.Time, ) *ReportRequestFields`

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

### GetStart

`func (o *ReportRequestFields) GetStart() time.Time`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *ReportRequestFields) GetStartOk() (*time.Time, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *ReportRequestFields) SetStart(v time.Time)`

SetStart sets Start field to given value.


### GetEnd

`func (o *ReportRequestFields) GetEnd() time.Time`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *ReportRequestFields) GetEndOk() (*time.Time, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *ReportRequestFields) SetEnd(v time.Time)`

SetEnd sets End field to given value.


### GetGroupBy

`func (o *ReportRequestFields) GetGroupBy() ReportGroupByOptions`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ReportRequestFields) GetGroupByOk() (*ReportGroupByOptions, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ReportRequestFields) SetGroupBy(v ReportGroupByOptions)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ReportRequestFields) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.

### SetGroupByNil

`func (o *ReportRequestFields) SetGroupByNil(b bool)`

 SetGroupByNil sets the value for GroupBy to be an explicit nil

### UnsetGroupBy
`func (o *ReportRequestFields) UnsetGroupBy()`

UnsetGroupBy ensures that no value is present for GroupBy, not even an explicit nil
### GetFilterBy

`func (o *ReportRequestFields) GetFilterBy() []string`

GetFilterBy returns the FilterBy field if non-nil, zero value otherwise.

### GetFilterByOk

`func (o *ReportRequestFields) GetFilterByOk() (*[]string, bool)`

GetFilterByOk returns a tuple with the FilterBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilterBy

`func (o *ReportRequestFields) SetFilterBy(v []string)`

SetFilterBy sets FilterBy field to given value.

### HasFilterBy

`func (o *ReportRequestFields) HasFilterBy() bool`

HasFilterBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


