# ServicePort

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The Service port name. | [optional] 
**Port** | **int32** | The port exposed by the Service. | 
**TargetPort** | Pointer to **NullableInt32** | The container port that receives traffic from the Service. | [optional] 
**NodePort** | Pointer to **NullableInt32** | The externally allocated port for NodePort Services. | [optional] 
**Protocol** | Pointer to [**NullableServiceProtocol**](ServiceProtocol.md) |  | [optional] [default to SERVICEPROTOCOL_TCP]

## Methods

### NewServicePort

`func NewServicePort(port int32, ) *ServicePort`

NewServicePort instantiates a new ServicePort object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServicePortWithDefaults

`func NewServicePortWithDefaults() *ServicePort`

NewServicePortWithDefaults instantiates a new ServicePort object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ServicePort) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ServicePort) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ServicePort) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ServicePort) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *ServicePort) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *ServicePort) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPort

`func (o *ServicePort) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ServicePort) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ServicePort) SetPort(v int32)`

SetPort sets Port field to given value.


### GetTargetPort

`func (o *ServicePort) GetTargetPort() int32`

GetTargetPort returns the TargetPort field if non-nil, zero value otherwise.

### GetTargetPortOk

`func (o *ServicePort) GetTargetPortOk() (*int32, bool)`

GetTargetPortOk returns a tuple with the TargetPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetPort

`func (o *ServicePort) SetTargetPort(v int32)`

SetTargetPort sets TargetPort field to given value.

### HasTargetPort

`func (o *ServicePort) HasTargetPort() bool`

HasTargetPort returns a boolean if a field has been set.

### SetTargetPortNil

`func (o *ServicePort) SetTargetPortNil(b bool)`

 SetTargetPortNil sets the value for TargetPort to be an explicit nil

### UnsetTargetPort
`func (o *ServicePort) UnsetTargetPort()`

UnsetTargetPort ensures that no value is present for TargetPort, not even an explicit nil
### GetNodePort

`func (o *ServicePort) GetNodePort() int32`

GetNodePort returns the NodePort field if non-nil, zero value otherwise.

### GetNodePortOk

`func (o *ServicePort) GetNodePortOk() (*int32, bool)`

GetNodePortOk returns a tuple with the NodePort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePort

`func (o *ServicePort) SetNodePort(v int32)`

SetNodePort sets NodePort field to given value.

### HasNodePort

`func (o *ServicePort) HasNodePort() bool`

HasNodePort returns a boolean if a field has been set.

### SetNodePortNil

`func (o *ServicePort) SetNodePortNil(b bool)`

 SetNodePortNil sets the value for NodePort to be an explicit nil

### UnsetNodePort
`func (o *ServicePort) UnsetNodePort()`

UnsetNodePort ensures that no value is present for NodePort, not even an explicit nil
### GetProtocol

`func (o *ServicePort) GetProtocol() ServiceProtocol`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServicePort) GetProtocolOk() (*ServiceProtocol, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServicePort) SetProtocol(v ServiceProtocol)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServicePort) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServicePort) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServicePort) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


