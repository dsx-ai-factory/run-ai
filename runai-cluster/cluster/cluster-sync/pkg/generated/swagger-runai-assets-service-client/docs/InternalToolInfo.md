# InternalToolInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToolType** | [**InternalToolType**](InternalToolType.md) |  | 
**ConnectionType** | [**InternalConnectionType**](InternalConnectionType.md) |  | 
**ContainerPort** | **int32** | The port within the container that the connection exposes. | 
**NodePortInfo** | Pointer to [**NullableNodePortInfo**](NodePortInfo.md) |  | [optional] 
**ExternalUrlInfo** | Pointer to [**NullableExternalUrlInfo**](ExternalUrlInfo.md) |  | [optional] 

## Methods

### NewInternalToolInfo

`func NewInternalToolInfo(toolType InternalToolType, connectionType InternalConnectionType, containerPort int32, ) *InternalToolInfo`

NewInternalToolInfo instantiates a new InternalToolInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalToolInfoWithDefaults

`func NewInternalToolInfoWithDefaults() *InternalToolInfo`

NewInternalToolInfoWithDefaults instantiates a new InternalToolInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToolType

`func (o *InternalToolInfo) GetToolType() InternalToolType`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *InternalToolInfo) GetToolTypeOk() (*InternalToolType, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *InternalToolInfo) SetToolType(v InternalToolType)`

SetToolType sets ToolType field to given value.


### GetConnectionType

`func (o *InternalToolInfo) GetConnectionType() InternalConnectionType`

GetConnectionType returns the ConnectionType field if non-nil, zero value otherwise.

### GetConnectionTypeOk

`func (o *InternalToolInfo) GetConnectionTypeOk() (*InternalConnectionType, bool)`

GetConnectionTypeOk returns a tuple with the ConnectionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionType

`func (o *InternalToolInfo) SetConnectionType(v InternalConnectionType)`

SetConnectionType sets ConnectionType field to given value.


### GetContainerPort

`func (o *InternalToolInfo) GetContainerPort() int32`

GetContainerPort returns the ContainerPort field if non-nil, zero value otherwise.

### GetContainerPortOk

`func (o *InternalToolInfo) GetContainerPortOk() (*int32, bool)`

GetContainerPortOk returns a tuple with the ContainerPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainerPort

`func (o *InternalToolInfo) SetContainerPort(v int32)`

SetContainerPort sets ContainerPort field to given value.


### GetNodePortInfo

`func (o *InternalToolInfo) GetNodePortInfo() NodePortInfo`

GetNodePortInfo returns the NodePortInfo field if non-nil, zero value otherwise.

### GetNodePortInfoOk

`func (o *InternalToolInfo) GetNodePortInfoOk() (*NodePortInfo, bool)`

GetNodePortInfoOk returns a tuple with the NodePortInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePortInfo

`func (o *InternalToolInfo) SetNodePortInfo(v NodePortInfo)`

SetNodePortInfo sets NodePortInfo field to given value.

### HasNodePortInfo

`func (o *InternalToolInfo) HasNodePortInfo() bool`

HasNodePortInfo returns a boolean if a field has been set.

### SetNodePortInfoNil

`func (o *InternalToolInfo) SetNodePortInfoNil(b bool)`

 SetNodePortInfoNil sets the value for NodePortInfo to be an explicit nil

### UnsetNodePortInfo
`func (o *InternalToolInfo) UnsetNodePortInfo()`

UnsetNodePortInfo ensures that no value is present for NodePortInfo, not even an explicit nil
### GetExternalUrlInfo

`func (o *InternalToolInfo) GetExternalUrlInfo() ExternalUrlInfo`

GetExternalUrlInfo returns the ExternalUrlInfo field if non-nil, zero value otherwise.

### GetExternalUrlInfoOk

`func (o *InternalToolInfo) GetExternalUrlInfoOk() (*ExternalUrlInfo, bool)`

GetExternalUrlInfoOk returns a tuple with the ExternalUrlInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalUrlInfo

`func (o *InternalToolInfo) SetExternalUrlInfo(v ExternalUrlInfo)`

SetExternalUrlInfo sets ExternalUrlInfo field to given value.

### HasExternalUrlInfo

`func (o *InternalToolInfo) HasExternalUrlInfo() bool`

HasExternalUrlInfo returns a boolean if a field has been set.

### SetExternalUrlInfoNil

`func (o *InternalToolInfo) SetExternalUrlInfoNil(b bool)`

 SetExternalUrlInfoNil sets the value for ExternalUrlInfo to be an explicit nil

### UnsetExternalUrlInfo
`func (o *InternalToolInfo) UnsetExternalUrlInfo()`

UnsetExternalUrlInfo ensures that no value is present for ExternalUrlInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


