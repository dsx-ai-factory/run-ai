# Config

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantName** | Pointer to **string** |  | [optional] 
**AuthClientName** | Pointer to **string** |  | [optional] 
**AuthUrl** | Pointer to **string** |  | [optional] 
**AuthRealm** | Pointer to **string** |  | [optional] 

## Methods

### NewConfig

`func NewConfig() *Config`

NewConfig instantiates a new Config object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigWithDefaults

`func NewConfigWithDefaults() *Config`

NewConfigWithDefaults instantiates a new Config object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantName

`func (o *Config) GetTenantName() string`

GetTenantName returns the TenantName field if non-nil, zero value otherwise.

### GetTenantNameOk

`func (o *Config) GetTenantNameOk() (*string, bool)`

GetTenantNameOk returns a tuple with the TenantName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantName

`func (o *Config) SetTenantName(v string)`

SetTenantName sets TenantName field to given value.

### HasTenantName

`func (o *Config) HasTenantName() bool`

HasTenantName returns a boolean if a field has been set.

### GetAuthClientName

`func (o *Config) GetAuthClientName() string`

GetAuthClientName returns the AuthClientName field if non-nil, zero value otherwise.

### GetAuthClientNameOk

`func (o *Config) GetAuthClientNameOk() (*string, bool)`

GetAuthClientNameOk returns a tuple with the AuthClientName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthClientName

`func (o *Config) SetAuthClientName(v string)`

SetAuthClientName sets AuthClientName field to given value.

### HasAuthClientName

`func (o *Config) HasAuthClientName() bool`

HasAuthClientName returns a boolean if a field has been set.

### GetAuthUrl

`func (o *Config) GetAuthUrl() string`

GetAuthUrl returns the AuthUrl field if non-nil, zero value otherwise.

### GetAuthUrlOk

`func (o *Config) GetAuthUrlOk() (*string, bool)`

GetAuthUrlOk returns a tuple with the AuthUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthUrl

`func (o *Config) SetAuthUrl(v string)`

SetAuthUrl sets AuthUrl field to given value.

### HasAuthUrl

`func (o *Config) HasAuthUrl() bool`

HasAuthUrl returns a boolean if a field has been set.

### GetAuthRealm

`func (o *Config) GetAuthRealm() string`

GetAuthRealm returns the AuthRealm field if non-nil, zero value otherwise.

### GetAuthRealmOk

`func (o *Config) GetAuthRealmOk() (*string, bool)`

GetAuthRealmOk returns a tuple with the AuthRealm field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthRealm

`func (o *Config) SetAuthRealm(v string)`

SetAuthRealm sets AuthRealm field to given value.

### HasAuthRealm

`func (o *Config) HasAuthRealm() bool`

HasAuthRealm returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


