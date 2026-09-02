# DatasourceListResponseAssetSpecHostPath

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the controller to which the host volume will be mapped. (mandatory) | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Force the volume to be mounted with read-only permissions. Defaults to false. | [optional] [default to true]
**MountPath** | Pointer to **NullableString** | The path that the host volume will be mounted to when in use. (mandatory) | [optional] 
**MountPropagation** | Pointer to [**NullableHostPathMountPropagation**](HostPathMountPropagation.md) |  | [optional] 

## Methods

### NewDatasourceListResponseAssetSpecHostPath

`func NewDatasourceListResponseAssetSpecHostPath() *DatasourceListResponseAssetSpecHostPath`

NewDatasourceListResponseAssetSpecHostPath instantiates a new DatasourceListResponseAssetSpecHostPath object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecHostPathWithDefaults

`func NewDatasourceListResponseAssetSpecHostPathWithDefaults() *DatasourceListResponseAssetSpecHostPath`

NewDatasourceListResponseAssetSpecHostPathWithDefaults instantiates a new DatasourceListResponseAssetSpecHostPath object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *DatasourceListResponseAssetSpecHostPath) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *DatasourceListResponseAssetSpecHostPath) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *DatasourceListResponseAssetSpecHostPath) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *DatasourceListResponseAssetSpecHostPath) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *DatasourceListResponseAssetSpecHostPath) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *DatasourceListResponseAssetSpecHostPath) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *DatasourceListResponseAssetSpecHostPath) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *DatasourceListResponseAssetSpecHostPath) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *DatasourceListResponseAssetSpecHostPath) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *DatasourceListResponseAssetSpecHostPath) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *DatasourceListResponseAssetSpecHostPath) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *DatasourceListResponseAssetSpecHostPath) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetMountPath

`func (o *DatasourceListResponseAssetSpecHostPath) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *DatasourceListResponseAssetSpecHostPath) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *DatasourceListResponseAssetSpecHostPath) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *DatasourceListResponseAssetSpecHostPath) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *DatasourceListResponseAssetSpecHostPath) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *DatasourceListResponseAssetSpecHostPath) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetMountPropagation

`func (o *DatasourceListResponseAssetSpecHostPath) GetMountPropagation() HostPathMountPropagation`

GetMountPropagation returns the MountPropagation field if non-nil, zero value otherwise.

### GetMountPropagationOk

`func (o *DatasourceListResponseAssetSpecHostPath) GetMountPropagationOk() (*HostPathMountPropagation, bool)`

GetMountPropagationOk returns a tuple with the MountPropagation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPropagation

`func (o *DatasourceListResponseAssetSpecHostPath) SetMountPropagation(v HostPathMountPropagation)`

SetMountPropagation sets MountPropagation field to given value.

### HasMountPropagation

`func (o *DatasourceListResponseAssetSpecHostPath) HasMountPropagation() bool`

HasMountPropagation returns a boolean if a field has been set.

### SetMountPropagationNil

`func (o *DatasourceListResponseAssetSpecHostPath) SetMountPropagationNil(b bool)`

 SetMountPropagationNil sets the value for MountPropagation to be an explicit nil

### UnsetMountPropagation
`func (o *DatasourceListResponseAssetSpecHostPath) UnsetMountPropagation()`

UnsetMountPropagation ensures that no value is present for MountPropagation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


