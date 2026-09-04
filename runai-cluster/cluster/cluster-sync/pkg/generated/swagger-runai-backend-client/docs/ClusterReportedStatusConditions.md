# ClusterReportedStatusConditions

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastTransitionTime** | [**time.Time**](time.Time.md) | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. | [default to null]
**Message** | **string** | message is a human readable message indicating details about the transition. This may be an empty string. | [default to null]
**ObservedGeneration** | **int64** | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. | [optional] [default to null]
**Reason** | **string** | reason contains a programmatic identifier indicating the reason for the condition&#x27;s last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. | [default to null]
**Status** | **string** | status of the condition, one of True, False, Unknown. | [default to null]
**Type_** | **string** | type of condition in CamelCase or in foo.example.com/CamelCase. --- Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be useful (see .node.status.conditions), the ability to deconflict is important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt) | [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

