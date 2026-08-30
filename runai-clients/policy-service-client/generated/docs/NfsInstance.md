# NfsInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Path** | Pointer to **NullableString** | Path that is exported by the NFS server (mandatory). For more information, see [NFS](https://kubernetes.io/docs/concepts/storage/volumes#nfs). | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Force the NFS export to be mounted with read-only permissions. | [optional] [default to true]
**Server** | Pointer to **NullableString** | The hostname or IP address of the NFS server. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | The path that the NFS volume will be mounted to when in use. (mandatory) | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewNfsInstance

`func NewNfsInstance() *NfsInstance`

NewNfsInstance instantiates a new NfsInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNfsInstanceWithDefaults

`func NewNfsInstanceWithDefaults() *NfsInstance`

NewNfsInstanceWithDefaults instantiates a new NfsInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NfsInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NfsInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NfsInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NfsInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *NfsInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *NfsInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPath

`func (o *NfsInstance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *NfsInstance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *NfsInstance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *NfsInstance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *NfsInstance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *NfsInstance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *NfsInstance) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *NfsInstance) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *NfsInstance) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *NfsInstance) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *NfsInstance) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *NfsInstance) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetServer

`func (o *NfsInstance) GetServer() string`

GetServer returns the Server field if non-nil, zero value otherwise.

### GetServerOk

`func (o *NfsInstance) GetServerOk() (*string, bool)`

GetServerOk returns a tuple with the Server field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServer

`func (o *NfsInstance) SetServer(v string)`

SetServer sets Server field to given value.

### HasServer

`func (o *NfsInstance) HasServer() bool`

HasServer returns a boolean if a field has been set.

### SetServerNil

`func (o *NfsInstance) SetServerNil(b bool)`

 SetServerNil sets the value for Server to be an explicit nil

### UnsetServer
`func (o *NfsInstance) UnsetServer()`

UnsetServer ensures that no value is present for Server, not even an explicit nil
### GetMountPath

`func (o *NfsInstance) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *NfsInstance) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *NfsInstance) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *NfsInstance) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *NfsInstance) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *NfsInstance) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetExclude

`func (o *NfsInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *NfsInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *NfsInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *NfsInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *NfsInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *NfsInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


