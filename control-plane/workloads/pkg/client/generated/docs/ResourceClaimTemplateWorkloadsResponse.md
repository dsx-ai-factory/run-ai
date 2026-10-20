# ResourceClaimTemplateWorkloadsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** | Offset for the next page of results. Present only when there are more workloads beyond the current page (i.e., the number of returned workloads equals the limit).  | [optional] 
**Workloads** | [**[]ResourceClaimTemplateWorkload**](ResourceClaimTemplateWorkload.md) | The workloads whose pods currently reference the resource claim template. | 

## Methods

### NewResourceClaimTemplateWorkloadsResponse

`func NewResourceClaimTemplateWorkloadsResponse(workloads []ResourceClaimTemplateWorkload, ) *ResourceClaimTemplateWorkloadsResponse`

NewResourceClaimTemplateWorkloadsResponse instantiates a new ResourceClaimTemplateWorkloadsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimTemplateWorkloadsResponseWithDefaults

`func NewResourceClaimTemplateWorkloadsResponseWithDefaults() *ResourceClaimTemplateWorkloadsResponse`

NewResourceClaimTemplateWorkloadsResponseWithDefaults instantiates a new ResourceClaimTemplateWorkloadsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *ResourceClaimTemplateWorkloadsResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *ResourceClaimTemplateWorkloadsResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *ResourceClaimTemplateWorkloadsResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *ResourceClaimTemplateWorkloadsResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetWorkloads

`func (o *ResourceClaimTemplateWorkloadsResponse) GetWorkloads() []ResourceClaimTemplateWorkload`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *ResourceClaimTemplateWorkloadsResponse) GetWorkloadsOk() (*[]ResourceClaimTemplateWorkload, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *ResourceClaimTemplateWorkloadsResponse) SetWorkloads(v []ResourceClaimTemplateWorkload)`

SetWorkloads sets Workloads field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


