# RealmCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RealmName** | **string** |  | 
**TenantId** | **string** |  | 
**ExternalRedirectUri** | Pointer to **[]string** |  | [optional] 

## Methods

### NewRealmCreationRequest

`func NewRealmCreationRequest(realmName string, tenantId string, ) *RealmCreationRequest`

NewRealmCreationRequest instantiates a new RealmCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRealmCreationRequestWithDefaults

`func NewRealmCreationRequestWithDefaults() *RealmCreationRequest`

NewRealmCreationRequestWithDefaults instantiates a new RealmCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRealmName

`func (o *RealmCreationRequest) GetRealmName() string`

GetRealmName returns the RealmName field if non-nil, zero value otherwise.

### GetRealmNameOk

`func (o *RealmCreationRequest) GetRealmNameOk() (*string, bool)`

GetRealmNameOk returns a tuple with the RealmName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRealmName

`func (o *RealmCreationRequest) SetRealmName(v string)`

SetRealmName sets RealmName field to given value.


### GetTenantId

`func (o *RealmCreationRequest) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *RealmCreationRequest) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *RealmCreationRequest) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.


### GetExternalRedirectUri

`func (o *RealmCreationRequest) GetExternalRedirectUri() []string`

GetExternalRedirectUri returns the ExternalRedirectUri field if non-nil, zero value otherwise.

### GetExternalRedirectUriOk

`func (o *RealmCreationRequest) GetExternalRedirectUriOk() (*[]string, bool)`

GetExternalRedirectUriOk returns a tuple with the ExternalRedirectUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalRedirectUri

`func (o *RealmCreationRequest) SetExternalRedirectUri(v []string)`

SetExternalRedirectUri sets ExternalRedirectUri field to given value.

### HasExternalRedirectUri

`func (o *RealmCreationRequest) HasExternalRedirectUri() bool`

HasExternalRedirectUri returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


