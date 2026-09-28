# ServingPortContainerAndProtocol

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to **NullableInt32** | The port that the container running the inference service exposes (mandatory). | [optional] 
**Protocol** | Pointer to [**NullableServingPortProtocol**](ServingPortProtocol.md) |  | [optional] 

## Methods

### NewServingPortContainerAndProtocol

`func NewServingPortContainerAndProtocol() *ServingPortContainerAndProtocol`

NewServingPortContainerAndProtocol instantiates a new ServingPortContainerAndProtocol object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortContainerAndProtocolWithDefaults

`func NewServingPortContainerAndProtocolWithDefaults() *ServingPortContainerAndProtocol`

NewServingPortContainerAndProtocolWithDefaults instantiates a new ServingPortContainerAndProtocol object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ServingPortContainerAndProtocol) GetContainer() int32`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ServingPortContainerAndProtocol) GetContainerOk() (*int32, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ServingPortContainerAndProtocol) SetContainer(v int32)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ServingPortContainerAndProtocol) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ServingPortContainerAndProtocol) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ServingPortContainerAndProtocol) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *ServingPortContainerAndProtocol) GetProtocol() ServingPortProtocol`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServingPortContainerAndProtocol) GetProtocolOk() (*ServingPortProtocol, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServingPortContainerAndProtocol) SetProtocol(v ServingPortProtocol)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServingPortContainerAndProtocol) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServingPortContainerAndProtocol) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServingPortContainerAndProtocol) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


