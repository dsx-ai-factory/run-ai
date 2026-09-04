# JobCurrent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TotalRunTime** | **float32** | Job total run time in seconds | 
**TotalWaitTime** | **float32** | Job total wait time in seconds | 
**Resources** | [**JobResourcesData**](JobResourcesData.md) |  | 

## Methods

### NewJobCurrent

`func NewJobCurrent(totalRunTime float32, totalWaitTime float32, resources JobResourcesData, ) *JobCurrent`

NewJobCurrent instantiates a new JobCurrent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobCurrentWithDefaults

`func NewJobCurrentWithDefaults() *JobCurrent`

NewJobCurrentWithDefaults instantiates a new JobCurrent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTotalRunTime

`func (o *JobCurrent) GetTotalRunTime() float32`

GetTotalRunTime returns the TotalRunTime field if non-nil, zero value otherwise.

### GetTotalRunTimeOk

`func (o *JobCurrent) GetTotalRunTimeOk() (*float32, bool)`

GetTotalRunTimeOk returns a tuple with the TotalRunTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRunTime

`func (o *JobCurrent) SetTotalRunTime(v float32)`

SetTotalRunTime sets TotalRunTime field to given value.


### GetTotalWaitTime

`func (o *JobCurrent) GetTotalWaitTime() float32`

GetTotalWaitTime returns the TotalWaitTime field if non-nil, zero value otherwise.

### GetTotalWaitTimeOk

`func (o *JobCurrent) GetTotalWaitTimeOk() (*float32, bool)`

GetTotalWaitTimeOk returns a tuple with the TotalWaitTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalWaitTime

`func (o *JobCurrent) SetTotalWaitTime(v float32)`

SetTotalWaitTime sets TotalWaitTime field to given value.


### GetResources

`func (o *JobCurrent) GetResources() JobResourcesData`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *JobCurrent) GetResourcesOk() (*JobResourcesData, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *JobCurrent) SetResources(v JobResourcesData)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


