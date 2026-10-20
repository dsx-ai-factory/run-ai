# GetWorkloadActionsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | [**[]WorkloadActionRecord**](WorkloadActionRecord.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetWorkloadActionsResponse

`func NewGetWorkloadActionsResponse(actions []WorkloadActionRecord, ) *GetWorkloadActionsResponse`

NewGetWorkloadActionsResponse instantiates a new GetWorkloadActionsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadActionsResponseWithDefaults

`func NewGetWorkloadActionsResponseWithDefaults() *GetWorkloadActionsResponse`

NewGetWorkloadActionsResponseWithDefaults instantiates a new GetWorkloadActionsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *GetWorkloadActionsResponse) GetActions() []WorkloadActionRecord`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *GetWorkloadActionsResponse) GetActionsOk() (*[]WorkloadActionRecord, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *GetWorkloadActionsResponse) SetActions(v []WorkloadActionRecord)`

SetActions sets Actions field to given value.


### GetNext

`func (o *GetWorkloadActionsResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloadActionsResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloadActionsResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloadActionsResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


