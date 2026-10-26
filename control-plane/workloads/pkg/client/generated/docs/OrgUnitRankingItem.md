# OrgUnitRankingItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the project or department | 
**RunningWorkloadsCount** | Pointer to **NullableInt32** | Number of running workloads (populated by the API response) | [optional] 
**PendingWorkloadsCount** | Pointer to **NullableInt32** | Number of pending workloads (populated by the API response) | [optional] 

## Methods

### NewOrgUnitRankingItem

`func NewOrgUnitRankingItem(id string, ) *OrgUnitRankingItem`

NewOrgUnitRankingItem instantiates a new OrgUnitRankingItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgUnitRankingItemWithDefaults

`func NewOrgUnitRankingItemWithDefaults() *OrgUnitRankingItem`

NewOrgUnitRankingItemWithDefaults instantiates a new OrgUnitRankingItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *OrgUnitRankingItem) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *OrgUnitRankingItem) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *OrgUnitRankingItem) SetId(v string)`

SetId sets Id field to given value.


### GetRunningWorkloadsCount

`func (o *OrgUnitRankingItem) GetRunningWorkloadsCount() int32`

GetRunningWorkloadsCount returns the RunningWorkloadsCount field if non-nil, zero value otherwise.

### GetRunningWorkloadsCountOk

`func (o *OrgUnitRankingItem) GetRunningWorkloadsCountOk() (*int32, bool)`

GetRunningWorkloadsCountOk returns a tuple with the RunningWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningWorkloadsCount

`func (o *OrgUnitRankingItem) SetRunningWorkloadsCount(v int32)`

SetRunningWorkloadsCount sets RunningWorkloadsCount field to given value.

### HasRunningWorkloadsCount

`func (o *OrgUnitRankingItem) HasRunningWorkloadsCount() bool`

HasRunningWorkloadsCount returns a boolean if a field has been set.

### SetRunningWorkloadsCountNil

`func (o *OrgUnitRankingItem) SetRunningWorkloadsCountNil(b bool)`

 SetRunningWorkloadsCountNil sets the value for RunningWorkloadsCount to be an explicit nil

### UnsetRunningWorkloadsCount
`func (o *OrgUnitRankingItem) UnsetRunningWorkloadsCount()`

UnsetRunningWorkloadsCount ensures that no value is present for RunningWorkloadsCount, not even an explicit nil
### GetPendingWorkloadsCount

`func (o *OrgUnitRankingItem) GetPendingWorkloadsCount() int32`

GetPendingWorkloadsCount returns the PendingWorkloadsCount field if non-nil, zero value otherwise.

### GetPendingWorkloadsCountOk

`func (o *OrgUnitRankingItem) GetPendingWorkloadsCountOk() (*int32, bool)`

GetPendingWorkloadsCountOk returns a tuple with the PendingWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPendingWorkloadsCount

`func (o *OrgUnitRankingItem) SetPendingWorkloadsCount(v int32)`

SetPendingWorkloadsCount sets PendingWorkloadsCount field to given value.

### HasPendingWorkloadsCount

`func (o *OrgUnitRankingItem) HasPendingWorkloadsCount() bool`

HasPendingWorkloadsCount returns a boolean if a field has been set.

### SetPendingWorkloadsCountNil

`func (o *OrgUnitRankingItem) SetPendingWorkloadsCountNil(b bool)`

 SetPendingWorkloadsCountNil sets the value for PendingWorkloadsCount to be an explicit nil

### UnsetPendingWorkloadsCount
`func (o *OrgUnitRankingItem) UnsetPendingWorkloadsCount()`

UnsetPendingWorkloadsCount ensures that no value is present for PendingWorkloadsCount, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


