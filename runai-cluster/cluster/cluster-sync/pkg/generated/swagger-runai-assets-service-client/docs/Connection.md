# Connection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | A unique name of this connection. This name correlates between the connection information specified at the environment asset, to the information about the connection as specified in SpecificEnv for a specific workspace. | 
**IsExternal** | Pointer to **bool** | Internal tools (isExternal&#x3D;false) are tools that run as part of the container. External tools (isExternal&#x3D;true) run outside the container, typically in the cloud. | [optional] [default to false]
**InternalToolInfo** | Pointer to [**NullableInternalToolInfo**](InternalToolInfo.md) |  | [optional] 
**ExternalToolInfo** | Pointer to [**NullableExternalToolInfo**](ExternalToolInfo.md) |  | [optional] 

## Methods

### NewConnection

`func NewConnection(name string, ) *Connection`

NewConnection instantiates a new Connection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectionWithDefaults

`func NewConnectionWithDefaults() *Connection`

NewConnectionWithDefaults instantiates a new Connection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Connection) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Connection) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Connection) SetName(v string)`

SetName sets Name field to given value.


### GetIsExternal

`func (o *Connection) GetIsExternal() bool`

GetIsExternal returns the IsExternal field if non-nil, zero value otherwise.

### GetIsExternalOk

`func (o *Connection) GetIsExternalOk() (*bool, bool)`

GetIsExternalOk returns a tuple with the IsExternal field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsExternal

`func (o *Connection) SetIsExternal(v bool)`

SetIsExternal sets IsExternal field to given value.

### HasIsExternal

`func (o *Connection) HasIsExternal() bool`

HasIsExternal returns a boolean if a field has been set.

### GetInternalToolInfo

`func (o *Connection) GetInternalToolInfo() InternalToolInfo`

GetInternalToolInfo returns the InternalToolInfo field if non-nil, zero value otherwise.

### GetInternalToolInfoOk

`func (o *Connection) GetInternalToolInfoOk() (*InternalToolInfo, bool)`

GetInternalToolInfoOk returns a tuple with the InternalToolInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInternalToolInfo

`func (o *Connection) SetInternalToolInfo(v InternalToolInfo)`

SetInternalToolInfo sets InternalToolInfo field to given value.

### HasInternalToolInfo

`func (o *Connection) HasInternalToolInfo() bool`

HasInternalToolInfo returns a boolean if a field has been set.

### SetInternalToolInfoNil

`func (o *Connection) SetInternalToolInfoNil(b bool)`

 SetInternalToolInfoNil sets the value for InternalToolInfo to be an explicit nil

### UnsetInternalToolInfo
`func (o *Connection) UnsetInternalToolInfo()`

UnsetInternalToolInfo ensures that no value is present for InternalToolInfo, not even an explicit nil
### GetExternalToolInfo

`func (o *Connection) GetExternalToolInfo() ExternalToolInfo`

GetExternalToolInfo returns the ExternalToolInfo field if non-nil, zero value otherwise.

### GetExternalToolInfoOk

`func (o *Connection) GetExternalToolInfoOk() (*ExternalToolInfo, bool)`

GetExternalToolInfoOk returns a tuple with the ExternalToolInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalToolInfo

`func (o *Connection) SetExternalToolInfo(v ExternalToolInfo)`

SetExternalToolInfo sets ExternalToolInfo field to given value.

### HasExternalToolInfo

`func (o *Connection) HasExternalToolInfo() bool`

HasExternalToolInfo returns a boolean if a field has been set.

### SetExternalToolInfoNil

`func (o *Connection) SetExternalToolInfoNil(b bool)`

 SetExternalToolInfoNil sets the value for ExternalToolInfo to be an explicit nil

### UnsetExternalToolInfo
`func (o *Connection) UnsetExternalToolInfo()`

UnsetExternalToolInfo ensures that no value is present for ExternalToolInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


