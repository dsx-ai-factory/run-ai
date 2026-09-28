# HostPathInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the controller to which the host volume will be mapped. (mandatory) | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Force the volume to be mounted with read-only permissions. Defaults to false. | [optional] [default to true]
**MountPath** | Pointer to **NullableString** | The path that the host volume will be mounted to when in use. (mandatory) | [optional] 
**MountPropagation** | Pointer to [**NullableHostPathMountPropagation**](HostPathMountPropagation.md) |  | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewHostPathInstance

`func NewHostPathInstance() *HostPathInstance`

NewHostPathInstance instantiates a new HostPathInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathInstanceWithDefaults

`func NewHostPathInstanceWithDefaults() *HostPathInstance`

NewHostPathInstanceWithDefaults instantiates a new HostPathInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *HostPathInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *HostPathInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *HostPathInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *HostPathInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *HostPathInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *HostPathInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPath

`func (o *HostPathInstance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *HostPathInstance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *HostPathInstance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *HostPathInstance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *HostPathInstance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *HostPathInstance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *HostPathInstance) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *HostPathInstance) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *HostPathInstance) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *HostPathInstance) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *HostPathInstance) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *HostPathInstance) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetMountPath

`func (o *HostPathInstance) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *HostPathInstance) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *HostPathInstance) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *HostPathInstance) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *HostPathInstance) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *HostPathInstance) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetMountPropagation

`func (o *HostPathInstance) GetMountPropagation() HostPathMountPropagation`

GetMountPropagation returns the MountPropagation field if non-nil, zero value otherwise.

### GetMountPropagationOk

`func (o *HostPathInstance) GetMountPropagationOk() (*HostPathMountPropagation, bool)`

GetMountPropagationOk returns a tuple with the MountPropagation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPropagation

`func (o *HostPathInstance) SetMountPropagation(v HostPathMountPropagation)`

SetMountPropagation sets MountPropagation field to given value.

### HasMountPropagation

`func (o *HostPathInstance) HasMountPropagation() bool`

HasMountPropagation returns a boolean if a field has been set.

### SetMountPropagationNil

`func (o *HostPathInstance) SetMountPropagationNil(b bool)`

 SetMountPropagationNil sets the value for MountPropagation to be an explicit nil

### UnsetMountPropagation
`func (o *HostPathInstance) UnsetMountPropagation()`

UnsetMountPropagation ensures that no value is present for MountPropagation, not even an explicit nil
### GetExclude

`func (o *HostPathInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *HostPathInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *HostPathInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *HostPathInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *HostPathInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *HostPathInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


