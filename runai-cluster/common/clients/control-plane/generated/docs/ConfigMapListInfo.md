# ConfigMapListInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**Name** | **string** | name of the ConfigMap. | 

## Methods

### NewConfigMapListInfo

`func NewConfigMapListInfo(clusterId string, scopeType string, scopeId string, name string, ) *ConfigMapListInfo`

NewConfigMapListInfo instantiates a new ConfigMapListInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapListInfoWithDefaults

`func NewConfigMapListInfoWithDefaults() *ConfigMapListInfo`

NewConfigMapListInfoWithDefaults instantiates a new ConfigMapListInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *ConfigMapListInfo) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ConfigMapListInfo) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ConfigMapListInfo) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetScopeType

`func (o *ConfigMapListInfo) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *ConfigMapListInfo) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *ConfigMapListInfo) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *ConfigMapListInfo) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *ConfigMapListInfo) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *ConfigMapListInfo) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *ConfigMapListInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConfigMapListInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConfigMapListInfo) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


