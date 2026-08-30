# Environment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Connections** | Pointer to [**[]Connection**](Connection.md) |  | [optional] 
**Name** | **string** |  | 
**Id** | **string** |  | 
**ReplicaType** | Pointer to [**NullableReplicaType**](ReplicaType.md) |  | [optional] 

## Methods

### NewEnvironment

`func NewEnvironment(name string, id string, ) *Environment`

NewEnvironment instantiates a new Environment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentWithDefaults

`func NewEnvironmentWithDefaults() *Environment`

NewEnvironmentWithDefaults instantiates a new Environment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnections

`func (o *Environment) GetConnections() []Connection`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *Environment) GetConnectionsOk() (*[]Connection, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *Environment) SetConnections(v []Connection)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *Environment) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### GetName

`func (o *Environment) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Environment) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Environment) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *Environment) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Environment) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Environment) SetId(v string)`

SetId sets Id field to given value.


### GetReplicaType

`func (o *Environment) GetReplicaType() ReplicaType`

GetReplicaType returns the ReplicaType field if non-nil, zero value otherwise.

### GetReplicaTypeOk

`func (o *Environment) GetReplicaTypeOk() (*ReplicaType, bool)`

GetReplicaTypeOk returns a tuple with the ReplicaType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicaType

`func (o *Environment) SetReplicaType(v ReplicaType)`

SetReplicaType sets ReplicaType field to given value.

### HasReplicaType

`func (o *Environment) HasReplicaType() bool`

HasReplicaType returns a boolean if a field has been set.

### SetReplicaTypeNil

`func (o *Environment) SetReplicaTypeNil(b bool)`

 SetReplicaTypeNil sets the value for ReplicaType to be an explicit nil

### UnsetReplicaType
`func (o *Environment) UnsetReplicaType()`

UnsetReplicaType ensures that no value is present for ReplicaType, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


