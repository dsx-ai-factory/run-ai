# PvcFieldsUpdatable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the workload to which the PVC bucket will be mapped. (mandatory) | [optional] 
**ClaimInfo** | Pointer to [**NullableClaimInfoUpdatable**](ClaimInfoUpdatable.md) |  | [optional] 

## Methods

### NewPvcFieldsUpdatable

`func NewPvcFieldsUpdatable() *PvcFieldsUpdatable`

NewPvcFieldsUpdatable instantiates a new PvcFieldsUpdatable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcFieldsUpdatableWithDefaults

`func NewPvcFieldsUpdatableWithDefaults() *PvcFieldsUpdatable`

NewPvcFieldsUpdatableWithDefaults instantiates a new PvcFieldsUpdatable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *PvcFieldsUpdatable) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *PvcFieldsUpdatable) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *PvcFieldsUpdatable) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *PvcFieldsUpdatable) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *PvcFieldsUpdatable) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *PvcFieldsUpdatable) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetClaimInfo

`func (o *PvcFieldsUpdatable) GetClaimInfo() ClaimInfoUpdatable`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *PvcFieldsUpdatable) GetClaimInfoOk() (*ClaimInfoUpdatable, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *PvcFieldsUpdatable) SetClaimInfo(v ClaimInfoUpdatable)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *PvcFieldsUpdatable) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *PvcFieldsUpdatable) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *PvcFieldsUpdatable) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


