# GetWorkloadsSlimResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Workloads** | [**[]WorkloadSlim**](WorkloadSlim.md) |  | 

## Methods

### NewGetWorkloadsSlimResponse

`func NewGetWorkloadsSlimResponse(workloads []WorkloadSlim, ) *GetWorkloadsSlimResponse`

NewGetWorkloadsSlimResponse instantiates a new GetWorkloadsSlimResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadsSlimResponseWithDefaults

`func NewGetWorkloadsSlimResponseWithDefaults() *GetWorkloadsSlimResponse`

NewGetWorkloadsSlimResponseWithDefaults instantiates a new GetWorkloadsSlimResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetWorkloadsSlimResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloadsSlimResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloadsSlimResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloadsSlimResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetWorkloads

`func (o *GetWorkloadsSlimResponse) GetWorkloads() []WorkloadSlim`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *GetWorkloadsSlimResponse) GetWorkloadsOk() (*[]WorkloadSlim, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *GetWorkloadsSlimResponse) SetWorkloads(v []WorkloadSlim)`

SetWorkloads sets Workloads field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


