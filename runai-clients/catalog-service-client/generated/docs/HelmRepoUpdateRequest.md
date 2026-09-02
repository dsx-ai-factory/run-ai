# HelmRepoUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | Pointer to **string** | The URL of the helm repo. | [optional] 
**Name** | Pointer to **string** | The name of the helm repo. | [optional] 
**IsPrivate** | Pointer to **NullableBool** | Whether the helm repo requires authentication to browse. | [optional] 

## Methods

### NewHelmRepoUpdateRequest

`func NewHelmRepoUpdateRequest() *HelmRepoUpdateRequest`

NewHelmRepoUpdateRequest instantiates a new HelmRepoUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmRepoUpdateRequestWithDefaults

`func NewHelmRepoUpdateRequestWithDefaults() *HelmRepoUpdateRequest`

NewHelmRepoUpdateRequestWithDefaults instantiates a new HelmRepoUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *HelmRepoUpdateRequest) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *HelmRepoUpdateRequest) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *HelmRepoUpdateRequest) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *HelmRepoUpdateRequest) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetName

`func (o *HelmRepoUpdateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *HelmRepoUpdateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *HelmRepoUpdateRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *HelmRepoUpdateRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetIsPrivate

`func (o *HelmRepoUpdateRequest) GetIsPrivate() bool`

GetIsPrivate returns the IsPrivate field if non-nil, zero value otherwise.

### GetIsPrivateOk

`func (o *HelmRepoUpdateRequest) GetIsPrivateOk() (*bool, bool)`

GetIsPrivateOk returns a tuple with the IsPrivate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPrivate

`func (o *HelmRepoUpdateRequest) SetIsPrivate(v bool)`

SetIsPrivate sets IsPrivate field to given value.

### HasIsPrivate

`func (o *HelmRepoUpdateRequest) HasIsPrivate() bool`

HasIsPrivate returns a boolean if a field has been set.

### SetIsPrivateNil

`func (o *HelmRepoUpdateRequest) SetIsPrivateNil(b bool)`

 SetIsPrivateNil sets the value for IsPrivate to be an explicit nil

### UnsetIsPrivate
`func (o *HelmRepoUpdateRequest) UnsetIsPrivate()`

UnsetIsPrivate ensures that no value is present for IsPrivate, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


