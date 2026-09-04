# ClusterSyncReportUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | a unique id of the entity being updated (asset, workload, policy, etc). | [optional] 
**UpdateCount** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 
**UpdateAttemptTime** | Pointer to **time.Time** | the time in which the cluster attempted the update. | [optional] 
**Message** | Pointer to **NullableString** | optional failure message for the case that the update has failed. should be sent null if the update was successful. | [optional] 

## Methods

### NewClusterSyncReportUpdate

`func NewClusterSyncReportUpdate() *ClusterSyncReportUpdate`

NewClusterSyncReportUpdate instantiates a new ClusterSyncReportUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterSyncReportUpdateWithDefaults

`func NewClusterSyncReportUpdateWithDefaults() *ClusterSyncReportUpdate`

NewClusterSyncReportUpdateWithDefaults instantiates a new ClusterSyncReportUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ClusterSyncReportUpdate) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ClusterSyncReportUpdate) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ClusterSyncReportUpdate) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ClusterSyncReportUpdate) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUpdateCount

`func (o *ClusterSyncReportUpdate) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *ClusterSyncReportUpdate) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *ClusterSyncReportUpdate) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *ClusterSyncReportUpdate) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.

### GetUpdateAttemptTime

`func (o *ClusterSyncReportUpdate) GetUpdateAttemptTime() time.Time`

GetUpdateAttemptTime returns the UpdateAttemptTime field if non-nil, zero value otherwise.

### GetUpdateAttemptTimeOk

`func (o *ClusterSyncReportUpdate) GetUpdateAttemptTimeOk() (*time.Time, bool)`

GetUpdateAttemptTimeOk returns a tuple with the UpdateAttemptTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateAttemptTime

`func (o *ClusterSyncReportUpdate) SetUpdateAttemptTime(v time.Time)`

SetUpdateAttemptTime sets UpdateAttemptTime field to given value.

### HasUpdateAttemptTime

`func (o *ClusterSyncReportUpdate) HasUpdateAttemptTime() bool`

HasUpdateAttemptTime returns a boolean if a field has been set.

### GetMessage

`func (o *ClusterSyncReportUpdate) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ClusterSyncReportUpdate) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ClusterSyncReportUpdate) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ClusterSyncReportUpdate) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### SetMessageNil

`func (o *ClusterSyncReportUpdate) SetMessageNil(b bool)`

 SetMessageNil sets the value for Message to be an explicit nil

### UnsetMessage
`func (o *ClusterSyncReportUpdate) UnsetMessage()`

UnsetMessage ensures that no value is present for Message, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


