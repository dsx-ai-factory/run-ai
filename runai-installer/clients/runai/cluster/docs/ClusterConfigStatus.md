# ClusterConfigStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActualConfig** | Pointer to [**ClusterUserConfigData**](ClusterUserConfigData.md) |  | [optional] 
**State** | Pointer to **string** | The state of the configuration. The possible values are: Updating - The control plane got a request to update the cluster configuration and did not receive a status update yet from the cluster. Applying - The configuration has been stored on the cluster, but is not yet applied. Applied - The configuration has been successfully applied in the cluster. Failed - The cluster has failed to apply the configuration.  | [optional] 
**StateMessage** | Pointer to **string** | A message describing the state of the configuration. | [optional] 

## Methods

### NewClusterConfigStatus

`func NewClusterConfigStatus() *ClusterConfigStatus`

NewClusterConfigStatus instantiates a new ClusterConfigStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterConfigStatusWithDefaults

`func NewClusterConfigStatusWithDefaults() *ClusterConfigStatus`

NewClusterConfigStatusWithDefaults instantiates a new ClusterConfigStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActualConfig

`func (o *ClusterConfigStatus) GetActualConfig() ClusterUserConfigData`

GetActualConfig returns the ActualConfig field if non-nil, zero value otherwise.

### GetActualConfigOk

`func (o *ClusterConfigStatus) GetActualConfigOk() (*ClusterUserConfigData, bool)`

GetActualConfigOk returns a tuple with the ActualConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualConfig

`func (o *ClusterConfigStatus) SetActualConfig(v ClusterUserConfigData)`

SetActualConfig sets ActualConfig field to given value.

### HasActualConfig

`func (o *ClusterConfigStatus) HasActualConfig() bool`

HasActualConfig returns a boolean if a field has been set.

### GetState

`func (o *ClusterConfigStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClusterConfigStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClusterConfigStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ClusterConfigStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetStateMessage

`func (o *ClusterConfigStatus) GetStateMessage() string`

GetStateMessage returns the StateMessage field if non-nil, zero value otherwise.

### GetStateMessageOk

`func (o *ClusterConfigStatus) GetStateMessageOk() (*string, bool)`

GetStateMessageOk returns a tuple with the StateMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateMessage

`func (o *ClusterConfigStatus) SetStateMessage(v string)`

SetStateMessage sets StateMessage field to given value.

### HasStateMessage

`func (o *ClusterConfigStatus) HasStateMessage() bool`

HasStateMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


