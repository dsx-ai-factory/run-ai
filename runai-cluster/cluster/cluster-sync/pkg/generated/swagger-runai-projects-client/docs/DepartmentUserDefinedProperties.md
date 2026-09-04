# DepartmentUserDefinedProperties

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the department. | [optional] [default to null]
**DeservedGpus** | **float64** | Deserved GPUs for the department. | [optional] [default to null]
**AllowOverQuota** | **bool** | Is over quota allowed for the department. | [optional] [default to null]
**MaxAllowedGpus** | **float64** | Max allowed GPUs for the department. | [optional] [default to null]
**Resources** | [***AssignedResources**](AssignedResources.md) |  | [optional] [default to null]
**NodePoolsResources** | [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to the Department per node pool | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

