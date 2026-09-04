# IdpCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Type** | [**IdpCreationRequestTypeEnum**](IdpCreationRequestTypeEnum.md) |  | 
**SamlData** | Pointer to [**NullableSamlCreationData**](SamlCreationData.md) |  | [optional] 
**OidcData** | Pointer to [**NullableOidcCreationData**](OidcCreationData.md) |  | [optional] 
**OcpData** | Pointer to [**NullableOcpCreationData**](OcpCreationData.md) |  | [optional] 
**Mappers** | Pointer to [**Mappers**](Mappers.md) |  | [optional] 

## Methods

### NewIdpCreationRequest

`func NewIdpCreationRequest(type_ IdpCreationRequestTypeEnum, ) *IdpCreationRequest`

NewIdpCreationRequest instantiates a new IdpCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIdpCreationRequestWithDefaults

`func NewIdpCreationRequestWithDefaults() *IdpCreationRequest`

NewIdpCreationRequestWithDefaults instantiates a new IdpCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *IdpCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *IdpCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *IdpCreationRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *IdpCreationRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetType

`func (o *IdpCreationRequest) GetType() IdpCreationRequestTypeEnum`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *IdpCreationRequest) GetTypeOk() (*IdpCreationRequestTypeEnum, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *IdpCreationRequest) SetType(v IdpCreationRequestTypeEnum)`

SetType sets Type field to given value.


### GetSamlData

`func (o *IdpCreationRequest) GetSamlData() SamlCreationData`

GetSamlData returns the SamlData field if non-nil, zero value otherwise.

### GetSamlDataOk

`func (o *IdpCreationRequest) GetSamlDataOk() (*SamlCreationData, bool)`

GetSamlDataOk returns a tuple with the SamlData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamlData

`func (o *IdpCreationRequest) SetSamlData(v SamlCreationData)`

SetSamlData sets SamlData field to given value.

### HasSamlData

`func (o *IdpCreationRequest) HasSamlData() bool`

HasSamlData returns a boolean if a field has been set.

### SetSamlDataNil

`func (o *IdpCreationRequest) SetSamlDataNil(b bool)`

 SetSamlDataNil sets the value for SamlData to be an explicit nil

### UnsetSamlData
`func (o *IdpCreationRequest) UnsetSamlData()`

UnsetSamlData ensures that no value is present for SamlData, not even an explicit nil
### GetOidcData

`func (o *IdpCreationRequest) GetOidcData() OidcCreationData`

GetOidcData returns the OidcData field if non-nil, zero value otherwise.

### GetOidcDataOk

`func (o *IdpCreationRequest) GetOidcDataOk() (*OidcCreationData, bool)`

GetOidcDataOk returns a tuple with the OidcData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOidcData

`func (o *IdpCreationRequest) SetOidcData(v OidcCreationData)`

SetOidcData sets OidcData field to given value.

### HasOidcData

`func (o *IdpCreationRequest) HasOidcData() bool`

HasOidcData returns a boolean if a field has been set.

### SetOidcDataNil

`func (o *IdpCreationRequest) SetOidcDataNil(b bool)`

 SetOidcDataNil sets the value for OidcData to be an explicit nil

### UnsetOidcData
`func (o *IdpCreationRequest) UnsetOidcData()`

UnsetOidcData ensures that no value is present for OidcData, not even an explicit nil
### GetOcpData

`func (o *IdpCreationRequest) GetOcpData() OcpCreationData`

GetOcpData returns the OcpData field if non-nil, zero value otherwise.

### GetOcpDataOk

`func (o *IdpCreationRequest) GetOcpDataOk() (*OcpCreationData, bool)`

GetOcpDataOk returns a tuple with the OcpData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOcpData

`func (o *IdpCreationRequest) SetOcpData(v OcpCreationData)`

SetOcpData sets OcpData field to given value.

### HasOcpData

`func (o *IdpCreationRequest) HasOcpData() bool`

HasOcpData returns a boolean if a field has been set.

### SetOcpDataNil

`func (o *IdpCreationRequest) SetOcpDataNil(b bool)`

 SetOcpDataNil sets the value for OcpData to be an explicit nil

### UnsetOcpData
`func (o *IdpCreationRequest) UnsetOcpData()`

UnsetOcpData ensures that no value is present for OcpData, not even an explicit nil
### GetMappers

`func (o *IdpCreationRequest) GetMappers() Mappers`

GetMappers returns the Mappers field if non-nil, zero value otherwise.

### GetMappersOk

`func (o *IdpCreationRequest) GetMappersOk() (*Mappers, bool)`

GetMappersOk returns a tuple with the Mappers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappers

`func (o *IdpCreationRequest) SetMappers(v Mappers)`

SetMappers sets Mappers field to given value.

### HasMappers

`func (o *IdpCreationRequest) HasMappers() bool`

HasMappers returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


