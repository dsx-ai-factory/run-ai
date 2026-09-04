# DatasourceListResponseAssetSpecConfigMap

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMap** | **string** | The name of the ConfigMap resource. ConfigMap is mandatory for creating a workspace. | 
**MountPath** | **string** | Local path within the workspace to which the PVC bucket will be mapped. ClaimName is mandatory for creating a workspace. | 

## Methods

### NewDatasourceListResponseAssetSpecConfigMap

`func NewDatasourceListResponseAssetSpecConfigMap(configMap string, mountPath string, ) *DatasourceListResponseAssetSpecConfigMap`

NewDatasourceListResponseAssetSpecConfigMap instantiates a new DatasourceListResponseAssetSpecConfigMap object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecConfigMapWithDefaults

`func NewDatasourceListResponseAssetSpecConfigMapWithDefaults() *DatasourceListResponseAssetSpecConfigMap`

NewDatasourceListResponseAssetSpecConfigMapWithDefaults instantiates a new DatasourceListResponseAssetSpecConfigMap object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMap

`func (o *DatasourceListResponseAssetSpecConfigMap) GetConfigMap() string`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *DatasourceListResponseAssetSpecConfigMap) GetConfigMapOk() (*string, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *DatasourceListResponseAssetSpecConfigMap) SetConfigMap(v string)`

SetConfigMap sets ConfigMap field to given value.


### GetMountPath

`func (o *DatasourceListResponseAssetSpecConfigMap) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *DatasourceListResponseAssetSpecConfigMap) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *DatasourceListResponseAssetSpecConfigMap) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


