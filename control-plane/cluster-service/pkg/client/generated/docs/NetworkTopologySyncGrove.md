# NetworkTopologySyncGrove

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the linked Grove ClusterTopology | 
**Domains** | Pointer to **[]string** | List of domains associated with the grove topology | [optional] 
**Conditions** | Pointer to [**[]NetworkTopologySyncGroveCondition**](NetworkTopologySyncGroveCondition.md) | A list of conditions that describe the current state of the Grove ClusterTopology | [optional] 

## Methods

### NewNetworkTopologySyncGrove

`func NewNetworkTopologySyncGrove(name string, ) *NetworkTopologySyncGrove`

NewNetworkTopologySyncGrove instantiates a new NetworkTopologySyncGrove object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologySyncGroveWithDefaults

`func NewNetworkTopologySyncGroveWithDefaults() *NetworkTopologySyncGrove`

NewNetworkTopologySyncGroveWithDefaults instantiates a new NetworkTopologySyncGrove object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NetworkTopologySyncGrove) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NetworkTopologySyncGrove) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NetworkTopologySyncGrove) SetName(v string)`

SetName sets Name field to given value.


### GetDomains

`func (o *NetworkTopologySyncGrove) GetDomains() []string`

GetDomains returns the Domains field if non-nil, zero value otherwise.

### GetDomainsOk

`func (o *NetworkTopologySyncGrove) GetDomainsOk() (*[]string, bool)`

GetDomainsOk returns a tuple with the Domains field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomains

`func (o *NetworkTopologySyncGrove) SetDomains(v []string)`

SetDomains sets Domains field to given value.

### HasDomains

`func (o *NetworkTopologySyncGrove) HasDomains() bool`

HasDomains returns a boolean if a field has been set.

### GetConditions

`func (o *NetworkTopologySyncGrove) GetConditions() []NetworkTopologySyncGroveCondition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *NetworkTopologySyncGrove) GetConditionsOk() (*[]NetworkTopologySyncGroveCondition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *NetworkTopologySyncGrove) SetConditions(v []NetworkTopologySyncGroveCondition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *NetworkTopologySyncGrove) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


