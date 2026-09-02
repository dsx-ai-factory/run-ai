# ClusterReportedStatusOperands

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastTransitionTime** | [**time.Time**](time.Time.md) | LastTransitionTime specifies the last time the operand readiness changed | [optional] [default to null]
**Ready** | **bool** | Ready specifies if the operand is ready or not | [default to null]
**Reasons** | **[]string** | Reasons specifies the reasons why the operand is not ready | [optional] [default to null]
**UnreadyThresholdCrossed** | **bool** | UnreadyThresholdCrossed specifies if the operand has been unready for longer than the threshold | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

