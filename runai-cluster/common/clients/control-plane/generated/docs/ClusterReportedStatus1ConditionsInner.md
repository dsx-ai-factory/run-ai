# ClusterReportedStatus1ConditionsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastTransitionTime** | **time.Time** | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. | 
**Message** | **string** | message is a human readable message indicating details about the transition. This may be an empty string. | 
**ObservedGeneration** | Pointer to **int64** | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. | [optional] 
**Reason** | **string** | reason contains a programmatic identifier indicating the reason for the condition&#39;s last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. | 
**Status** | **string** | status of the condition, one of True, False, Unknown. | 
**Type** | **string** | type of condition in CamelCase or in foo.example.com/CamelCase. --- Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be useful (see .node.status.conditions), the ability to deconflict is important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt) | 

## Methods

### NewClusterReportedStatus1ConditionsInner

`func NewClusterReportedStatus1ConditionsInner(lastTransitionTime time.Time, message string, reason string, status string, type_ string, ) *ClusterReportedStatus1ConditionsInner`

NewClusterReportedStatus1ConditionsInner instantiates a new ClusterReportedStatus1ConditionsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedStatus1ConditionsInnerWithDefaults

`func NewClusterReportedStatus1ConditionsInnerWithDefaults() *ClusterReportedStatus1ConditionsInner`

NewClusterReportedStatus1ConditionsInnerWithDefaults instantiates a new ClusterReportedStatus1ConditionsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastTransitionTime

`func (o *ClusterReportedStatus1ConditionsInner) GetLastTransitionTime() time.Time`

GetLastTransitionTime returns the LastTransitionTime field if non-nil, zero value otherwise.

### GetLastTransitionTimeOk

`func (o *ClusterReportedStatus1ConditionsInner) GetLastTransitionTimeOk() (*time.Time, bool)`

GetLastTransitionTimeOk returns a tuple with the LastTransitionTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastTransitionTime

`func (o *ClusterReportedStatus1ConditionsInner) SetLastTransitionTime(v time.Time)`

SetLastTransitionTime sets LastTransitionTime field to given value.


### GetMessage

`func (o *ClusterReportedStatus1ConditionsInner) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ClusterReportedStatus1ConditionsInner) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ClusterReportedStatus1ConditionsInner) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetObservedGeneration

`func (o *ClusterReportedStatus1ConditionsInner) GetObservedGeneration() int64`

GetObservedGeneration returns the ObservedGeneration field if non-nil, zero value otherwise.

### GetObservedGenerationOk

`func (o *ClusterReportedStatus1ConditionsInner) GetObservedGenerationOk() (*int64, bool)`

GetObservedGenerationOk returns a tuple with the ObservedGeneration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedGeneration

`func (o *ClusterReportedStatus1ConditionsInner) SetObservedGeneration(v int64)`

SetObservedGeneration sets ObservedGeneration field to given value.

### HasObservedGeneration

`func (o *ClusterReportedStatus1ConditionsInner) HasObservedGeneration() bool`

HasObservedGeneration returns a boolean if a field has been set.

### GetReason

`func (o *ClusterReportedStatus1ConditionsInner) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ClusterReportedStatus1ConditionsInner) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ClusterReportedStatus1ConditionsInner) SetReason(v string)`

SetReason sets Reason field to given value.


### GetStatus

`func (o *ClusterReportedStatus1ConditionsInner) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ClusterReportedStatus1ConditionsInner) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ClusterReportedStatus1ConditionsInner) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetType

`func (o *ClusterReportedStatus1ConditionsInner) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ClusterReportedStatus1ConditionsInner) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ClusterReportedStatus1ConditionsInner) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


