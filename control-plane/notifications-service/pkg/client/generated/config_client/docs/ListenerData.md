# ListenerData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ListenerId** | **string** |  | 
**ListenerType** | **string** |  | 
**ListenerConfig** | **map[string]interface{}** |  | 

## Methods

### NewListenerData

`func NewListenerData(listenerId string, listenerType string, listenerConfig map[string]interface{}, ) *ListenerData`

NewListenerData instantiates a new ListenerData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListenerDataWithDefaults

`func NewListenerDataWithDefaults() *ListenerData`

NewListenerDataWithDefaults instantiates a new ListenerData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetListenerId

`func (o *ListenerData) GetListenerId() string`

GetListenerId returns the ListenerId field if non-nil, zero value otherwise.

### GetListenerIdOk

`func (o *ListenerData) GetListenerIdOk() (*string, bool)`

GetListenerIdOk returns a tuple with the ListenerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListenerId

`func (o *ListenerData) SetListenerId(v string)`

SetListenerId sets ListenerId field to given value.


### GetListenerType

`func (o *ListenerData) GetListenerType() string`

GetListenerType returns the ListenerType field if non-nil, zero value otherwise.

### GetListenerTypeOk

`func (o *ListenerData) GetListenerTypeOk() (*string, bool)`

GetListenerTypeOk returns a tuple with the ListenerType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListenerType

`func (o *ListenerData) SetListenerType(v string)`

SetListenerType sets ListenerType field to given value.


### GetListenerConfig

`func (o *ListenerData) GetListenerConfig() map[string]interface{}`

GetListenerConfig returns the ListenerConfig field if non-nil, zero value otherwise.

### GetListenerConfigOk

`func (o *ListenerData) GetListenerConfigOk() (*map[string]interface{}, bool)`

GetListenerConfigOk returns a tuple with the ListenerConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListenerConfig

`func (o *ListenerData) SetListenerConfig(v map[string]interface{})`

SetListenerConfig sets ListenerConfig field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


