# AssetUsageInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Workspaces** | Pointer to [**[]WorkloadRefAndStatus**](WorkloadRefAndStatus.md) | workspaces that rely on this asset. | [optional] 
**Trainings** | Pointer to [**[]WorkloadRefAndStatus**](WorkloadRefAndStatus.md) | trainings that rely on this asset. | [optional] 
**Distributed** | Pointer to [**[]WorkloadRefAndStatus**](WorkloadRefAndStatus.md) | distributed trainings that rely on this asset. | [optional] 
**Inferences** | Pointer to [**[]WorkloadRefAndStatus**](WorkloadRefAndStatus.md) | inferences that rely on this asset. | [optional] 
**Templates** | Pointer to [**[]AssetRef**](AssetRef.md) | templates that rely on this asset. | [optional] 
**Assets** | Pointer to [**AssetsUsageRef**](AssetsUsageRef.md) |  | [optional] 

## Methods

### NewAssetUsageInfo

`func NewAssetUsageInfo() *AssetUsageInfo`

NewAssetUsageInfo instantiates a new AssetUsageInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetUsageInfoWithDefaults

`func NewAssetUsageInfoWithDefaults() *AssetUsageInfo`

NewAssetUsageInfoWithDefaults instantiates a new AssetUsageInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkspaces

`func (o *AssetUsageInfo) GetWorkspaces() []WorkloadRefAndStatus`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *AssetUsageInfo) GetWorkspacesOk() (*[]WorkloadRefAndStatus, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *AssetUsageInfo) SetWorkspaces(v []WorkloadRefAndStatus)`

SetWorkspaces sets Workspaces field to given value.

### HasWorkspaces

`func (o *AssetUsageInfo) HasWorkspaces() bool`

HasWorkspaces returns a boolean if a field has been set.

### GetTrainings

`func (o *AssetUsageInfo) GetTrainings() []WorkloadRefAndStatus`

GetTrainings returns the Trainings field if non-nil, zero value otherwise.

### GetTrainingsOk

`func (o *AssetUsageInfo) GetTrainingsOk() (*[]WorkloadRefAndStatus, bool)`

GetTrainingsOk returns a tuple with the Trainings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainings

`func (o *AssetUsageInfo) SetTrainings(v []WorkloadRefAndStatus)`

SetTrainings sets Trainings field to given value.

### HasTrainings

`func (o *AssetUsageInfo) HasTrainings() bool`

HasTrainings returns a boolean if a field has been set.

### GetDistributed

`func (o *AssetUsageInfo) GetDistributed() []WorkloadRefAndStatus`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *AssetUsageInfo) GetDistributedOk() (*[]WorkloadRefAndStatus, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *AssetUsageInfo) SetDistributed(v []WorkloadRefAndStatus)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *AssetUsageInfo) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### GetInferences

`func (o *AssetUsageInfo) GetInferences() []WorkloadRefAndStatus`

GetInferences returns the Inferences field if non-nil, zero value otherwise.

### GetInferencesOk

`func (o *AssetUsageInfo) GetInferencesOk() (*[]WorkloadRefAndStatus, bool)`

GetInferencesOk returns a tuple with the Inferences field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferences

`func (o *AssetUsageInfo) SetInferences(v []WorkloadRefAndStatus)`

SetInferences sets Inferences field to given value.

### HasInferences

`func (o *AssetUsageInfo) HasInferences() bool`

HasInferences returns a boolean if a field has been set.

### GetTemplates

`func (o *AssetUsageInfo) GetTemplates() []AssetRef`

GetTemplates returns the Templates field if non-nil, zero value otherwise.

### GetTemplatesOk

`func (o *AssetUsageInfo) GetTemplatesOk() (*[]AssetRef, bool)`

GetTemplatesOk returns a tuple with the Templates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplates

`func (o *AssetUsageInfo) SetTemplates(v []AssetRef)`

SetTemplates sets Templates field to given value.

### HasTemplates

`func (o *AssetUsageInfo) HasTemplates() bool`

HasTemplates returns a boolean if a field has been set.

### GetAssets

`func (o *AssetUsageInfo) GetAssets() AssetsUsageRef`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *AssetUsageInfo) GetAssetsOk() (*AssetsUsageRef, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *AssetUsageInfo) SetAssets(v AssetsUsageRef)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *AssetUsageInfo) HasAssets() bool`

HasAssets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


