# ClusterInfoSyncRequestStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]ClusterReportedStatusConditionsInner**](ClusterReportedStatusConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Platform** | Pointer to [**NullableClusterReportedStatusPlatform**](ClusterReportedStatusPlatform.md) |  | [optional] 
**Config** | Pointer to [**NullableClusterReportedConfig**](ClusterReportedConfig.md) |  | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewClusterInfoSyncRequestStatus

`func NewClusterInfoSyncRequestStatus() *ClusterInfoSyncRequestStatus`

NewClusterInfoSyncRequestStatus instantiates a new ClusterInfoSyncRequestStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInfoSyncRequestStatusWithDefaults

`func NewClusterInfoSyncRequestStatusWithDefaults() *ClusterInfoSyncRequestStatus`

NewClusterInfoSyncRequestStatusWithDefaults instantiates a new ClusterInfoSyncRequestStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *ClusterInfoSyncRequestStatus) GetConditions() []ClusterReportedStatusConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterInfoSyncRequestStatus) GetConditionsOk() (*[]ClusterReportedStatusConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterInfoSyncRequestStatus) SetConditions(v []ClusterReportedStatusConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterInfoSyncRequestStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *ClusterInfoSyncRequestStatus) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *ClusterInfoSyncRequestStatus) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *ClusterInfoSyncRequestStatus) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *ClusterInfoSyncRequestStatus) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetPlatform

`func (o *ClusterInfoSyncRequestStatus) GetPlatform() ClusterReportedStatusPlatform`

GetPlatform returns the Platform field if non-nil, zero value otherwise.

### GetPlatformOk

`func (o *ClusterInfoSyncRequestStatus) GetPlatformOk() (*ClusterReportedStatusPlatform, bool)`

GetPlatformOk returns a tuple with the Platform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlatform

`func (o *ClusterInfoSyncRequestStatus) SetPlatform(v ClusterReportedStatusPlatform)`

SetPlatform sets Platform field to given value.

### HasPlatform

`func (o *ClusterInfoSyncRequestStatus) HasPlatform() bool`

HasPlatform returns a boolean if a field has been set.

### SetPlatformNil

`func (o *ClusterInfoSyncRequestStatus) SetPlatformNil(b bool)`

 SetPlatformNil sets the value for Platform to be an explicit nil

### UnsetPlatform
`func (o *ClusterInfoSyncRequestStatus) UnsetPlatform()`

UnsetPlatform ensures that no value is present for Platform, not even an explicit nil
### GetConfig

`func (o *ClusterInfoSyncRequestStatus) GetConfig() ClusterReportedConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ClusterInfoSyncRequestStatus) GetConfigOk() (*ClusterReportedConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ClusterInfoSyncRequestStatus) SetConfig(v ClusterReportedConfig)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *ClusterInfoSyncRequestStatus) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### SetConfigNil

`func (o *ClusterInfoSyncRequestStatus) SetConfigNil(b bool)`

 SetConfigNil sets the value for Config to be an explicit nil

### UnsetConfig
`func (o *ClusterInfoSyncRequestStatus) UnsetConfig()`

UnsetConfig ensures that no value is present for Config, not even an explicit nil
### GetDependencies

`func (o *ClusterInfoSyncRequestStatus) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterInfoSyncRequestStatus) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterInfoSyncRequestStatus) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterInfoSyncRequestStatus) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


