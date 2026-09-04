# GetWorkloads200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Workloads** | [**[]Workload**](Workload.md) |  | 

## Methods

### NewGetWorkloads200Response

`func NewGetWorkloads200Response(workloads []Workload, ) *GetWorkloads200Response`

NewGetWorkloads200Response instantiates a new GetWorkloads200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloads200ResponseWithDefaults

`func NewGetWorkloads200ResponseWithDefaults() *GetWorkloads200Response`

NewGetWorkloads200ResponseWithDefaults instantiates a new GetWorkloads200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetWorkloads200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloads200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloads200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloads200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetWorkloads

`func (o *GetWorkloads200Response) GetWorkloads() []Workload`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *GetWorkloads200Response) GetWorkloadsOk() (*[]Workload, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *GetWorkloads200Response) SetWorkloads(v []Workload)`

SetWorkloads sets Workloads field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


