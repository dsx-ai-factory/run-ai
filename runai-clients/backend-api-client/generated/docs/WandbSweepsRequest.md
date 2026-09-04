# WandbSweepsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiKey** | **string** | Wandb API key. | 
**BaseUrl** | Pointer to **NullableString** | Wandb host URL, to be provided in case of a self-hosted W&amp;B installation. | [optional] 
**Project** | **string** | Wandb project. | 
**Config** | **map[string]interface{}** | Sweep config JSON object. | 

## Methods

### NewWandbSweepsRequest

`func NewWandbSweepsRequest(apiKey string, project string, config map[string]interface{}, ) *WandbSweepsRequest`

NewWandbSweepsRequest instantiates a new WandbSweepsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWandbSweepsRequestWithDefaults

`func NewWandbSweepsRequestWithDefaults() *WandbSweepsRequest`

NewWandbSweepsRequestWithDefaults instantiates a new WandbSweepsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiKey

`func (o *WandbSweepsRequest) GetApiKey() string`

GetApiKey returns the ApiKey field if non-nil, zero value otherwise.

### GetApiKeyOk

`func (o *WandbSweepsRequest) GetApiKeyOk() (*string, bool)`

GetApiKeyOk returns a tuple with the ApiKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiKey

`func (o *WandbSweepsRequest) SetApiKey(v string)`

SetApiKey sets ApiKey field to given value.


### GetBaseUrl

`func (o *WandbSweepsRequest) GetBaseUrl() string`

GetBaseUrl returns the BaseUrl field if non-nil, zero value otherwise.

### GetBaseUrlOk

`func (o *WandbSweepsRequest) GetBaseUrlOk() (*string, bool)`

GetBaseUrlOk returns a tuple with the BaseUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseUrl

`func (o *WandbSweepsRequest) SetBaseUrl(v string)`

SetBaseUrl sets BaseUrl field to given value.

### HasBaseUrl

`func (o *WandbSweepsRequest) HasBaseUrl() bool`

HasBaseUrl returns a boolean if a field has been set.

### SetBaseUrlNil

`func (o *WandbSweepsRequest) SetBaseUrlNil(b bool)`

 SetBaseUrlNil sets the value for BaseUrl to be an explicit nil

### UnsetBaseUrl
`func (o *WandbSweepsRequest) UnsetBaseUrl()`

UnsetBaseUrl ensures that no value is present for BaseUrl, not even an explicit nil
### GetProject

`func (o *WandbSweepsRequest) GetProject() string`

GetProject returns the Project field if non-nil, zero value otherwise.

### GetProjectOk

`func (o *WandbSweepsRequest) GetProjectOk() (*string, bool)`

GetProjectOk returns a tuple with the Project field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProject

`func (o *WandbSweepsRequest) SetProject(v string)`

SetProject sets Project field to given value.


### GetConfig

`func (o *WandbSweepsRequest) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *WandbSweepsRequest) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *WandbSweepsRequest) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


