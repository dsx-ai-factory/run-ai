# ClusterDisplayedStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]ClusterReportedStatusConditionsInner**](ClusterReportedStatusConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Platform** | Pointer to [**NullableClusterReportedStatusPlatform**](ClusterReportedStatusPlatform.md) |  | [optional] 
**Config** | Pointer to [**NullableClusterReportedConfig**](ClusterReportedConfig.md) |  | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 
**State** | Pointer to **string** | The cluster&#39;s state. | [optional] 

## Methods

### NewClusterDisplayedStatus

`func NewClusterDisplayedStatus() *ClusterDisplayedStatus`

NewClusterDisplayedStatus instantiates a new ClusterDisplayedStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterDisplayedStatusWithDefaults

`func NewClusterDisplayedStatusWithDefaults() *ClusterDisplayedStatus`

NewClusterDisplayedStatusWithDefaults instantiates a new ClusterDisplayedStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *ClusterDisplayedStatus) GetConditions() []ClusterReportedStatusConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterDisplayedStatus) GetConditionsOk() (*[]ClusterReportedStatusConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterDisplayedStatus) SetConditions(v []ClusterReportedStatusConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterDisplayedStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *ClusterDisplayedStatus) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *ClusterDisplayedStatus) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *ClusterDisplayedStatus) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *ClusterDisplayedStatus) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetPlatform

`func (o *ClusterDisplayedStatus) GetPlatform() ClusterReportedStatusPlatform`

GetPlatform returns the Platform field if non-nil, zero value otherwise.

### GetPlatformOk

`func (o *ClusterDisplayedStatus) GetPlatformOk() (*ClusterReportedStatusPlatform, bool)`

GetPlatformOk returns a tuple with the Platform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlatform

`func (o *ClusterDisplayedStatus) SetPlatform(v ClusterReportedStatusPlatform)`

SetPlatform sets Platform field to given value.

### HasPlatform

`func (o *ClusterDisplayedStatus) HasPlatform() bool`

HasPlatform returns a boolean if a field has been set.

### SetPlatformNil

`func (o *ClusterDisplayedStatus) SetPlatformNil(b bool)`

 SetPlatformNil sets the value for Platform to be an explicit nil

### UnsetPlatform
`func (o *ClusterDisplayedStatus) UnsetPlatform()`

UnsetPlatform ensures that no value is present for Platform, not even an explicit nil
### GetConfig

`func (o *ClusterDisplayedStatus) GetConfig() ClusterReportedConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ClusterDisplayedStatus) GetConfigOk() (*ClusterReportedConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ClusterDisplayedStatus) SetConfig(v ClusterReportedConfig)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *ClusterDisplayedStatus) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### SetConfigNil

`func (o *ClusterDisplayedStatus) SetConfigNil(b bool)`

 SetConfigNil sets the value for Config to be an explicit nil

### UnsetConfig
`func (o *ClusterDisplayedStatus) UnsetConfig()`

UnsetConfig ensures that no value is present for Config, not even an explicit nil
### GetDependencies

`func (o *ClusterDisplayedStatus) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterDisplayedStatus) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterDisplayedStatus) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterDisplayedStatus) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.

### GetState

`func (o *ClusterDisplayedStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClusterDisplayedStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClusterDisplayedStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ClusterDisplayedStatus) HasState() bool`

HasState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


