# ConfigMapsListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMaps** | [**[]ConfigMapListInfo**](ConfigMapListInfo.md) |  | 

## Methods

### NewConfigMapsListResponse

`func NewConfigMapsListResponse(configMaps []ConfigMapListInfo, ) *ConfigMapsListResponse`

NewConfigMapsListResponse instantiates a new ConfigMapsListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapsListResponseWithDefaults

`func NewConfigMapsListResponseWithDefaults() *ConfigMapsListResponse`

NewConfigMapsListResponseWithDefaults instantiates a new ConfigMapsListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMaps

`func (o *ConfigMapsListResponse) GetConfigMaps() []ConfigMapListInfo`

GetConfigMaps returns the ConfigMaps field if non-nil, zero value otherwise.

### GetConfigMapsOk

`func (o *ConfigMapsListResponse) GetConfigMapsOk() (*[]ConfigMapListInfo, bool)`

GetConfigMapsOk returns a tuple with the ConfigMaps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMaps

`func (o *ConfigMapsListResponse) SetConfigMaps(v []ConfigMapListInfo)`

SetConfigMaps sets ConfigMaps field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


