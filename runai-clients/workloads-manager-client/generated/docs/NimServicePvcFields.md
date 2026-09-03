# NimServicePvcFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingPvc** | Pointer to **NullableBool** | Verify existing PVC. PVC is assumed to exist when set to &#x60;true&#x60;. If set to &#x60;false&#x60;, the PVC will be created, if it does not exist. | [optional] [default to false]
**ClaimName** | Pointer to **NullableString** | Name for the PVC. Allow referencing it across workloads. If not provided, a name based on the workload name and scope will be auto-generated. | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Permit only read access to PVC. | [optional] [default to false]
**ClaimInfo** | Pointer to [**NullableClaimInfo**](ClaimInfo.md) |  | [optional] 

## Methods

### NewNimServicePvcFields

`func NewNimServicePvcFields() *NimServicePvcFields`

NewNimServicePvcFields instantiates a new NimServicePvcFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServicePvcFieldsWithDefaults

`func NewNimServicePvcFieldsWithDefaults() *NimServicePvcFields`

NewNimServicePvcFieldsWithDefaults instantiates a new NimServicePvcFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingPvc

`func (o *NimServicePvcFields) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *NimServicePvcFields) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *NimServicePvcFields) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *NimServicePvcFields) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *NimServicePvcFields) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *NimServicePvcFields) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *NimServicePvcFields) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *NimServicePvcFields) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *NimServicePvcFields) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *NimServicePvcFields) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *NimServicePvcFields) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *NimServicePvcFields) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetReadOnly

`func (o *NimServicePvcFields) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *NimServicePvcFields) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *NimServicePvcFields) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *NimServicePvcFields) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *NimServicePvcFields) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *NimServicePvcFields) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetClaimInfo

`func (o *NimServicePvcFields) GetClaimInfo() ClaimInfo`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *NimServicePvcFields) GetClaimInfoOk() (*ClaimInfo, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *NimServicePvcFields) SetClaimInfo(v ClaimInfo)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *NimServicePvcFields) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *NimServicePvcFields) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *NimServicePvcFields) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


