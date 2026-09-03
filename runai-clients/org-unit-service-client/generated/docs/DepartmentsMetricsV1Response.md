# DepartmentsMetricsV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | [**[]DepartmentsMetricsV1ResponseDataInner**](DepartmentsMetricsV1ResponseDataInner.md) |  | 
**TimeRange** | Pointer to [**TimeRangeV1**](TimeRangeV1.md) |  | [optional] 

## Methods

### NewDepartmentsMetricsV1Response

`func NewDepartmentsMetricsV1Response(data []DepartmentsMetricsV1ResponseDataInner, ) *DepartmentsMetricsV1Response`

NewDepartmentsMetricsV1Response instantiates a new DepartmentsMetricsV1Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentsMetricsV1ResponseWithDefaults

`func NewDepartmentsMetricsV1ResponseWithDefaults() *DepartmentsMetricsV1Response`

NewDepartmentsMetricsV1ResponseWithDefaults instantiates a new DepartmentsMetricsV1Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *DepartmentsMetricsV1Response) GetData() []DepartmentsMetricsV1ResponseDataInner`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *DepartmentsMetricsV1Response) GetDataOk() (*[]DepartmentsMetricsV1ResponseDataInner, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *DepartmentsMetricsV1Response) SetData(v []DepartmentsMetricsV1ResponseDataInner)`

SetData sets Data field to given value.


### GetTimeRange

`func (o *DepartmentsMetricsV1Response) GetTimeRange() TimeRangeV1`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *DepartmentsMetricsV1Response) GetTimeRangeOk() (*TimeRangeV1, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *DepartmentsMetricsV1Response) SetTimeRange(v TimeRangeV1)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *DepartmentsMetricsV1Response) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


