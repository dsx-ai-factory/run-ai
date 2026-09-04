# NodePool

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Node Pool Name | [optional] [default to null]
**OverProvisioningRatio** | **int32** |  | [optional] [default to 1]
**LabelKey** | **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] [default to null]
**LabelValue** | **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] [default to null]
**PlacementStrategy** | [***PlacementStrategy**](PlacementStrategy.md) |  | [optional] [default to null]
**Id** | **float64** | Node Pool unique id | [optional] [default to null]
**ClusterId** | **string** | Node Pool cluster id | [optional] [default to null]
**CreatedAt** | [**time.Time**](time.Time.md) | Node Pool creation time | [optional] [default to null]
**UpdatedAt** | [**time.Time**](time.Time.md) | Node Pool update time | [optional] [default to null]
**DeletedAt** | [**time.Time**](time.Time.md) | Node Pool delete time | [optional] [default to null]
**Status** | **string** | Node Pool status | [optional] [default to null]
**StatusMessage** | **string** | Node Pool status details | [optional] [default to null]
**Nodes** | **string** | List of Nodes that are assigned to this nodepool - as json string | [optional] [default to null]
**CreatedBy** | **string** | Node Pool creator | [optional] [default to null]
**UpdatedBy** | **string** | Node Pool updater | [optional] [default to null]
**IsDefault** | **bool** | Is the Node Pool the default Node Pool for all nodes not assigned to any other Node Pool | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

