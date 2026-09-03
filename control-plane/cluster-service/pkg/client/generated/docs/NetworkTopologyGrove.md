# NetworkTopologyGrove

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the linked Grove ClusterTopology | 
**Domains** | Pointer to **[]string** | Topology levels defined in the linked Grove ClusterTopology | [optional] 
**Conditions** | Pointer to [**[]NetworkTopologyGroveCondition**](NetworkTopologyGroveCondition.md) | A list of conditions that describe the current state of the Grove ClusterTopology | [optional] 

## Methods

### NewNetworkTopologyGrove

`func NewNetworkTopologyGrove(name string, ) *NetworkTopologyGrove`

NewNetworkTopologyGrove instantiates a new NetworkTopologyGrove object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologyGroveWithDefaults

`func NewNetworkTopologyGroveWithDefaults() *NetworkTopologyGrove`

NewNetworkTopologyGroveWithDefaults instantiates a new NetworkTopologyGrove object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NetworkTopologyGrove) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NetworkTopologyGrove) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NetworkTopologyGrove) SetName(v string)`

SetName sets Name field to given value.


### GetDomains

`func (o *NetworkTopologyGrove) GetDomains() []string`

GetDomains returns the Domains field if non-nil, zero value otherwise.

### GetDomainsOk

`func (o *NetworkTopologyGrove) GetDomainsOk() (*[]string, bool)`

GetDomainsOk returns a tuple with the Domains field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomains

`func (o *NetworkTopologyGrove) SetDomains(v []string)`

SetDomains sets Domains field to given value.

### HasDomains

`func (o *NetworkTopologyGrove) HasDomains() bool`

HasDomains returns a boolean if a field has been set.

### GetConditions

`func (o *NetworkTopologyGrove) GetConditions() []NetworkTopologyGroveCondition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *NetworkTopologyGrove) GetConditionsOk() (*[]NetworkTopologyGroveCondition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *NetworkTopologyGrove) SetConditions(v []NetworkTopologyGroveCondition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *NetworkTopologyGrove) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


