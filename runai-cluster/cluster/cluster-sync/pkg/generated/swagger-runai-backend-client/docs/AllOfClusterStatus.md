# AllOfClusterStatus

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | [**[]ClusterReportedStatusConditions**](ClusterReportedStatus_conditions.md) |  | [optional] [default to null]
**Operands** | [**map[string]ClusterReportedStatusOperands**](ClusterReportedStatus_operands.md) | OperandStatuses specifies the status of the managed operands | [optional] [default to null]
**Dependencies** | [***ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] [default to null]
**State** | **string** | The cluster&#x27;s state. | [optional] [default to null]
**LastConnected** | [**time.Time**](time.Time.md) | The last time the cluster was connected to the control plane. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

