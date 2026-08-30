# NodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Node Pool Name | [optional] 
**OverProvisioningRatio** | Pointer to **int32** |  | [optional] [default to 1]
**LabelKey** | Pointer to **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] 
**LabelValue** | Pointer to **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] 
**PlacementStrategy** | Pointer to [**PlacementStrategy**](PlacementStrategy.md) |  | [optional] 
**Id** | Pointer to **float32** | Node Pool unique id | [optional] [readonly] 
**ClusterId** | Pointer to **string** | Node Pool cluster id | [optional] 
**CreatedAt** | Pointer to **time.Time** | Node Pool creation time | [optional] 
**UpdatedAt** | Pointer to **time.Time** | Node Pool update time | [optional] 
**DeletedAt** | Pointer to **time.Time** | Node Pool delete time | [optional] 
**Status** | Pointer to **string** | Node Pool status | [optional] 
**StatusMessage** | Pointer to **string** | Node Pool status details | [optional] 
**Nodes** | Pointer to **string** | List of Nodes that are assigned to this nodepool - as json string | [optional] 
**CreatedBy** | Pointer to **string** | Node Pool creator | [optional] 
**UpdatedBy** | Pointer to **string** | Node Pool updater | [optional] 
**IsDefault** | Pointer to **bool** | Is the Node Pool the default Node Pool for all nodes not assigned to any other Node Pool | [optional] 

## Methods

### NewNodePool

`func NewNodePool() *NodePool`

NewNodePool instantiates a new NodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolWithDefaults

`func NewNodePoolWithDefaults() *NodePool`

NewNodePoolWithDefaults instantiates a new NodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodePool) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodePool) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodePool) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NodePool) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *NodePool) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *NodePool) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *NodePool) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *NodePool) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### GetLabelKey

`func (o *NodePool) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodePool) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodePool) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodePool) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *NodePool) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodePool) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodePool) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodePool) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetPlacementStrategy

`func (o *NodePool) GetPlacementStrategy() PlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *NodePool) GetPlacementStrategyOk() (*PlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *NodePool) SetPlacementStrategy(v PlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *NodePool) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetId

`func (o *NodePool) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePool) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePool) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *NodePool) HasId() bool`

HasId returns a boolean if a field has been set.

### GetClusterId

`func (o *NodePool) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodePool) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodePool) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *NodePool) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *NodePool) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *NodePool) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *NodePool) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *NodePool) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *NodePool) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *NodePool) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *NodePool) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *NodePool) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetDeletedAt

`func (o *NodePool) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *NodePool) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *NodePool) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *NodePool) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### GetStatus

`func (o *NodePool) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodePool) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodePool) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NodePool) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetStatusMessage

`func (o *NodePool) GetStatusMessage() string`

GetStatusMessage returns the StatusMessage field if non-nil, zero value otherwise.

### GetStatusMessageOk

`func (o *NodePool) GetStatusMessageOk() (*string, bool)`

GetStatusMessageOk returns a tuple with the StatusMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusMessage

`func (o *NodePool) SetStatusMessage(v string)`

SetStatusMessage sets StatusMessage field to given value.

### HasStatusMessage

`func (o *NodePool) HasStatusMessage() bool`

HasStatusMessage returns a boolean if a field has been set.

### GetNodes

`func (o *NodePool) GetNodes() string`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NodePool) GetNodesOk() (*string, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NodePool) SetNodes(v string)`

SetNodes sets Nodes field to given value.

### HasNodes

`func (o *NodePool) HasNodes() bool`

HasNodes returns a boolean if a field has been set.

### GetCreatedBy

`func (o *NodePool) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *NodePool) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *NodePool) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *NodePool) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *NodePool) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *NodePool) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *NodePool) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *NodePool) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetIsDefault

`func (o *NodePool) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *NodePool) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *NodePool) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.

### HasIsDefault

`func (o *NodePool) HasIsDefault() bool`

HasIsDefault returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


