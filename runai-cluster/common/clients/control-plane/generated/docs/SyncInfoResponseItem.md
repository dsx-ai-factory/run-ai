# SyncInfoResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectName** | Pointer to **string** | The name of the project in which the assets reside. If empty - tenant scoped assets. | [optional] 
**SecondsFromLatestChange** | **int32** |  | 

## Methods

### NewSyncInfoResponseItem

`func NewSyncInfoResponseItem(secondsFromLatestChange int32, ) *SyncInfoResponseItem`

NewSyncInfoResponseItem instantiates a new SyncInfoResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncInfoResponseItemWithDefaults

`func NewSyncInfoResponseItemWithDefaults() *SyncInfoResponseItem`

NewSyncInfoResponseItemWithDefaults instantiates a new SyncInfoResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectName

`func (o *SyncInfoResponseItem) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *SyncInfoResponseItem) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *SyncInfoResponseItem) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *SyncInfoResponseItem) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetSecondsFromLatestChange

`func (o *SyncInfoResponseItem) GetSecondsFromLatestChange() int32`

GetSecondsFromLatestChange returns the SecondsFromLatestChange field if non-nil, zero value otherwise.

### GetSecondsFromLatestChangeOk

`func (o *SyncInfoResponseItem) GetSecondsFromLatestChangeOk() (*int32, bool)`

GetSecondsFromLatestChangeOk returns a tuple with the SecondsFromLatestChange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecondsFromLatestChange

`func (o *SyncInfoResponseItem) SetSecondsFromLatestChange(v int32)`

SetSecondsFromLatestChange sets SecondsFromLatestChange field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


