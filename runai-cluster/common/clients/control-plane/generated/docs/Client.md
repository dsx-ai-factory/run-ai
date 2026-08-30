# Client

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**ClientId** | **string** |  | 
**Secret** | **string** |  | 
**RedirectUris** | Pointer to **[]string** |  | [optional] 
**Attributes** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewClient

`func NewClient(id string, clientId string, secret string, ) *Client`

NewClient instantiates a new Client object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClientWithDefaults

`func NewClientWithDefaults() *Client`

NewClientWithDefaults instantiates a new Client object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Client) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Client) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Client) SetId(v string)`

SetId sets Id field to given value.


### GetClientId

`func (o *Client) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *Client) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *Client) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetSecret

`func (o *Client) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *Client) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *Client) SetSecret(v string)`

SetSecret sets Secret field to given value.


### GetRedirectUris

`func (o *Client) GetRedirectUris() []string`

GetRedirectUris returns the RedirectUris field if non-nil, zero value otherwise.

### GetRedirectUrisOk

`func (o *Client) GetRedirectUrisOk() (*[]string, bool)`

GetRedirectUrisOk returns a tuple with the RedirectUris field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRedirectUris

`func (o *Client) SetRedirectUris(v []string)`

SetRedirectUris sets RedirectUris field to given value.

### HasRedirectUris

`func (o *Client) HasRedirectUris() bool`

HasRedirectUris returns a boolean if a field has been set.

### GetAttributes

`func (o *Client) GetAttributes() map[string]interface{}`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *Client) GetAttributesOk() (*map[string]interface{}, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *Client) SetAttributes(v map[string]interface{})`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *Client) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


