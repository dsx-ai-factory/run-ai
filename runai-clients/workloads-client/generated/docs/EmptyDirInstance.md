# EmptyDirInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the EmptyDir volume will be mapped. (mandatory) | [optional] 
**Medium** | Pointer to **NullableString** | The type of storage medium for the volume. Use \&quot;Memory\&quot; for memory-backed storage, or leave empty for disk-backed storage. | [optional] 
**SizeLimit** | Pointer to **NullableString** | The total amount of local storage or memory required for the emptyDir volume. Specify using Kubernetes quantity format (e.g., 1G, 500Mi). | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewEmptyDirInstance

`func NewEmptyDirInstance() *EmptyDirInstance`

NewEmptyDirInstance instantiates a new EmptyDirInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmptyDirInstanceWithDefaults

`func NewEmptyDirInstanceWithDefaults() *EmptyDirInstance`

NewEmptyDirInstanceWithDefaults instantiates a new EmptyDirInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EmptyDirInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EmptyDirInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EmptyDirInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *EmptyDirInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *EmptyDirInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *EmptyDirInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPath

`func (o *EmptyDirInstance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *EmptyDirInstance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *EmptyDirInstance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *EmptyDirInstance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *EmptyDirInstance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *EmptyDirInstance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetMedium

`func (o *EmptyDirInstance) GetMedium() string`

GetMedium returns the Medium field if non-nil, zero value otherwise.

### GetMediumOk

`func (o *EmptyDirInstance) GetMediumOk() (*string, bool)`

GetMediumOk returns a tuple with the Medium field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMedium

`func (o *EmptyDirInstance) SetMedium(v string)`

SetMedium sets Medium field to given value.

### HasMedium

`func (o *EmptyDirInstance) HasMedium() bool`

HasMedium returns a boolean if a field has been set.

### SetMediumNil

`func (o *EmptyDirInstance) SetMediumNil(b bool)`

 SetMediumNil sets the value for Medium to be an explicit nil

### UnsetMedium
`func (o *EmptyDirInstance) UnsetMedium()`

UnsetMedium ensures that no value is present for Medium, not even an explicit nil
### GetSizeLimit

`func (o *EmptyDirInstance) GetSizeLimit() string`

GetSizeLimit returns the SizeLimit field if non-nil, zero value otherwise.

### GetSizeLimitOk

`func (o *EmptyDirInstance) GetSizeLimitOk() (*string, bool)`

GetSizeLimitOk returns a tuple with the SizeLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSizeLimit

`func (o *EmptyDirInstance) SetSizeLimit(v string)`

SetSizeLimit sets SizeLimit field to given value.

### HasSizeLimit

`func (o *EmptyDirInstance) HasSizeLimit() bool`

HasSizeLimit returns a boolean if a field has been set.

### SetSizeLimitNil

`func (o *EmptyDirInstance) SetSizeLimitNil(b bool)`

 SetSizeLimitNil sets the value for SizeLimit to be an explicit nil

### UnsetSizeLimit
`func (o *EmptyDirInstance) UnsetSizeLimit()`

UnsetSizeLimit ensures that no value is present for SizeLimit, not even an explicit nil
### GetExclude

`func (o *EmptyDirInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *EmptyDirInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *EmptyDirInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *EmptyDirInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *EmptyDirInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *EmptyDirInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


