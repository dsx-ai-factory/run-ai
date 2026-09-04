# NgcApiKeySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NgcApiKey** | Pointer to **string** | The NGC API key value. | [optional] 
**KeyId** | **int32** | id of the encryption key which has been used to encrypt an asset&#39;s data. | 

## Methods

### NewNgcApiKeySpec

`func NewNgcApiKeySpec(keyId int32, ) *NgcApiKeySpec`

NewNgcApiKeySpec instantiates a new NgcApiKeySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNgcApiKeySpecWithDefaults

`func NewNgcApiKeySpecWithDefaults() *NgcApiKeySpec`

NewNgcApiKeySpecWithDefaults instantiates a new NgcApiKeySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNgcApiKey

`func (o *NgcApiKeySpec) GetNgcApiKey() string`

GetNgcApiKey returns the NgcApiKey field if non-nil, zero value otherwise.

### GetNgcApiKeyOk

`func (o *NgcApiKeySpec) GetNgcApiKeyOk() (*string, bool)`

GetNgcApiKeyOk returns a tuple with the NgcApiKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNgcApiKey

`func (o *NgcApiKeySpec) SetNgcApiKey(v string)`

SetNgcApiKey sets NgcApiKey field to given value.

### HasNgcApiKey

`func (o *NgcApiKeySpec) HasNgcApiKey() bool`

HasNgcApiKey returns a boolean if a field has been set.

### GetKeyId

`func (o *NgcApiKeySpec) GetKeyId() int32`

GetKeyId returns the KeyId field if non-nil, zero value otherwise.

### GetKeyIdOk

`func (o *NgcApiKeySpec) GetKeyIdOk() (*int32, bool)`

GetKeyIdOk returns a tuple with the KeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyId

`func (o *NgcApiKeySpec) SetKeyId(v int32)`

SetKeyId sets KeyId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


