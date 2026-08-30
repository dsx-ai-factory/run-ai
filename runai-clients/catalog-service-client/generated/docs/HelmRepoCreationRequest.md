# HelmRepoCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the helm repo. | 
**Url** | **string** | The URL of the helm repo. | 
**IsPrivate** | Pointer to **bool** | Whether the helm repo requires authentication to browse. | [optional] 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**Description** | Pointer to **string** | The description of the helm repo. | [optional] 

## Methods

### NewHelmRepoCreationRequest

`func NewHelmRepoCreationRequest(name string, url string, scopeType ScopeType, scopeId string, ) *HelmRepoCreationRequest`

NewHelmRepoCreationRequest instantiates a new HelmRepoCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmRepoCreationRequestWithDefaults

`func NewHelmRepoCreationRequestWithDefaults() *HelmRepoCreationRequest`

NewHelmRepoCreationRequestWithDefaults instantiates a new HelmRepoCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *HelmRepoCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *HelmRepoCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *HelmRepoCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUrl

`func (o *HelmRepoCreationRequest) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *HelmRepoCreationRequest) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *HelmRepoCreationRequest) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetIsPrivate

`func (o *HelmRepoCreationRequest) GetIsPrivate() bool`

GetIsPrivate returns the IsPrivate field if non-nil, zero value otherwise.

### GetIsPrivateOk

`func (o *HelmRepoCreationRequest) GetIsPrivateOk() (*bool, bool)`

GetIsPrivateOk returns a tuple with the IsPrivate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPrivate

`func (o *HelmRepoCreationRequest) SetIsPrivate(v bool)`

SetIsPrivate sets IsPrivate field to given value.

### HasIsPrivate

`func (o *HelmRepoCreationRequest) HasIsPrivate() bool`

HasIsPrivate returns a boolean if a field has been set.

### GetScopeType

`func (o *HelmRepoCreationRequest) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *HelmRepoCreationRequest) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *HelmRepoCreationRequest) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *HelmRepoCreationRequest) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *HelmRepoCreationRequest) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *HelmRepoCreationRequest) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetDescription

`func (o *HelmRepoCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *HelmRepoCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *HelmRepoCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *HelmRepoCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


