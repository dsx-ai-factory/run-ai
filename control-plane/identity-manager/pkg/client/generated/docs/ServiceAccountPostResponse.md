# ServiceAccountPostResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the service account. | 
**Name** | **string** | The name of the service account. | 
**ClientSecret** | **string** | Secret credential paired with the client ID. Used to obtain authentication tokens. Displayed only once upon creation. | 
**ClientId** | Pointer to **string** | The client identifier associated with the service account. Used when requesting authentication tokens. | [optional] 

## Methods

### NewServiceAccountPostResponse

`func NewServiceAccountPostResponse(id string, name string, clientSecret string, ) *ServiceAccountPostResponse`

NewServiceAccountPostResponse instantiates a new ServiceAccountPostResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServiceAccountPostResponseWithDefaults

`func NewServiceAccountPostResponseWithDefaults() *ServiceAccountPostResponse`

NewServiceAccountPostResponseWithDefaults instantiates a new ServiceAccountPostResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ServiceAccountPostResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ServiceAccountPostResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ServiceAccountPostResponse) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ServiceAccountPostResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ServiceAccountPostResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ServiceAccountPostResponse) SetName(v string)`

SetName sets Name field to given value.


### GetClientSecret

`func (o *ServiceAccountPostResponse) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *ServiceAccountPostResponse) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *ServiceAccountPostResponse) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.


### GetClientId

`func (o *ServiceAccountPostResponse) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *ServiceAccountPostResponse) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *ServiceAccountPostResponse) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *ServiceAccountPostResponse) HasClientId() bool`

HasClientId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


