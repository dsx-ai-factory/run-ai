# CommonSecurityFlatFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ReadOnlyRootFilesystem** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewCommonSecurityFlatFieldsRules

`func NewCommonSecurityFlatFieldsRules() *CommonSecurityFlatFieldsRules`

NewCommonSecurityFlatFieldsRules instantiates a new CommonSecurityFlatFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonSecurityFlatFieldsRulesWithDefaults

`func NewCommonSecurityFlatFieldsRulesWithDefaults() *CommonSecurityFlatFieldsRules`

NewCommonSecurityFlatFieldsRulesWithDefaults instantiates a new CommonSecurityFlatFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUidGidSource

`func (o *CommonSecurityFlatFieldsRules) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *CommonSecurityFlatFieldsRules) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *CommonSecurityFlatFieldsRules) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *CommonSecurityFlatFieldsRules) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *CommonSecurityFlatFieldsRules) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *CommonSecurityFlatFieldsRules) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *CommonSecurityFlatFieldsRules) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *CommonSecurityFlatFieldsRules) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *CommonSecurityFlatFieldsRules) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *CommonSecurityFlatFieldsRules) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *CommonSecurityFlatFieldsRules) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *CommonSecurityFlatFieldsRules) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *CommonSecurityFlatFieldsRules) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *CommonSecurityFlatFieldsRules) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *CommonSecurityFlatFieldsRules) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *CommonSecurityFlatFieldsRules) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *CommonSecurityFlatFieldsRules) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *CommonSecurityFlatFieldsRules) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *CommonSecurityFlatFieldsRules) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *CommonSecurityFlatFieldsRules) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *CommonSecurityFlatFieldsRules) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *CommonSecurityFlatFieldsRules) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *CommonSecurityFlatFieldsRules) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *CommonSecurityFlatFieldsRules) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFieldsRules) GetReadOnlyRootFilesystem() BooleanRules`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *CommonSecurityFlatFieldsRules) GetReadOnlyRootFilesystemOk() (*BooleanRules, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFieldsRules) SetReadOnlyRootFilesystem(v BooleanRules)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFieldsRules) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *CommonSecurityFlatFieldsRules) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *CommonSecurityFlatFieldsRules) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsUid

`func (o *CommonSecurityFlatFieldsRules) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *CommonSecurityFlatFieldsRules) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *CommonSecurityFlatFieldsRules) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *CommonSecurityFlatFieldsRules) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *CommonSecurityFlatFieldsRules) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *CommonSecurityFlatFieldsRules) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *CommonSecurityFlatFieldsRules) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *CommonSecurityFlatFieldsRules) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *CommonSecurityFlatFieldsRules) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *CommonSecurityFlatFieldsRules) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *CommonSecurityFlatFieldsRules) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *CommonSecurityFlatFieldsRules) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *CommonSecurityFlatFieldsRules) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *CommonSecurityFlatFieldsRules) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *CommonSecurityFlatFieldsRules) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *CommonSecurityFlatFieldsRules) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *CommonSecurityFlatFieldsRules) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *CommonSecurityFlatFieldsRules) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


