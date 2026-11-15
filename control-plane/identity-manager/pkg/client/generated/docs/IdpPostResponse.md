# IdpPostResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | **string** |  | 
**SamlData** | Pointer to [**NullableSamlData**](SamlData.md) |  | [optional] 

## Methods

### NewIdpPostResponse

`func NewIdpPostResponse(alias string, ) *IdpPostResponse`

NewIdpPostResponse instantiates a new IdpPostResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIdpPostResponseWithDefaults

`func NewIdpPostResponseWithDefaults() *IdpPostResponse`

NewIdpPostResponseWithDefaults instantiates a new IdpPostResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *IdpPostResponse) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *IdpPostResponse) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *IdpPostResponse) SetAlias(v string)`

SetAlias sets Alias field to given value.


### GetSamlData

`func (o *IdpPostResponse) GetSamlData() SamlData`

GetSamlData returns the SamlData field if non-nil, zero value otherwise.

### GetSamlDataOk

`func (o *IdpPostResponse) GetSamlDataOk() (*SamlData, bool)`

GetSamlDataOk returns a tuple with the SamlData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamlData

`func (o *IdpPostResponse) SetSamlData(v SamlData)`

SetSamlData sets SamlData field to given value.

### HasSamlData

`func (o *IdpPostResponse) HasSamlData() bool`

HasSamlData returns a boolean if a field has been set.

### SetSamlDataNil

`func (o *IdpPostResponse) SetSamlDataNil(b bool)`

 SetSamlDataNil sets the value for SamlData to be an explicit nil

### UnsetSamlData
`func (o *IdpPostResponse) UnsetSamlData()`

UnsetSamlData ensures that no value is present for SamlData, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


