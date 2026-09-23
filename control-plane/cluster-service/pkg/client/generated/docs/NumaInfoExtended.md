# NumaInfoExtended

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TopologyManagerPolicy** | Pointer to [**NullableTopologyManagerPolicy**](TopologyManagerPolicy.md) |  | [optional] 
**Nodes** | [**[]NumaNodeDetails**](NumaNodeDetails.md) | The list of NUMA nodes on the physical machine and their resource availability. | 

## Methods

### NewNumaInfoExtended

`func NewNumaInfoExtended(nodes []NumaNodeDetails, ) *NumaInfoExtended`

NewNumaInfoExtended instantiates a new NumaInfoExtended object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNumaInfoExtendedWithDefaults

`func NewNumaInfoExtendedWithDefaults() *NumaInfoExtended`

NewNumaInfoExtendedWithDefaults instantiates a new NumaInfoExtended object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTopologyManagerPolicy

`func (o *NumaInfoExtended) GetTopologyManagerPolicy() TopologyManagerPolicy`

GetTopologyManagerPolicy returns the TopologyManagerPolicy field if non-nil, zero value otherwise.

### GetTopologyManagerPolicyOk

`func (o *NumaInfoExtended) GetTopologyManagerPolicyOk() (*TopologyManagerPolicy, bool)`

GetTopologyManagerPolicyOk returns a tuple with the TopologyManagerPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopologyManagerPolicy

`func (o *NumaInfoExtended) SetTopologyManagerPolicy(v TopologyManagerPolicy)`

SetTopologyManagerPolicy sets TopologyManagerPolicy field to given value.

### HasTopologyManagerPolicy

`func (o *NumaInfoExtended) HasTopologyManagerPolicy() bool`

HasTopologyManagerPolicy returns a boolean if a field has been set.

### SetTopologyManagerPolicyNil

`func (o *NumaInfoExtended) SetTopologyManagerPolicyNil(b bool)`

 SetTopologyManagerPolicyNil sets the value for TopologyManagerPolicy to be an explicit nil

### UnsetTopologyManagerPolicy
`func (o *NumaInfoExtended) UnsetTopologyManagerPolicy()`

UnsetTopologyManagerPolicy ensures that no value is present for TopologyManagerPolicy, not even an explicit nil
### GetNodes

`func (o *NumaInfoExtended) GetNodes() []NumaNodeDetails`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NumaInfoExtended) GetNodesOk() (*[]NumaNodeDetails, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NumaInfoExtended) SetNodes(v []NumaNodeDetails)`

SetNodes sets Nodes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


