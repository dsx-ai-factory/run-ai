# TenantLogo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | **int64** |  | 
**Logo** | **string** | Data URL of the logo (e.g. \&quot;data:image/png;base64,...\&quot;) | 

## Methods

### NewTenantLogo

`func NewTenantLogo(tenantId int64, logo string, ) *TenantLogo`

NewTenantLogo instantiates a new TenantLogo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantLogoWithDefaults

`func NewTenantLogoWithDefaults() *TenantLogo`

NewTenantLogoWithDefaults instantiates a new TenantLogo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *TenantLogo) GetTenantId() int64`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *TenantLogo) GetTenantIdOk() (*int64, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *TenantLogo) SetTenantId(v int64)`

SetTenantId sets TenantId field to given value.


### GetLogo

`func (o *TenantLogo) GetLogo() string`

GetLogo returns the Logo field if non-nil, zero value otherwise.

### GetLogoOk

`func (o *TenantLogo) GetLogoOk() (*string, bool)`

GetLogoOk returns a tuple with the Logo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogo

`func (o *TenantLogo) SetLogo(v string)`

SetLogo sets Logo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


