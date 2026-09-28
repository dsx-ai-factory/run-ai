# ListPods200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Pods** | [**[]Pod**](Pod.md) |  | 

## Methods

### NewListPods200Response

`func NewListPods200Response(pods []Pod, ) *ListPods200Response`

NewListPods200Response instantiates a new ListPods200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListPods200ResponseWithDefaults

`func NewListPods200ResponseWithDefaults() *ListPods200Response`

NewListPods200ResponseWithDefaults instantiates a new ListPods200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *ListPods200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *ListPods200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *ListPods200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *ListPods200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetPods

`func (o *ListPods200Response) GetPods() []Pod`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *ListPods200Response) GetPodsOk() (*[]Pod, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *ListPods200Response) SetPods(v []Pod)`

SetPods sets Pods field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


