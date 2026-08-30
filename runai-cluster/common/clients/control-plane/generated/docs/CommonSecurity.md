# CommonSecurity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Security** | Pointer to [**NullableCommonSecurityFlatFields**](CommonSecurityFlatFields.md) |  | [optional] 

## Methods

### NewCommonSecurity

`func NewCommonSecurity() *CommonSecurity`

NewCommonSecurity instantiates a new CommonSecurity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonSecurityWithDefaults

`func NewCommonSecurityWithDefaults() *CommonSecurity`

NewCommonSecurityWithDefaults instantiates a new CommonSecurity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSecurity

`func (o *CommonSecurity) GetSecurity() CommonSecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *CommonSecurity) GetSecurityOk() (*CommonSecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *CommonSecurity) SetSecurity(v CommonSecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *CommonSecurity) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *CommonSecurity) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *CommonSecurity) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


