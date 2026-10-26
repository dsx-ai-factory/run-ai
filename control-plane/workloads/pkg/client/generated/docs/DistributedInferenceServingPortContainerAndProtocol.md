# DistributedInferenceServingPortContainerAndProtocol

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to **NullableInt32** | The port exposed by the container running the distributed inference service. | [optional] 
**Protocol** | Pointer to [**NullableDistributedInferenceServingPortProtocol**](DistributedInferenceServingPortProtocol.md) |  | [optional] [default to DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP]

## Methods

### NewDistributedInferenceServingPortContainerAndProtocol

`func NewDistributedInferenceServingPortContainerAndProtocol() *DistributedInferenceServingPortContainerAndProtocol`

NewDistributedInferenceServingPortContainerAndProtocol instantiates a new DistributedInferenceServingPortContainerAndProtocol object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceServingPortContainerAndProtocolWithDefaults

`func NewDistributedInferenceServingPortContainerAndProtocolWithDefaults() *DistributedInferenceServingPortContainerAndProtocol`

NewDistributedInferenceServingPortContainerAndProtocolWithDefaults instantiates a new DistributedInferenceServingPortContainerAndProtocol object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *DistributedInferenceServingPortContainerAndProtocol) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *DistributedInferenceServingPortContainerAndProtocol) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *DistributedInferenceServingPortContainerAndProtocol) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *DistributedInferenceServingPortContainerAndProtocol) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *DistributedInferenceServingPortContainerAndProtocol) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *DistributedInferenceServingPortContainerAndProtocol) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetProtocol

`func (o *DistributedInferenceServingPortContainerAndProtocol) GetProtocol() DistributedInferenceServingPortProtocol`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *DistributedInferenceServingPortContainerAndProtocol) GetProtocolOk() (*DistributedInferenceServingPortProtocol, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *DistributedInferenceServingPortContainerAndProtocol) SetProtocol(v DistributedInferenceServingPortProtocol)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *DistributedInferenceServingPortContainerAndProtocol) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *DistributedInferenceServingPortContainerAndProtocol) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *DistributedInferenceServingPortContainerAndProtocol) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


