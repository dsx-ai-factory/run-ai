# Pvc

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the workspace to which the PVC bucket will be mapped. (mandatory) | [optional] 
**ExistingPvc** | Pointer to **NullableBool** | Verify existing PVC. PVC is assumed to exist when set to &#x60;true&#x60;. If set to &#x60;false&#x60;, the PVC will be created, if it does not exist. | [optional] [default to false]
**ClaimName** | Pointer to **NullableString** | Name for the PVC. Allow referencing it across workloads. If not provided, a name based on the workload name and scope will be auto-generated. | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Permit only read access to PVC. | [optional] [default to false]
**Ephemeral** | Pointer to **NullableBool** | Use &#x60;true&#x60; to set PVC to ephemeral. If set to &#x60;true&#x60;, the PVC will be deleted when the workspace is stopped. | [optional] [default to false]
**ClaimInfo** | Pointer to [**NullableClaimInfo**](ClaimInfo.md) |  | [optional] 

## Methods

### NewPvc

`func NewPvc() *Pvc`

NewPvc instantiates a new Pvc object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcWithDefaults

`func NewPvcWithDefaults() *Pvc`

NewPvcWithDefaults instantiates a new Pvc object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *Pvc) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *Pvc) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *Pvc) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *Pvc) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *Pvc) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *Pvc) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetExistingPvc

`func (o *Pvc) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *Pvc) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *Pvc) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *Pvc) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *Pvc) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *Pvc) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *Pvc) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *Pvc) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *Pvc) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *Pvc) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *Pvc) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *Pvc) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetReadOnly

`func (o *Pvc) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *Pvc) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *Pvc) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *Pvc) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *Pvc) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *Pvc) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetEphemeral

`func (o *Pvc) GetEphemeral() bool`

GetEphemeral returns the Ephemeral field if non-nil, zero value otherwise.

### GetEphemeralOk

`func (o *Pvc) GetEphemeralOk() (*bool, bool)`

GetEphemeralOk returns a tuple with the Ephemeral field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEphemeral

`func (o *Pvc) SetEphemeral(v bool)`

SetEphemeral sets Ephemeral field to given value.

### HasEphemeral

`func (o *Pvc) HasEphemeral() bool`

HasEphemeral returns a boolean if a field has been set.

### SetEphemeralNil

`func (o *Pvc) SetEphemeralNil(b bool)`

 SetEphemeralNil sets the value for Ephemeral to be an explicit nil

### UnsetEphemeral
`func (o *Pvc) UnsetEphemeral()`

UnsetEphemeral ensures that no value is present for Ephemeral, not even an explicit nil
### GetClaimInfo

`func (o *Pvc) GetClaimInfo() ClaimInfo`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *Pvc) GetClaimInfoOk() (*ClaimInfo, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *Pvc) SetClaimInfo(v ClaimInfo)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *Pvc) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *Pvc) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *Pvc) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


