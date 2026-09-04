# Pod

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PodId** | **string** | Identifier of the pod running the job. | [default to null]
**JobId** | **string** | Unique identifier of the job. | [default to null]
**PodGroupId** | **string** | This had been used as jobId. Remained here for backward compatibility | [optional] [default to null]
**ClusterUuid** | **string** | Unique identifier of the cluster. | [default to null]
**PodName** | **string** | The name of the pod running the job. | [default to null]
**ImageName** | **string** | The name of the image executed by the pod. | [default to null]
**NodeId** | **string** | Unique identifier of the node. | [optional] [default to null]
**Phase** | **string** |  | [default to null]
**Status** | **string** |  | [optional] [default to null]
**Created** | **int64** | Creation time of the pod. | [default to null]
**Completed** | **int64** | Completion time of the pod. | [default to null]
**Started** | **int64** | The time when the pod started executing. | [optional] [default to null]
**LastUpdated** | **int64** | Last time the pod details were updated. | [default to null]
**DynamicData** | [***interface{}**](interface{}.md) |  | [optional] [default to null]
**ExistsInCluster** | **bool** |  | [optional] [default to null]
**ResourceRequest** | **map[string]float64** |  | [optional] [default to null]
**ResourceAllocation** | **map[string]float64** |  | [optional] [default to null]
**NodePool** | **string** | The node pool of the pod. | [optional] [default to null]
**Namespace** | **string** | The namespace of the pod. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

