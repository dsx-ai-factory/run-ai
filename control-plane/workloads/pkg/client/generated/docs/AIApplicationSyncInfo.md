# AIApplicationSyncInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UpdateCount** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 
**Application** | Pointer to [**AIApplicationSyncItem**](AIApplicationSyncItem.md) |  | [optional] 

## Methods

### NewAIApplicationSyncInfo

`func NewAIApplicationSyncInfo() *AIApplicationSyncInfo`

NewAIApplicationSyncInfo instantiates a new AIApplicationSyncInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSyncInfoWithDefaults

`func NewAIApplicationSyncInfoWithDefaults() *AIApplicationSyncInfo`

NewAIApplicationSyncInfoWithDefaults instantiates a new AIApplicationSyncInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUpdateCount

`func (o *AIApplicationSyncInfo) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *AIApplicationSyncInfo) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *AIApplicationSyncInfo) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *AIApplicationSyncInfo) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.

### GetApplication

`func (o *AIApplicationSyncInfo) GetApplication() AIApplicationSyncItem`

GetApplication returns the Application field if non-nil, zero value otherwise.

### GetApplicationOk

`func (o *AIApplicationSyncInfo) GetApplicationOk() (*AIApplicationSyncItem, bool)`

GetApplicationOk returns a tuple with the Application field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplication

`func (o *AIApplicationSyncInfo) SetApplication(v AIApplicationSyncItem)`

SetApplication sets Application field to given value.

### HasApplication

`func (o *AIApplicationSyncInfo) HasApplication() bool`

HasApplication returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


