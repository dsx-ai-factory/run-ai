# PortRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ServiceType** | Pointer to [**NullablePortServiceTypeRules**](PortServiceTypeRules.md) |  | [optional] 
**CustomExternalPort** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**External** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ToolType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ToolName** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewPortRules

`func NewPortRules() *PortRules`

NewPortRules instantiates a new PortRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPortRulesWithDefaults

`func NewPortRulesWithDefaults() *PortRules`

NewPortRulesWithDefaults instantiates a new PortRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *PortRules) GetContainer() IntegerRules`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *PortRules) GetContainerOk() (*IntegerRules, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *PortRules) SetContainer(v IntegerRules)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *PortRules) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *PortRules) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *PortRules) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetServiceType

`func (o *PortRules) GetServiceType() PortServiceTypeRules`

GetServiceType returns the ServiceType field if non-nil, zero value otherwise.

### GetServiceTypeOk

`func (o *PortRules) GetServiceTypeOk() (*PortServiceTypeRules, bool)`

GetServiceTypeOk returns a tuple with the ServiceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceType

`func (o *PortRules) SetServiceType(v PortServiceTypeRules)`

SetServiceType sets ServiceType field to given value.

### HasServiceType

`func (o *PortRules) HasServiceType() bool`

HasServiceType returns a boolean if a field has been set.

### SetServiceTypeNil

`func (o *PortRules) SetServiceTypeNil(b bool)`

 SetServiceTypeNil sets the value for ServiceType to be an explicit nil

### UnsetServiceType
`func (o *PortRules) UnsetServiceType()`

UnsetServiceType ensures that no value is present for ServiceType, not even an explicit nil
### GetCustomExternalPort

`func (o *PortRules) GetCustomExternalPort() BooleanRules`

GetCustomExternalPort returns the CustomExternalPort field if non-nil, zero value otherwise.

### GetCustomExternalPortOk

`func (o *PortRules) GetCustomExternalPortOk() (*BooleanRules, bool)`

GetCustomExternalPortOk returns a tuple with the CustomExternalPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomExternalPort

`func (o *PortRules) SetCustomExternalPort(v BooleanRules)`

SetCustomExternalPort sets CustomExternalPort field to given value.

### HasCustomExternalPort

`func (o *PortRules) HasCustomExternalPort() bool`

HasCustomExternalPort returns a boolean if a field has been set.

### SetCustomExternalPortNil

`func (o *PortRules) SetCustomExternalPortNil(b bool)`

 SetCustomExternalPortNil sets the value for CustomExternalPort to be an explicit nil

### UnsetCustomExternalPort
`func (o *PortRules) UnsetCustomExternalPort()`

UnsetCustomExternalPort ensures that no value is present for CustomExternalPort, not even an explicit nil
### GetExternal

`func (o *PortRules) GetExternal() IntegerRules`

GetExternal returns the External field if non-nil, zero value otherwise.

### GetExternalOk

`func (o *PortRules) GetExternalOk() (*IntegerRules, bool)`

GetExternalOk returns a tuple with the External field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternal

`func (o *PortRules) SetExternal(v IntegerRules)`

SetExternal sets External field to given value.

### HasExternal

`func (o *PortRules) HasExternal() bool`

HasExternal returns a boolean if a field has been set.

### SetExternalNil

`func (o *PortRules) SetExternalNil(b bool)`

 SetExternalNil sets the value for External to be an explicit nil

### UnsetExternal
`func (o *PortRules) UnsetExternal()`

UnsetExternal ensures that no value is present for External, not even an explicit nil
### GetToolType

`func (o *PortRules) GetToolType() StringRules`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *PortRules) GetToolTypeOk() (*StringRules, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *PortRules) SetToolType(v StringRules)`

SetToolType sets ToolType field to given value.

### HasToolType

`func (o *PortRules) HasToolType() bool`

HasToolType returns a boolean if a field has been set.

### SetToolTypeNil

`func (o *PortRules) SetToolTypeNil(b bool)`

 SetToolTypeNil sets the value for ToolType to be an explicit nil

### UnsetToolType
`func (o *PortRules) UnsetToolType()`

UnsetToolType ensures that no value is present for ToolType, not even an explicit nil
### GetToolName

`func (o *PortRules) GetToolName() StringRules`

GetToolName returns the ToolName field if non-nil, zero value otherwise.

### GetToolNameOk

`func (o *PortRules) GetToolNameOk() (*StringRules, bool)`

GetToolNameOk returns a tuple with the ToolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolName

`func (o *PortRules) SetToolName(v StringRules)`

SetToolName sets ToolName field to given value.

### HasToolName

`func (o *PortRules) HasToolName() bool`

HasToolName returns a boolean if a field has been set.

### SetToolNameNil

`func (o *PortRules) SetToolNameNil(b bool)`

 SetToolNameNil sets the value for ToolName to be an explicit nil

### UnsetToolName
`func (o *PortRules) UnsetToolName()`

UnsetToolName ensures that no value is present for ToolName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


