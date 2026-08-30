# NodePoolResponse

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float64** | The NodePool id | [optional] [default to null]
**LabelKey** | **string** | Key of node label for pool | [optional] [default to null]
**LabelValue** | **string** | Value of node label for pool | [optional] [default to null]
**OverProvisioningRatio** | **int32** |  | [optional] [default to 1]
**PlacementStrategy** | [***PlacementStrategy**](PlacementStrategy.md) |  | [optional] [default to null]
**IsDefault** | **bool** | Is the Node Pool default for cluster | [optional] [default to null]
**Status** | **string** | Status of Node Pool | [optional] [default to null]
**StatusMessage** | **string** | Message for status of Node Pool | [optional] [default to null]
**Nodes** | **string** | List of Nodes that are assigned to this nodepool - as json string | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

