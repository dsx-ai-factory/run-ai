# ClusterReportedStatusOperandsValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastTransitionTime** | Pointer to **time.Time** | LastTransitionTime specifies the last time the operand readiness changed | [optional] 
**Ready** | **bool** | Ready specifies if the operand is ready or not | 
**Reasons** | Pointer to **[]string** | Reasons specifies the reasons why the operand is not ready | [optional] 
**UnreadyThresholdCrossed** | Pointer to **bool** | UnreadyThresholdCrossed specifies if the operand has been unready for longer than the threshold | [optional] 

## Methods

### NewClusterReportedStatusOperandsValue

`func NewClusterReportedStatusOperandsValue(ready bool, ) *ClusterReportedStatusOperandsValue`

NewClusterReportedStatusOperandsValue instantiates a new ClusterReportedStatusOperandsValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedStatusOperandsValueWithDefaults

`func NewClusterReportedStatusOperandsValueWithDefaults() *ClusterReportedStatusOperandsValue`

NewClusterReportedStatusOperandsValueWithDefaults instantiates a new ClusterReportedStatusOperandsValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastTransitionTime

`func (o *ClusterReportedStatusOperandsValue) GetLastTransitionTime() time.Time`

GetLastTransitionTime returns the LastTransitionTime field if non-nil, zero value otherwise.

### GetLastTransitionTimeOk

`func (o *ClusterReportedStatusOperandsValue) GetLastTransitionTimeOk() (*time.Time, bool)`

GetLastTransitionTimeOk returns a tuple with the LastTransitionTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastTransitionTime

`func (o *ClusterReportedStatusOperandsValue) SetLastTransitionTime(v time.Time)`

SetLastTransitionTime sets LastTransitionTime field to given value.

### HasLastTransitionTime

`func (o *ClusterReportedStatusOperandsValue) HasLastTransitionTime() bool`

HasLastTransitionTime returns a boolean if a field has been set.

### GetReady

`func (o *ClusterReportedStatusOperandsValue) GetReady() bool`

GetReady returns the Ready field if non-nil, zero value otherwise.

### GetReadyOk

`func (o *ClusterReportedStatusOperandsValue) GetReadyOk() (*bool, bool)`

GetReadyOk returns a tuple with the Ready field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReady

`func (o *ClusterReportedStatusOperandsValue) SetReady(v bool)`

SetReady sets Ready field to given value.


### GetReasons

`func (o *ClusterReportedStatusOperandsValue) GetReasons() []string`

GetReasons returns the Reasons field if non-nil, zero value otherwise.

### GetReasonsOk

`func (o *ClusterReportedStatusOperandsValue) GetReasonsOk() (*[]string, bool)`

GetReasonsOk returns a tuple with the Reasons field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReasons

`func (o *ClusterReportedStatusOperandsValue) SetReasons(v []string)`

SetReasons sets Reasons field to given value.

### HasReasons

`func (o *ClusterReportedStatusOperandsValue) HasReasons() bool`

HasReasons returns a boolean if a field has been set.

### GetUnreadyThresholdCrossed

`func (o *ClusterReportedStatusOperandsValue) GetUnreadyThresholdCrossed() bool`

GetUnreadyThresholdCrossed returns the UnreadyThresholdCrossed field if non-nil, zero value otherwise.

### GetUnreadyThresholdCrossedOk

`func (o *ClusterReportedStatusOperandsValue) GetUnreadyThresholdCrossedOk() (*bool, bool)`

GetUnreadyThresholdCrossedOk returns a tuple with the UnreadyThresholdCrossed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnreadyThresholdCrossed

`func (o *ClusterReportedStatusOperandsValue) SetUnreadyThresholdCrossed(v bool)`

SetUnreadyThresholdCrossed sets UnreadyThresholdCrossed field to given value.

### HasUnreadyThresholdCrossed

`func (o *ClusterReportedStatusOperandsValue) HasUnreadyThresholdCrossed() bool`

HasUnreadyThresholdCrossed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


