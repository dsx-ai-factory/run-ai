# ProjectsMetricsV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | [**[]ProjectsMetricsV1ResponseDataInner**](ProjectsMetricsV1ResponseDataInner.md) |  | 
**TimeRange** | Pointer to [**TimeRangeV1**](TimeRangeV1.md) |  | [optional] 

## Methods

### NewProjectsMetricsV1Response

`func NewProjectsMetricsV1Response(data []ProjectsMetricsV1ResponseDataInner, ) *ProjectsMetricsV1Response`

NewProjectsMetricsV1Response instantiates a new ProjectsMetricsV1Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectsMetricsV1ResponseWithDefaults

`func NewProjectsMetricsV1ResponseWithDefaults() *ProjectsMetricsV1Response`

NewProjectsMetricsV1ResponseWithDefaults instantiates a new ProjectsMetricsV1Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *ProjectsMetricsV1Response) GetData() []ProjectsMetricsV1ResponseDataInner`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *ProjectsMetricsV1Response) GetDataOk() (*[]ProjectsMetricsV1ResponseDataInner, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *ProjectsMetricsV1Response) SetData(v []ProjectsMetricsV1ResponseDataInner)`

SetData sets Data field to given value.


### GetTimeRange

`func (o *ProjectsMetricsV1Response) GetTimeRange() TimeRangeV1`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *ProjectsMetricsV1Response) GetTimeRangeOk() (*TimeRangeV1, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *ProjectsMetricsV1Response) SetTimeRange(v TimeRangeV1)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *ProjectsMetricsV1Response) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


