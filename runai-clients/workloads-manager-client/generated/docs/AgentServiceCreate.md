# AgentServiceCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | **int32** | The port that the service exposes. | 
**Selector** | Pointer to **map[string]string** | Label selector to identify the pods that this service routes traffic to. | [optional] 

## Methods

### NewAgentServiceCreate

`func NewAgentServiceCreate(port int32, ) *AgentServiceCreate`

NewAgentServiceCreate instantiates a new AgentServiceCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAgentServiceCreateWithDefaults

`func NewAgentServiceCreateWithDefaults() *AgentServiceCreate`

NewAgentServiceCreateWithDefaults instantiates a new AgentServiceCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *AgentServiceCreate) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *AgentServiceCreate) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *AgentServiceCreate) SetPort(v int32)`

SetPort sets Port field to given value.


### GetSelector

`func (o *AgentServiceCreate) GetSelector() map[string]string`

GetSelector returns the Selector field if non-nil, zero value otherwise.

### GetSelectorOk

`func (o *AgentServiceCreate) GetSelectorOk() (*map[string]string, bool)`

GetSelectorOk returns a tuple with the Selector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelector

`func (o *AgentServiceCreate) SetSelector(v map[string]string)`

SetSelector sets Selector field to given value.

### HasSelector

`func (o *AgentServiceCreate) HasSelector() bool`

HasSelector returns a boolean if a field has been set.

### SetSelectorNil

`func (o *AgentServiceCreate) SetSelectorNil(b bool)`

 SetSelectorNil sets the value for Selector to be an explicit nil

### UnsetSelector
`func (o *AgentServiceCreate) UnsetSelector()`

UnsetSelector ensures that no value is present for Selector, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


