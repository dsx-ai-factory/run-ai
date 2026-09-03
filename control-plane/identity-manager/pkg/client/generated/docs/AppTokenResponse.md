# AppTokenResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessToken** | **string** |  | 
**IdToken** | **string** |  | 

## Methods

### NewAppTokenResponse

`func NewAppTokenResponse(accessToken string, idToken string, ) *AppTokenResponse`

NewAppTokenResponse instantiates a new AppTokenResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAppTokenResponseWithDefaults

`func NewAppTokenResponseWithDefaults() *AppTokenResponse`

NewAppTokenResponseWithDefaults instantiates a new AppTokenResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessToken

`func (o *AppTokenResponse) GetAccessToken() string`

GetAccessToken returns the AccessToken field if non-nil, zero value otherwise.

### GetAccessTokenOk

`func (o *AppTokenResponse) GetAccessTokenOk() (*string, bool)`

GetAccessTokenOk returns a tuple with the AccessToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessToken

`func (o *AppTokenResponse) SetAccessToken(v string)`

SetAccessToken sets AccessToken field to given value.


### GetIdToken

`func (o *AppTokenResponse) GetIdToken() string`

GetIdToken returns the IdToken field if non-nil, zero value otherwise.

### GetIdTokenOk

`func (o *AppTokenResponse) GetIdTokenOk() (*string, bool)`

GetIdTokenOk returns a tuple with the IdToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdToken

`func (o *AppTokenResponse) SetIdToken(v string)`

SetIdToken sets IdToken field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


