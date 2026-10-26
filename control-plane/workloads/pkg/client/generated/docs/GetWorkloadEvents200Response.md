# GetWorkloadEvents200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Events** | [**[]Event**](Event.md) |  | 

## Methods

### NewGetWorkloadEvents200Response

`func NewGetWorkloadEvents200Response(events []Event, ) *GetWorkloadEvents200Response`

NewGetWorkloadEvents200Response instantiates a new GetWorkloadEvents200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadEvents200ResponseWithDefaults

`func NewGetWorkloadEvents200ResponseWithDefaults() *GetWorkloadEvents200Response`

NewGetWorkloadEvents200ResponseWithDefaults instantiates a new GetWorkloadEvents200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetWorkloadEvents200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloadEvents200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloadEvents200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloadEvents200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetEvents

`func (o *GetWorkloadEvents200Response) GetEvents() []Event`

GetEvents returns the Events field if non-nil, zero value otherwise.

### GetEventsOk

`func (o *GetWorkloadEvents200Response) GetEventsOk() (*[]Event, bool)`

GetEventsOk returns a tuple with the Events field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvents

`func (o *GetWorkloadEvents200Response) SetEvents(v []Event)`

SetEvents sets Events field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


