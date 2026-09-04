# NfsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ReadOnly** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Server** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**MountPath** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewNfsRules

`func NewNfsRules() *NfsRules`

NewNfsRules instantiates a new NfsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNfsRulesWithDefaults

`func NewNfsRulesWithDefaults() *NfsRules`

NewNfsRulesWithDefaults instantiates a new NfsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *NfsRules) GetPath() StringRules`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *NfsRules) GetPathOk() (*StringRules, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *NfsRules) SetPath(v StringRules)`

SetPath sets Path field to given value.

### HasPath

`func (o *NfsRules) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *NfsRules) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *NfsRules) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *NfsRules) GetReadOnly() BooleanRules`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *NfsRules) GetReadOnlyOk() (*BooleanRules, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *NfsRules) SetReadOnly(v BooleanRules)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *NfsRules) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *NfsRules) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *NfsRules) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetServer

`func (o *NfsRules) GetServer() StringRules`

GetServer returns the Server field if non-nil, zero value otherwise.

### GetServerOk

`func (o *NfsRules) GetServerOk() (*StringRules, bool)`

GetServerOk returns a tuple with the Server field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServer

`func (o *NfsRules) SetServer(v StringRules)`

SetServer sets Server field to given value.

### HasServer

`func (o *NfsRules) HasServer() bool`

HasServer returns a boolean if a field has been set.

### SetServerNil

`func (o *NfsRules) SetServerNil(b bool)`

 SetServerNil sets the value for Server to be an explicit nil

### UnsetServer
`func (o *NfsRules) UnsetServer()`

UnsetServer ensures that no value is present for Server, not even an explicit nil
### GetMountPath

`func (o *NfsRules) GetMountPath() StringRules`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *NfsRules) GetMountPathOk() (*StringRules, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *NfsRules) SetMountPath(v StringRules)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *NfsRules) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *NfsRules) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *NfsRules) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


