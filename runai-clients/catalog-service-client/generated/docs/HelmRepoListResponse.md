# HelmRepoListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HelmRepos** | [**[]HelmRepo**](HelmRepo.md) | The list of helm repos. | 

## Methods

### NewHelmRepoListResponse

`func NewHelmRepoListResponse(helmRepos []HelmRepo, ) *HelmRepoListResponse`

NewHelmRepoListResponse instantiates a new HelmRepoListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmRepoListResponseWithDefaults

`func NewHelmRepoListResponseWithDefaults() *HelmRepoListResponse`

NewHelmRepoListResponseWithDefaults instantiates a new HelmRepoListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHelmRepos

`func (o *HelmRepoListResponse) GetHelmRepos() []HelmRepo`

GetHelmRepos returns the HelmRepos field if non-nil, zero value otherwise.

### GetHelmReposOk

`func (o *HelmRepoListResponse) GetHelmReposOk() (*[]HelmRepo, bool)`

GetHelmReposOk returns a tuple with the HelmRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHelmRepos

`func (o *HelmRepoListResponse) SetHelmRepos(v []HelmRepo)`

SetHelmRepos sets HelmRepos field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


