# ServingPort

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to **NullableInt32** | The port that the container running the inference service exposes (mandatory) | [optional] 
**Protocol** | Pointer to [**NullableServingPortProtocol**](ServingPortProtocol.md) |  | [optional] 

## Methods

### NewServingPort

`func NewServingPort() *ServingPort`

NewServingPort instantiates a new ServingPort object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortWithDefaults

`func NewServingPortWithDefaults() *ServingPort`

NewServingPortWithDefaults instantiates a new ServingPort object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ServingPort) GetContainer() int32`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ServingPort) GetContainerOk() (*int32, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ServingPort) SetContainer(v int32)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ServingPort) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ServingPort) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ServingPort) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *ServingPort) GetProtocol() ServingPortProtocol`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServingPort) GetProtocolOk() (*ServingPortProtocol, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServingPort) SetProtocol(v ServingPortProtocol)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServingPort) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServingPort) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServingPort) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


