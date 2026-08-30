# HostPathSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | **NullableString** | Local path within the controller to which the host volume will be mapped. Path is mandatory for creating a workspace. | 
**ReadOnly** | Pointer to **NullableBool** | Whether to force the volume to be mounted with read-only permissions. Defaults to false. | [optional] [default to true]
**MountPath** | **NullableString** | The path that the host volume will be mounted to when in use. MountPath is mandatory for creating a workspace. | 

## Methods

### NewHostPathSpec

`func NewHostPathSpec(path NullableString, mountPath NullableString, ) *HostPathSpec`

NewHostPathSpec instantiates a new HostPathSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathSpecWithDefaults

`func NewHostPathSpecWithDefaults() *HostPathSpec`

NewHostPathSpecWithDefaults instantiates a new HostPathSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *HostPathSpec) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *HostPathSpec) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *HostPathSpec) SetPath(v string)`

SetPath sets Path field to given value.


### SetPathNil

`func (o *HostPathSpec) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *HostPathSpec) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *HostPathSpec) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *HostPathSpec) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *HostPathSpec) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *HostPathSpec) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *HostPathSpec) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *HostPathSpec) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetMountPath

`func (o *HostPathSpec) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *HostPathSpec) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *HostPathSpec) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.


### SetMountPathNil

`func (o *HostPathSpec) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *HostPathSpec) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


