# AuthConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantName** | **string** |  | 
**AuthClientID** | **string** |  | 
**AuthHost** | **string** |  | 
**IsMultiTenant** | **bool** |  | 

## Methods

### NewAuthConfiguration

`func NewAuthConfiguration(tenantName string, authClientID string, authHost string, isMultiTenant bool, ) *AuthConfiguration`

NewAuthConfiguration instantiates a new AuthConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthConfigurationWithDefaults

`func NewAuthConfigurationWithDefaults() *AuthConfiguration`

NewAuthConfigurationWithDefaults instantiates a new AuthConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantName

`func (o *AuthConfiguration) GetTenantName() string`

GetTenantName returns the TenantName field if non-nil, zero value otherwise.

### GetTenantNameOk

`func (o *AuthConfiguration) GetTenantNameOk() (*string, bool)`

GetTenantNameOk returns a tuple with the TenantName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantName

`func (o *AuthConfiguration) SetTenantName(v string)`

SetTenantName sets TenantName field to given value.


### GetAuthClientID

`func (o *AuthConfiguration) GetAuthClientID() string`

GetAuthClientID returns the AuthClientID field if non-nil, zero value otherwise.

### GetAuthClientIDOk

`func (o *AuthConfiguration) GetAuthClientIDOk() (*string, bool)`

GetAuthClientIDOk returns a tuple with the AuthClientID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthClientID

`func (o *AuthConfiguration) SetAuthClientID(v string)`

SetAuthClientID sets AuthClientID field to given value.


### GetAuthHost

`func (o *AuthConfiguration) GetAuthHost() string`

GetAuthHost returns the AuthHost field if non-nil, zero value otherwise.

### GetAuthHostOk

`func (o *AuthConfiguration) GetAuthHostOk() (*string, bool)`

GetAuthHostOk returns a tuple with the AuthHost field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthHost

`func (o *AuthConfiguration) SetAuthHost(v string)`

SetAuthHost sets AuthHost field to given value.


### GetIsMultiTenant

`func (o *AuthConfiguration) GetIsMultiTenant() bool`

GetIsMultiTenant returns the IsMultiTenant field if non-nil, zero value otherwise.

### GetIsMultiTenantOk

`func (o *AuthConfiguration) GetIsMultiTenantOk() (*bool, bool)`

GetIsMultiTenantOk returns a tuple with the IsMultiTenant field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMultiTenant

`func (o *AuthConfiguration) SetIsMultiTenant(v bool)`

SetIsMultiTenant sets IsMultiTenant field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


