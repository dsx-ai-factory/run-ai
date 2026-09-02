# NonOverridableSpecFieldsAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Connections** | Pointer to [**[]Connection**](Connection.md) | List of connections that either expose ports from the container (each port is associated with a tool that the container runs), or URL&#39;s to be used for connecting to an external tool that is related to the action of the container (such as Weights &amp; Biases). | [optional] 
**OverrideUidGidInWorkspace** | Pointer to **bool** | Allow specifying uid/gid as part of create workspace. This is relevant only for custom uigGidSource. | [optional] [default to false]

## Methods

### NewNonOverridableSpecFieldsAllOf

`func NewNonOverridableSpecFieldsAllOf() *NonOverridableSpecFieldsAllOf`

NewNonOverridableSpecFieldsAllOf instantiates a new NonOverridableSpecFieldsAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonOverridableSpecFieldsAllOfWithDefaults

`func NewNonOverridableSpecFieldsAllOfWithDefaults() *NonOverridableSpecFieldsAllOf`

NewNonOverridableSpecFieldsAllOfWithDefaults instantiates a new NonOverridableSpecFieldsAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnections

`func (o *NonOverridableSpecFieldsAllOf) GetConnections() []Connection`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *NonOverridableSpecFieldsAllOf) GetConnectionsOk() (*[]Connection, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *NonOverridableSpecFieldsAllOf) SetConnections(v []Connection)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *NonOverridableSpecFieldsAllOf) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### GetOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFieldsAllOf) GetOverrideUidGidInWorkspace() bool`

GetOverrideUidGidInWorkspace returns the OverrideUidGidInWorkspace field if non-nil, zero value otherwise.

### GetOverrideUidGidInWorkspaceOk

`func (o *NonOverridableSpecFieldsAllOf) GetOverrideUidGidInWorkspaceOk() (*bool, bool)`

GetOverrideUidGidInWorkspaceOk returns a tuple with the OverrideUidGidInWorkspace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFieldsAllOf) SetOverrideUidGidInWorkspace(v bool)`

SetOverrideUidGidInWorkspace sets OverrideUidGidInWorkspace field to given value.

### HasOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFieldsAllOf) HasOverrideUidGidInWorkspace() bool`

HasOverrideUidGidInWorkspace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


