# ClusterControlPlaneConfigResearcherService

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**JwksURL** | Pointer to **string** | URL to the endpoint that returns JWT validation keys | [optional] 
**LogoutLink** | Pointer to **string** | Logout link. DEPRECARED from 2.9 | [optional] 
**OauthProxy** | Pointer to [**ClusterControlPlaneConfigResearcherServiceOauthProxy**](ClusterControlPlaneConfigResearcherServiceOauthProxy.md) |  | [optional] 

## Methods

### NewClusterControlPlaneConfigResearcherService

`func NewClusterControlPlaneConfigResearcherService() *ClusterControlPlaneConfigResearcherService`

NewClusterControlPlaneConfigResearcherService instantiates a new ClusterControlPlaneConfigResearcherService object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterControlPlaneConfigResearcherServiceWithDefaults

`func NewClusterControlPlaneConfigResearcherServiceWithDefaults() *ClusterControlPlaneConfigResearcherService`

NewClusterControlPlaneConfigResearcherServiceWithDefaults instantiates a new ClusterControlPlaneConfigResearcherService object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetJwksURL

`func (o *ClusterControlPlaneConfigResearcherService) GetJwksURL() string`

GetJwksURL returns the JwksURL field if non-nil, zero value otherwise.

### GetJwksURLOk

`func (o *ClusterControlPlaneConfigResearcherService) GetJwksURLOk() (*string, bool)`

GetJwksURLOk returns a tuple with the JwksURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwksURL

`func (o *ClusterControlPlaneConfigResearcherService) SetJwksURL(v string)`

SetJwksURL sets JwksURL field to given value.

### HasJwksURL

`func (o *ClusterControlPlaneConfigResearcherService) HasJwksURL() bool`

HasJwksURL returns a boolean if a field has been set.

### GetLogoutLink

`func (o *ClusterControlPlaneConfigResearcherService) GetLogoutLink() string`

GetLogoutLink returns the LogoutLink field if non-nil, zero value otherwise.

### GetLogoutLinkOk

`func (o *ClusterControlPlaneConfigResearcherService) GetLogoutLinkOk() (*string, bool)`

GetLogoutLinkOk returns a tuple with the LogoutLink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogoutLink

`func (o *ClusterControlPlaneConfigResearcherService) SetLogoutLink(v string)`

SetLogoutLink sets LogoutLink field to given value.

### HasLogoutLink

`func (o *ClusterControlPlaneConfigResearcherService) HasLogoutLink() bool`

HasLogoutLink returns a boolean if a field has been set.

### GetOauthProxy

`func (o *ClusterControlPlaneConfigResearcherService) GetOauthProxy() ClusterControlPlaneConfigResearcherServiceOauthProxy`

GetOauthProxy returns the OauthProxy field if non-nil, zero value otherwise.

### GetOauthProxyOk

`func (o *ClusterControlPlaneConfigResearcherService) GetOauthProxyOk() (*ClusterControlPlaneConfigResearcherServiceOauthProxy, bool)`

GetOauthProxyOk returns a tuple with the OauthProxy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOauthProxy

`func (o *ClusterControlPlaneConfigResearcherService) SetOauthProxy(v ClusterControlPlaneConfigResearcherServiceOauthProxy)`

SetOauthProxy sets OauthProxy field to given value.

### HasOauthProxy

`func (o *ClusterControlPlaneConfigResearcherService) HasOauthProxy() bool`

HasOauthProxy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


