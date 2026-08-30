# NetworkTopology

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Levels** | **[]string** | Ordered array of network topology levels | 
**Name** | **string** | Name of the network topology | 
**ClusterId** | **string** | The id of the cluster. | 
**Id** | **string** | The unique identifier for the network topology | 
**TenantId** | **int32** | The id of the tenant. | 
**CreatedAt** | **time.Time** | The timestamp for when the network topology was created | 
**CreatedBy** | Pointer to **string** | Identifier of the user who created the network topology | [optional] 
**Phase** | Pointer to **string** | Phase of the network topology in the cluster. Enum: Creating, Ready | [optional] 
**Source** | Pointer to **string** | Identifier of the the source of the topology. Useful for tracking where the topology data originated (i.e. Run:ai, Cluster). | [optional] 
**NodePools** | Pointer to **[]string** | List of node-pools associated with this network topology | [optional] [readonly] 

## Methods

### NewNetworkTopology

`func NewNetworkTopology(levels []string, name string, clusterId string, id string, tenantId int32, createdAt time.Time, ) *NetworkTopology`

NewNetworkTopology instantiates a new NetworkTopology object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologyWithDefaults

`func NewNetworkTopologyWithDefaults() *NetworkTopology`

NewNetworkTopologyWithDefaults instantiates a new NetworkTopology object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLevels

`func (o *NetworkTopology) GetLevels() []string`

GetLevels returns the Levels field if non-nil, zero value otherwise.

### GetLevelsOk

`func (o *NetworkTopology) GetLevelsOk() (*[]string, bool)`

GetLevelsOk returns a tuple with the Levels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevels

`func (o *NetworkTopology) SetLevels(v []string)`

SetLevels sets Levels field to given value.


### GetName

`func (o *NetworkTopology) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NetworkTopology) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NetworkTopology) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *NetworkTopology) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NetworkTopology) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NetworkTopology) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetId

`func (o *NetworkTopology) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NetworkTopology) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NetworkTopology) SetId(v string)`

SetId sets Id field to given value.


### GetTenantId

`func (o *NetworkTopology) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *NetworkTopology) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *NetworkTopology) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetCreatedAt

`func (o *NetworkTopology) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *NetworkTopology) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *NetworkTopology) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *NetworkTopology) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *NetworkTopology) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *NetworkTopology) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *NetworkTopology) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetPhase

`func (o *NetworkTopology) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *NetworkTopology) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *NetworkTopology) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *NetworkTopology) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetSource

`func (o *NetworkTopology) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *NetworkTopology) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *NetworkTopology) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *NetworkTopology) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetNodePools

`func (o *NetworkTopology) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *NetworkTopology) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *NetworkTopology) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *NetworkTopology) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *NetworkTopology) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *NetworkTopology) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


