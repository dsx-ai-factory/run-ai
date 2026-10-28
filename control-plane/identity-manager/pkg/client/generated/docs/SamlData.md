# SamlData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SigningCertificate** | Pointer to **string** |  | [optional] 
**SingleSignOnServiceUrl** | Pointer to **string** |  | [optional] 
**EntityId** | Pointer to **string** |  | [optional] 
**ServiceProviderMetadataUrl** | Pointer to **string** | Service Provider (SP) metadata url | [optional] 
**ValidateSignature** | **bool** | Whether the signature of SAML responses from the identity provider is validated | 

## Methods

### NewSamlData

`func NewSamlData(validateSignature bool, ) *SamlData`

NewSamlData instantiates a new SamlData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSamlDataWithDefaults

`func NewSamlDataWithDefaults() *SamlData`

NewSamlDataWithDefaults instantiates a new SamlData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSigningCertificate

`func (o *SamlData) GetSigningCertificate() string`

GetSigningCertificate returns the SigningCertificate field if non-nil, zero value otherwise.

### GetSigningCertificateOk

`func (o *SamlData) GetSigningCertificateOk() (*string, bool)`

GetSigningCertificateOk returns a tuple with the SigningCertificate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSigningCertificate

`func (o *SamlData) SetSigningCertificate(v string)`

SetSigningCertificate sets SigningCertificate field to given value.

### HasSigningCertificate

`func (o *SamlData) HasSigningCertificate() bool`

HasSigningCertificate returns a boolean if a field has been set.

### GetSingleSignOnServiceUrl

`func (o *SamlData) GetSingleSignOnServiceUrl() string`

GetSingleSignOnServiceUrl returns the SingleSignOnServiceUrl field if non-nil, zero value otherwise.

### GetSingleSignOnServiceUrlOk

`func (o *SamlData) GetSingleSignOnServiceUrlOk() (*string, bool)`

GetSingleSignOnServiceUrlOk returns a tuple with the SingleSignOnServiceUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSingleSignOnServiceUrl

`func (o *SamlData) SetSingleSignOnServiceUrl(v string)`

SetSingleSignOnServiceUrl sets SingleSignOnServiceUrl field to given value.

### HasSingleSignOnServiceUrl

`func (o *SamlData) HasSingleSignOnServiceUrl() bool`

HasSingleSignOnServiceUrl returns a boolean if a field has been set.

### GetEntityId

`func (o *SamlData) GetEntityId() string`

GetEntityId returns the EntityId field if non-nil, zero value otherwise.

### GetEntityIdOk

`func (o *SamlData) GetEntityIdOk() (*string, bool)`

GetEntityIdOk returns a tuple with the EntityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityId

`func (o *SamlData) SetEntityId(v string)`

SetEntityId sets EntityId field to given value.

### HasEntityId

`func (o *SamlData) HasEntityId() bool`

HasEntityId returns a boolean if a field has been set.

### GetServiceProviderMetadataUrl

`func (o *SamlData) GetServiceProviderMetadataUrl() string`

GetServiceProviderMetadataUrl returns the ServiceProviderMetadataUrl field if non-nil, zero value otherwise.

### GetServiceProviderMetadataUrlOk

`func (o *SamlData) GetServiceProviderMetadataUrlOk() (*string, bool)`

GetServiceProviderMetadataUrlOk returns a tuple with the ServiceProviderMetadataUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceProviderMetadataUrl

`func (o *SamlData) SetServiceProviderMetadataUrl(v string)`

SetServiceProviderMetadataUrl sets ServiceProviderMetadataUrl field to given value.

### HasServiceProviderMetadataUrl

`func (o *SamlData) HasServiceProviderMetadataUrl() bool`

HasServiceProviderMetadataUrl returns a boolean if a field has been set.

### GetValidateSignature

`func (o *SamlData) GetValidateSignature() bool`

GetValidateSignature returns the ValidateSignature field if non-nil, zero value otherwise.

### GetValidateSignatureOk

`func (o *SamlData) GetValidateSignatureOk() (*bool, bool)`

GetValidateSignatureOk returns a tuple with the ValidateSignature field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidateSignature

`func (o *SamlData) SetValidateSignature(v bool)`

SetValidateSignature sets ValidateSignature field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


