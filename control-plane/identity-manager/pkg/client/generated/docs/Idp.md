# Idp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | **string** |  | 
**Type** | **string** |  | 
**RedirectUri** | **string** |  | 
**SamlData** | Pointer to [**NullableSamlData**](SamlData.md) |  | [optional] 
**OidcData** | Pointer to [**NullableOidcData**](OidcData.md) |  | [optional] 
**OcpData** | Pointer to [**NullableOcpData**](OcpData.md) |  | [optional] 
**Mappers** | Pointer to [**Mappers**](Mappers.md) |  | [optional] 

## Methods

### NewIdp

`func NewIdp(alias string, type_ string, redirectUri string, ) *Idp`

NewIdp instantiates a new Idp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIdpWithDefaults

`func NewIdpWithDefaults() *Idp`

NewIdpWithDefaults instantiates a new Idp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *Idp) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *Idp) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *Idp) SetAlias(v string)`

SetAlias sets Alias field to given value.


### GetType

`func (o *Idp) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Idp) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Idp) SetType(v string)`

SetType sets Type field to given value.


### GetRedirectUri

`func (o *Idp) GetRedirectUri() string`

GetRedirectUri returns the RedirectUri field if non-nil, zero value otherwise.

### GetRedirectUriOk

`func (o *Idp) GetRedirectUriOk() (*string, bool)`

GetRedirectUriOk returns a tuple with the RedirectUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRedirectUri

`func (o *Idp) SetRedirectUri(v string)`

SetRedirectUri sets RedirectUri field to given value.


### GetSamlData

`func (o *Idp) GetSamlData() SamlData`

GetSamlData returns the SamlData field if non-nil, zero value otherwise.

### GetSamlDataOk

`func (o *Idp) GetSamlDataOk() (*SamlData, bool)`

GetSamlDataOk returns a tuple with the SamlData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamlData

`func (o *Idp) SetSamlData(v SamlData)`

SetSamlData sets SamlData field to given value.

### HasSamlData

`func (o *Idp) HasSamlData() bool`

HasSamlData returns a boolean if a field has been set.

### SetSamlDataNil

`func (o *Idp) SetSamlDataNil(b bool)`

 SetSamlDataNil sets the value for SamlData to be an explicit nil

### UnsetSamlData
`func (o *Idp) UnsetSamlData()`

UnsetSamlData ensures that no value is present for SamlData, not even an explicit nil
### GetOidcData

`func (o *Idp) GetOidcData() OidcData`

GetOidcData returns the OidcData field if non-nil, zero value otherwise.

### GetOidcDataOk

`func (o *Idp) GetOidcDataOk() (*OidcData, bool)`

GetOidcDataOk returns a tuple with the OidcData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOidcData

`func (o *Idp) SetOidcData(v OidcData)`

SetOidcData sets OidcData field to given value.

### HasOidcData

`func (o *Idp) HasOidcData() bool`

HasOidcData returns a boolean if a field has been set.

### SetOidcDataNil

`func (o *Idp) SetOidcDataNil(b bool)`

 SetOidcDataNil sets the value for OidcData to be an explicit nil

### UnsetOidcData
`func (o *Idp) UnsetOidcData()`

UnsetOidcData ensures that no value is present for OidcData, not even an explicit nil
### GetOcpData

`func (o *Idp) GetOcpData() OcpData`

GetOcpData returns the OcpData field if non-nil, zero value otherwise.

### GetOcpDataOk

`func (o *Idp) GetOcpDataOk() (*OcpData, bool)`

GetOcpDataOk returns a tuple with the OcpData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOcpData

`func (o *Idp) SetOcpData(v OcpData)`

SetOcpData sets OcpData field to given value.

### HasOcpData

`func (o *Idp) HasOcpData() bool`

HasOcpData returns a boolean if a field has been set.

### SetOcpDataNil

`func (o *Idp) SetOcpDataNil(b bool)`

 SetOcpDataNil sets the value for OcpData to be an explicit nil

### UnsetOcpData
`func (o *Idp) UnsetOcpData()`

UnsetOcpData ensures that no value is present for OcpData, not even an explicit nil
### GetMappers

`func (o *Idp) GetMappers() Mappers`

GetMappers returns the Mappers field if non-nil, zero value otherwise.

### GetMappersOk

`func (o *Idp) GetMappersOk() (*Mappers, bool)`

GetMappersOk returns a tuple with the Mappers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappers

`func (o *Idp) SetMappers(v Mappers)`

SetMappers sets Mappers field to given value.

### HasMappers

`func (o *Idp) HasMappers() bool`

HasMappers returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


