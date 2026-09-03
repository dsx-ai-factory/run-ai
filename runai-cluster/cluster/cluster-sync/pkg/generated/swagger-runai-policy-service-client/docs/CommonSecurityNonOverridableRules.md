# CommonSecurityNonOverridableRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ReadOnlyRootFilesystem** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewCommonSecurityNonOverridableRules

`func NewCommonSecurityNonOverridableRules() *CommonSecurityNonOverridableRules`

NewCommonSecurityNonOverridableRules instantiates a new CommonSecurityNonOverridableRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonSecurityNonOverridableRulesWithDefaults

`func NewCommonSecurityNonOverridableRulesWithDefaults() *CommonSecurityNonOverridableRules`

NewCommonSecurityNonOverridableRulesWithDefaults instantiates a new CommonSecurityNonOverridableRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUidGidSource

`func (o *CommonSecurityNonOverridableRules) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *CommonSecurityNonOverridableRules) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *CommonSecurityNonOverridableRules) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *CommonSecurityNonOverridableRules) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *CommonSecurityNonOverridableRules) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *CommonSecurityNonOverridableRules) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *CommonSecurityNonOverridableRules) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *CommonSecurityNonOverridableRules) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *CommonSecurityNonOverridableRules) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *CommonSecurityNonOverridableRules) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *CommonSecurityNonOverridableRules) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *CommonSecurityNonOverridableRules) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *CommonSecurityNonOverridableRules) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *CommonSecurityNonOverridableRules) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *CommonSecurityNonOverridableRules) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *CommonSecurityNonOverridableRules) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *CommonSecurityNonOverridableRules) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *CommonSecurityNonOverridableRules) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *CommonSecurityNonOverridableRules) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *CommonSecurityNonOverridableRules) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *CommonSecurityNonOverridableRules) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *CommonSecurityNonOverridableRules) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *CommonSecurityNonOverridableRules) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *CommonSecurityNonOverridableRules) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *CommonSecurityNonOverridableRules) GetReadOnlyRootFilesystem() BooleanRules`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *CommonSecurityNonOverridableRules) GetReadOnlyRootFilesystemOk() (*BooleanRules, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *CommonSecurityNonOverridableRules) SetReadOnlyRootFilesystem(v BooleanRules)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *CommonSecurityNonOverridableRules) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *CommonSecurityNonOverridableRules) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *CommonSecurityNonOverridableRules) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


