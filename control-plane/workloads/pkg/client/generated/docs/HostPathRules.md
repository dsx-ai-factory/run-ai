# HostPathRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ReadOnly** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**MountPath** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**MountPropagation** | Pointer to [**NullableHostPathMountPropagationRules**](HostPathMountPropagationRules.md) |  | [optional] 

## Methods

### NewHostPathRules

`func NewHostPathRules() *HostPathRules`

NewHostPathRules instantiates a new HostPathRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathRulesWithDefaults

`func NewHostPathRulesWithDefaults() *HostPathRules`

NewHostPathRulesWithDefaults instantiates a new HostPathRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *HostPathRules) GetPath() StringRules`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *HostPathRules) GetPathOk() (*StringRules, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *HostPathRules) SetPath(v StringRules)`

SetPath sets Path field to given value.

### HasPath

`func (o *HostPathRules) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *HostPathRules) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *HostPathRules) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *HostPathRules) GetReadOnly() BooleanRules`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *HostPathRules) GetReadOnlyOk() (*BooleanRules, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *HostPathRules) SetReadOnly(v BooleanRules)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *HostPathRules) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *HostPathRules) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *HostPathRules) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetMountPath

`func (o *HostPathRules) GetMountPath() StringRules`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *HostPathRules) GetMountPathOk() (*StringRules, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *HostPathRules) SetMountPath(v StringRules)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *HostPathRules) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *HostPathRules) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *HostPathRules) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetMountPropagation

`func (o *HostPathRules) GetMountPropagation() HostPathMountPropagationRules`

GetMountPropagation returns the MountPropagation field if non-nil, zero value otherwise.

### GetMountPropagationOk

`func (o *HostPathRules) GetMountPropagationOk() (*HostPathMountPropagationRules, bool)`

GetMountPropagationOk returns a tuple with the MountPropagation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPropagation

`func (o *HostPathRules) SetMountPropagation(v HostPathMountPropagationRules)`

SetMountPropagation sets MountPropagation field to given value.

### HasMountPropagation

`func (o *HostPathRules) HasMountPropagation() bool`

HasMountPropagation returns a boolean if a field has been set.

### SetMountPropagationNil

`func (o *HostPathRules) SetMountPropagationNil(b bool)`

 SetMountPropagationNil sets the value for MountPropagation to be an explicit nil

### UnsetMountPropagation
`func (o *HostPathRules) UnsetMountPropagation()`

UnsetMountPropagation ensures that no value is present for MountPropagation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


