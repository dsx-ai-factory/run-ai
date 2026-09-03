# Departments

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | [**[]DepartmentsDataInner**](DepartmentsDataInner.md) |  | 
**TimeRange** | Pointer to [**TimeRange**](TimeRange.md) |  | [optional] 

## Methods

### NewDepartments

`func NewDepartments(data []DepartmentsDataInner, ) *Departments`

NewDepartments instantiates a new Departments object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentsWithDefaults

`func NewDepartmentsWithDefaults() *Departments`

NewDepartmentsWithDefaults instantiates a new Departments object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *Departments) GetData() []DepartmentsDataInner`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *Departments) GetDataOk() (*[]DepartmentsDataInner, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *Departments) SetData(v []DepartmentsDataInner)`

SetData sets Data field to given value.


### GetTimeRange

`func (o *Departments) GetTimeRange() TimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Departments) GetTimeRangeOk() (*TimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Departments) SetTimeRange(v TimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Departments) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


