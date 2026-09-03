# GetWorkloadPods200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Pods** | [**[]Pod**](Pod.md) |  | 

## Methods

### NewGetWorkloadPods200Response

`func NewGetWorkloadPods200Response(pods []Pod, ) *GetWorkloadPods200Response`

NewGetWorkloadPods200Response instantiates a new GetWorkloadPods200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadPods200ResponseWithDefaults

`func NewGetWorkloadPods200ResponseWithDefaults() *GetWorkloadPods200Response`

NewGetWorkloadPods200ResponseWithDefaults instantiates a new GetWorkloadPods200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetWorkloadPods200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloadPods200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloadPods200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloadPods200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetPods

`func (o *GetWorkloadPods200Response) GetPods() []Pod`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *GetWorkloadPods200Response) GetPodsOk() (*[]Pod, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *GetWorkloadPods200Response) SetPods(v []Pod)`

SetPods sets Pods field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


