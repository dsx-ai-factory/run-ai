# CommonSecurityOverridable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RunAsUid** | Pointer to **NullableInt64** | The user id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsUid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**RunAsGid** | Pointer to **NullableInt64** | The group id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsGid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**SupplementalGroups** | Pointer to **NullableString** | Comma separated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. Using an empty string implies reverting the supplementary groups of the image. | [optional] 

## Methods

### NewCommonSecurityOverridable

`func NewCommonSecurityOverridable() *CommonSecurityOverridable`

NewCommonSecurityOverridable instantiates a new CommonSecurityOverridable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonSecurityOverridableWithDefaults

`func NewCommonSecurityOverridableWithDefaults() *CommonSecurityOverridable`

NewCommonSecurityOverridableWithDefaults instantiates a new CommonSecurityOverridable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRunAsUid

`func (o *CommonSecurityOverridable) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *CommonSecurityOverridable) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *CommonSecurityOverridable) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *CommonSecurityOverridable) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *CommonSecurityOverridable) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *CommonSecurityOverridable) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *CommonSecurityOverridable) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *CommonSecurityOverridable) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *CommonSecurityOverridable) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *CommonSecurityOverridable) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *CommonSecurityOverridable) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *CommonSecurityOverridable) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *CommonSecurityOverridable) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *CommonSecurityOverridable) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *CommonSecurityOverridable) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *CommonSecurityOverridable) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *CommonSecurityOverridable) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *CommonSecurityOverridable) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


