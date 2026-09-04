# AccessKeyPostResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the access key. | 
**ClientId** | **string** | The client identifier associated with the access key. Used when requesting authentication tokens. | 
**Name** | **string** | The name of the access key. | 
**ClientSecret** | **string** | Secret credential paired with the client ID. Used to obtain authentication tokens. Displayed only once upon creation. | 
**CreatedAt** | **NullableTime** | The timestamp for when the access key was created. | 

## Methods

### NewAccessKeyPostResponse

`func NewAccessKeyPostResponse(id string, clientId string, name string, clientSecret string, createdAt NullableTime, ) *AccessKeyPostResponse`

NewAccessKeyPostResponse instantiates a new AccessKeyPostResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyPostResponseWithDefaults

`func NewAccessKeyPostResponseWithDefaults() *AccessKeyPostResponse`

NewAccessKeyPostResponseWithDefaults instantiates a new AccessKeyPostResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AccessKeyPostResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessKeyPostResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessKeyPostResponse) SetId(v string)`

SetId sets Id field to given value.


### GetClientId

`func (o *AccessKeyPostResponse) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *AccessKeyPostResponse) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *AccessKeyPostResponse) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetName

`func (o *AccessKeyPostResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AccessKeyPostResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AccessKeyPostResponse) SetName(v string)`

SetName sets Name field to given value.


### GetClientSecret

`func (o *AccessKeyPostResponse) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *AccessKeyPostResponse) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *AccessKeyPostResponse) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.


### GetCreatedAt

`func (o *AccessKeyPostResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccessKeyPostResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccessKeyPostResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### SetCreatedAtNil

`func (o *AccessKeyPostResponse) SetCreatedAtNil(b bool)`

 SetCreatedAtNil sets the value for CreatedAt to be an explicit nil

### UnsetCreatedAt
`func (o *AccessKeyPostResponse) UnsetCreatedAt()`

UnsetCreatedAt ensures that no value is present for CreatedAt, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


