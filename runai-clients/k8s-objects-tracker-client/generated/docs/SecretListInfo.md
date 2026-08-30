# SecretListInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the tenant, cluster, department or project, depending on the scope type. | 
**Name** | **string** | name of the secret. | 
**Type** | [**SecretType**](SecretType.md) |  | 
**Origin** | [**NullableOrigin**](Origin.md) |  | 
**Private** | Pointer to **NullableBool** | Whether the secret is private (user-credential). | [optional] [default to false]

## Methods

### NewSecretListInfo

`func NewSecretListInfo(clusterId string, scopeType string, scopeId string, name string, type_ SecretType, origin NullableOrigin, ) *SecretListInfo`

NewSecretListInfo instantiates a new SecretListInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretListInfoWithDefaults

`func NewSecretListInfoWithDefaults() *SecretListInfo`

NewSecretListInfoWithDefaults instantiates a new SecretListInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SecretListInfo) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SecretListInfo) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SecretListInfo) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetScopeType

`func (o *SecretListInfo) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *SecretListInfo) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *SecretListInfo) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *SecretListInfo) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *SecretListInfo) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *SecretListInfo) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *SecretListInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SecretListInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SecretListInfo) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *SecretListInfo) GetType() SecretType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SecretListInfo) GetTypeOk() (*SecretType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SecretListInfo) SetType(v SecretType)`

SetType sets Type field to given value.


### GetOrigin

`func (o *SecretListInfo) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *SecretListInfo) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *SecretListInfo) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.


### SetOriginNil

`func (o *SecretListInfo) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *SecretListInfo) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetPrivate

`func (o *SecretListInfo) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *SecretListInfo) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *SecretListInfo) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *SecretListInfo) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### SetPrivateNil

`func (o *SecretListInfo) SetPrivateNil(b bool)`

 SetPrivateNil sets the value for Private to be an explicit nil

### UnsetPrivate
`func (o *SecretListInfo) UnsetPrivate()`

UnsetPrivate ensures that no value is present for Private, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


