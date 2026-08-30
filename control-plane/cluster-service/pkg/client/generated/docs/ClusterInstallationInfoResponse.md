# ClusterInstallationInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InstallationStr** | **string** |  | 
**RepositoryName** | **string** |  | 
**ChartRepoURL** | **string** |  | 
**BootstrapToken** | **string** | Short-lived JWT (2h TTL) for the installer service account. Pass to the Helm hook via --set cluster.bootstrapToken&#x3D;&lt;token&gt;.  | 
**ClientSecret** | Pointer to **NullableString** | Deprecated. Omitted on CP &gt;&#x3D; 3.0. Use bootstrapToken instead. | [optional] 

## Methods

### NewClusterInstallationInfoResponse

`func NewClusterInstallationInfoResponse(installationStr string, repositoryName string, chartRepoURL string, bootstrapToken string, ) *ClusterInstallationInfoResponse`

NewClusterInstallationInfoResponse instantiates a new ClusterInstallationInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInstallationInfoResponseWithDefaults

`func NewClusterInstallationInfoResponseWithDefaults() *ClusterInstallationInfoResponse`

NewClusterInstallationInfoResponseWithDefaults instantiates a new ClusterInstallationInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInstallationStr

`func (o *ClusterInstallationInfoResponse) GetInstallationStr() string`

GetInstallationStr returns the InstallationStr field if non-nil, zero value otherwise.

### GetInstallationStrOk

`func (o *ClusterInstallationInfoResponse) GetInstallationStrOk() (*string, bool)`

GetInstallationStrOk returns a tuple with the InstallationStr field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstallationStr

`func (o *ClusterInstallationInfoResponse) SetInstallationStr(v string)`

SetInstallationStr sets InstallationStr field to given value.


### GetRepositoryName

`func (o *ClusterInstallationInfoResponse) GetRepositoryName() string`

GetRepositoryName returns the RepositoryName field if non-nil, zero value otherwise.

### GetRepositoryNameOk

`func (o *ClusterInstallationInfoResponse) GetRepositoryNameOk() (*string, bool)`

GetRepositoryNameOk returns a tuple with the RepositoryName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositoryName

`func (o *ClusterInstallationInfoResponse) SetRepositoryName(v string)`

SetRepositoryName sets RepositoryName field to given value.


### GetChartRepoURL

`func (o *ClusterInstallationInfoResponse) GetChartRepoURL() string`

GetChartRepoURL returns the ChartRepoURL field if non-nil, zero value otherwise.

### GetChartRepoURLOk

`func (o *ClusterInstallationInfoResponse) GetChartRepoURLOk() (*string, bool)`

GetChartRepoURLOk returns a tuple with the ChartRepoURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartRepoURL

`func (o *ClusterInstallationInfoResponse) SetChartRepoURL(v string)`

SetChartRepoURL sets ChartRepoURL field to given value.


### GetBootstrapToken

`func (o *ClusterInstallationInfoResponse) GetBootstrapToken() string`

GetBootstrapToken returns the BootstrapToken field if non-nil, zero value otherwise.

### GetBootstrapTokenOk

`func (o *ClusterInstallationInfoResponse) GetBootstrapTokenOk() (*string, bool)`

GetBootstrapTokenOk returns a tuple with the BootstrapToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBootstrapToken

`func (o *ClusterInstallationInfoResponse) SetBootstrapToken(v string)`

SetBootstrapToken sets BootstrapToken field to given value.


### GetClientSecret

`func (o *ClusterInstallationInfoResponse) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *ClusterInstallationInfoResponse) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *ClusterInstallationInfoResponse) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.

### HasClientSecret

`func (o *ClusterInstallationInfoResponse) HasClientSecret() bool`

HasClientSecret returns a boolean if a field has been set.

### SetClientSecretNil

`func (o *ClusterInstallationInfoResponse) SetClientSecretNil(b bool)`

 SetClientSecretNil sets the value for ClientSecret to be an explicit nil

### UnsetClientSecret
`func (o *ClusterInstallationInfoResponse) UnsetClientSecret()`

UnsetClientSecret ensures that no value is present for ClientSecret, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


