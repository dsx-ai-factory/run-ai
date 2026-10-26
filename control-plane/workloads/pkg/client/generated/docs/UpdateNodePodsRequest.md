# UpdateNodePodsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NetworkTopologyLabelValues** | Pointer to **[]string** | The topology levels of node (node address) according to nodepool assigned topology | [optional] 

## Methods

### NewUpdateNodePodsRequest

`func NewUpdateNodePodsRequest() *UpdateNodePodsRequest`

NewUpdateNodePodsRequest instantiates a new UpdateNodePodsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateNodePodsRequestWithDefaults

`func NewUpdateNodePodsRequestWithDefaults() *UpdateNodePodsRequest`

NewUpdateNodePodsRequestWithDefaults instantiates a new UpdateNodePodsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNetworkTopologyLabelValues

`func (o *UpdateNodePodsRequest) GetNetworkTopologyLabelValues() []string`

GetNetworkTopologyLabelValues returns the NetworkTopologyLabelValues field if non-nil, zero value otherwise.

### GetNetworkTopologyLabelValuesOk

`func (o *UpdateNodePodsRequest) GetNetworkTopologyLabelValuesOk() (*[]string, bool)`

GetNetworkTopologyLabelValuesOk returns a tuple with the NetworkTopologyLabelValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyLabelValues

`func (o *UpdateNodePodsRequest) SetNetworkTopologyLabelValues(v []string)`

SetNetworkTopologyLabelValues sets NetworkTopologyLabelValues field to given value.

### HasNetworkTopologyLabelValues

`func (o *UpdateNodePodsRequest) HasNetworkTopologyLabelValues() bool`

HasNetworkTopologyLabelValues returns a boolean if a field has been set.

### SetNetworkTopologyLabelValuesNil

`func (o *UpdateNodePodsRequest) SetNetworkTopologyLabelValuesNil(b bool)`

 SetNetworkTopologyLabelValuesNil sets the value for NetworkTopologyLabelValues to be an explicit nil

### UnsetNetworkTopologyLabelValues
`func (o *UpdateNodePodsRequest) UnsetNetworkTopologyLabelValues()`

UnsetNetworkTopologyLabelValues ensures that no value is present for NetworkTopologyLabelValues, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


