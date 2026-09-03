# DatasourceListResponseAssetSpecNfs

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Path that is exported by the NFS server (mandatory). For more information, see [NFS](https://kubernetes.io/docs/concepts/storage/volumes#nfs). | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Force the NFS export to be mounted with read-only permissions. | [optional] [default to true]
**Server** | Pointer to **NullableString** | The hostname or IP address of the NFS server. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | The path that the NFS volume will be mounted to when in use. (mandatory) | [optional] 

## Methods

### NewDatasourceListResponseAssetSpecNfs

`func NewDatasourceListResponseAssetSpecNfs() *DatasourceListResponseAssetSpecNfs`

NewDatasourceListResponseAssetSpecNfs instantiates a new DatasourceListResponseAssetSpecNfs object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecNfsWithDefaults

`func NewDatasourceListResponseAssetSpecNfsWithDefaults() *DatasourceListResponseAssetSpecNfs`

NewDatasourceListResponseAssetSpecNfsWithDefaults instantiates a new DatasourceListResponseAssetSpecNfs object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *DatasourceListResponseAssetSpecNfs) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *DatasourceListResponseAssetSpecNfs) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *DatasourceListResponseAssetSpecNfs) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *DatasourceListResponseAssetSpecNfs) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *DatasourceListResponseAssetSpecNfs) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *DatasourceListResponseAssetSpecNfs) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *DatasourceListResponseAssetSpecNfs) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *DatasourceListResponseAssetSpecNfs) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *DatasourceListResponseAssetSpecNfs) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *DatasourceListResponseAssetSpecNfs) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *DatasourceListResponseAssetSpecNfs) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *DatasourceListResponseAssetSpecNfs) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetServer

`func (o *DatasourceListResponseAssetSpecNfs) GetServer() string`

GetServer returns the Server field if non-nil, zero value otherwise.

### GetServerOk

`func (o *DatasourceListResponseAssetSpecNfs) GetServerOk() (*string, bool)`

GetServerOk returns a tuple with the Server field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServer

`func (o *DatasourceListResponseAssetSpecNfs) SetServer(v string)`

SetServer sets Server field to given value.

### HasServer

`func (o *DatasourceListResponseAssetSpecNfs) HasServer() bool`

HasServer returns a boolean if a field has been set.

### SetServerNil

`func (o *DatasourceListResponseAssetSpecNfs) SetServerNil(b bool)`

 SetServerNil sets the value for Server to be an explicit nil

### UnsetServer
`func (o *DatasourceListResponseAssetSpecNfs) UnsetServer()`

UnsetServer ensures that no value is present for Server, not even an explicit nil
### GetMountPath

`func (o *DatasourceListResponseAssetSpecNfs) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *DatasourceListResponseAssetSpecNfs) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *DatasourceListResponseAssetSpecNfs) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *DatasourceListResponseAssetSpecNfs) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *DatasourceListResponseAssetSpecNfs) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *DatasourceListResponseAssetSpecNfs) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


