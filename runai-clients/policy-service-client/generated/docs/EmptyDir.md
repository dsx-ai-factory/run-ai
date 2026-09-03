# EmptyDir

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the workload to which the EmptyDir volume will be mapped. (mandatory) | [optional] 
**Medium** | Pointer to **NullableString** | The type of storage medium for the volume. Use \&quot;Memory\&quot; for memory-backed storage, or leave empty for disk-backed storage. | [optional] 
**SizeLimit** | Pointer to **NullableString** | The total amount of local storage or memory required for the emptyDir volume. Specify using Kubernetes quantity format (e.g., 1G, 500Mi). | [optional] 

## Methods

### NewEmptyDir

`func NewEmptyDir() *EmptyDir`

NewEmptyDir instantiates a new EmptyDir object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmptyDirWithDefaults

`func NewEmptyDirWithDefaults() *EmptyDir`

NewEmptyDirWithDefaults instantiates a new EmptyDir object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *EmptyDir) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *EmptyDir) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *EmptyDir) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *EmptyDir) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *EmptyDir) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *EmptyDir) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetMedium

`func (o *EmptyDir) GetMedium() string`

GetMedium returns the Medium field if non-nil, zero value otherwise.

### GetMediumOk

`func (o *EmptyDir) GetMediumOk() (*string, bool)`

GetMediumOk returns a tuple with the Medium field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMedium

`func (o *EmptyDir) SetMedium(v string)`

SetMedium sets Medium field to given value.

### HasMedium

`func (o *EmptyDir) HasMedium() bool`

HasMedium returns a boolean if a field has been set.

### SetMediumNil

`func (o *EmptyDir) SetMediumNil(b bool)`

 SetMediumNil sets the value for Medium to be an explicit nil

### UnsetMedium
`func (o *EmptyDir) UnsetMedium()`

UnsetMedium ensures that no value is present for Medium, not even an explicit nil
### GetSizeLimit

`func (o *EmptyDir) GetSizeLimit() string`

GetSizeLimit returns the SizeLimit field if non-nil, zero value otherwise.

### GetSizeLimitOk

`func (o *EmptyDir) GetSizeLimitOk() (*string, bool)`

GetSizeLimitOk returns a tuple with the SizeLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSizeLimit

`func (o *EmptyDir) SetSizeLimit(v string)`

SetSizeLimit sets SizeLimit field to given value.

### HasSizeLimit

`func (o *EmptyDir) HasSizeLimit() bool`

HasSizeLimit returns a boolean if a field has been set.

### SetSizeLimitNil

`func (o *EmptyDir) SetSizeLimitNil(b bool)`

 SetSizeLimitNil sets the value for SizeLimit to be an explicit nil

### UnsetSizeLimit
`func (o *EmptyDir) UnsetSizeLimit()`

UnsetSizeLimit ensures that no value is present for SizeLimit, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


