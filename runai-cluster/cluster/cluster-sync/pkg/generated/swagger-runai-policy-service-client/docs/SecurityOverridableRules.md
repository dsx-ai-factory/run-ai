# SecurityOverridableRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewSecurityOverridableRules

`func NewSecurityOverridableRules() *SecurityOverridableRules`

NewSecurityOverridableRules instantiates a new SecurityOverridableRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecurityOverridableRulesWithDefaults

`func NewSecurityOverridableRulesWithDefaults() *SecurityOverridableRules`

NewSecurityOverridableRulesWithDefaults instantiates a new SecurityOverridableRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRunAsUid

`func (o *SecurityOverridableRules) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SecurityOverridableRules) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SecurityOverridableRules) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SecurityOverridableRules) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SecurityOverridableRules) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SecurityOverridableRules) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *SecurityOverridableRules) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SecurityOverridableRules) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SecurityOverridableRules) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SecurityOverridableRules) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SecurityOverridableRules) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SecurityOverridableRules) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *SecurityOverridableRules) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SecurityOverridableRules) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SecurityOverridableRules) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SecurityOverridableRules) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SecurityOverridableRules) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SecurityOverridableRules) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


