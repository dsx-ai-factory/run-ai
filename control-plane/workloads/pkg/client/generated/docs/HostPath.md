# HostPath

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the controller to which the host volume will be mapped. (mandatory) | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Force the volume to be mounted with read-only permissions. Defaults to false. | [optional] [default to true]
**MountPath** | Pointer to **NullableString** | The path that the host volume will be mounted to when in use. (mandatory) | [optional] 
**MountPropagation** | Pointer to [**NullableHostPathMountPropagation**](HostPathMountPropagation.md) |  | [optional] 

## Methods

### NewHostPath

`func NewHostPath() *HostPath`

NewHostPath instantiates a new HostPath object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathWithDefaults

`func NewHostPathWithDefaults() *HostPath`

NewHostPathWithDefaults instantiates a new HostPath object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *HostPath) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *HostPath) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *HostPath) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *HostPath) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *HostPath) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *HostPath) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *HostPath) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *HostPath) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *HostPath) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *HostPath) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *HostPath) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *HostPath) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetMountPath

`func (o *HostPath) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *HostPath) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *HostPath) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *HostPath) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *HostPath) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *HostPath) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetMountPropagation

`func (o *HostPath) GetMountPropagation() HostPathMountPropagation`

GetMountPropagation returns the MountPropagation field if non-nil, zero value otherwise.

### GetMountPropagationOk

`func (o *HostPath) GetMountPropagationOk() (*HostPathMountPropagation, bool)`

GetMountPropagationOk returns a tuple with the MountPropagation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPropagation

`func (o *HostPath) SetMountPropagation(v HostPathMountPropagation)`

SetMountPropagation sets MountPropagation field to given value.

### HasMountPropagation

`func (o *HostPath) HasMountPropagation() bool`

HasMountPropagation returns a boolean if a field has been set.

### SetMountPropagationNil

`func (o *HostPath) SetMountPropagationNil(b bool)`

 SetMountPropagationNil sets the value for MountPropagation to be an explicit nil

### UnsetMountPropagation
`func (o *HostPath) UnsetMountPropagation()`

UnsetMountPropagation ensures that no value is present for MountPropagation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


