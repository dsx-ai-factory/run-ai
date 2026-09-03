# InstallerBootstrapResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Token** | **string** | Short-lived Keycloak JWT (2h TTL) for the installer service account. Pass as Authorization Bearer to GET /clusters/{uuid}/control-plane-config. | 

## Methods

### NewInstallerBootstrapResponse

`func NewInstallerBootstrapResponse(token string, ) *InstallerBootstrapResponse`

NewInstallerBootstrapResponse instantiates a new InstallerBootstrapResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstallerBootstrapResponseWithDefaults

`func NewInstallerBootstrapResponseWithDefaults() *InstallerBootstrapResponse`

NewInstallerBootstrapResponseWithDefaults instantiates a new InstallerBootstrapResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToken

`func (o *InstallerBootstrapResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *InstallerBootstrapResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *InstallerBootstrapResponse) SetToken(v string)`

SetToken sets Token field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


