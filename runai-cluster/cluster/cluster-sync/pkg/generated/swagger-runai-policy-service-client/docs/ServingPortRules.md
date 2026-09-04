# ServingPortRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Protocol** | Pointer to [**NullableServingPortProtocolRules**](ServingPortProtocolRules.md) |  | [optional] 

## Methods

### NewServingPortRules

`func NewServingPortRules() *ServingPortRules`

NewServingPortRules instantiates a new ServingPortRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortRulesWithDefaults

`func NewServingPortRulesWithDefaults() *ServingPortRules`

NewServingPortRulesWithDefaults instantiates a new ServingPortRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ServingPortRules) GetContainer() IntegerRules`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ServingPortRules) GetContainerOk() (*IntegerRules, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ServingPortRules) SetContainer(v IntegerRules)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ServingPortRules) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ServingPortRules) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ServingPortRules) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *ServingPortRules) GetProtocol() ServingPortProtocolRules`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServingPortRules) GetProtocolOk() (*ServingPortProtocolRules, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServingPortRules) SetProtocol(v ServingPortProtocolRules)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServingPortRules) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServingPortRules) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServingPortRules) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


