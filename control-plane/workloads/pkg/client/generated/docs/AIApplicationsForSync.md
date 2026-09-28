# AIApplicationsForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullSync** | Pointer to **bool** | Upsert all items and delete all missing items | [optional] 
**Applications** | Pointer to [**[]AIApplicationForSync**](AIApplicationForSync.md) |  | [optional] 

## Methods

### NewAIApplicationsForSync

`func NewAIApplicationsForSync() *AIApplicationsForSync`

NewAIApplicationsForSync instantiates a new AIApplicationsForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationsForSyncWithDefaults

`func NewAIApplicationsForSyncWithDefaults() *AIApplicationsForSync`

NewAIApplicationsForSyncWithDefaults instantiates a new AIApplicationsForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullSync

`func (o *AIApplicationsForSync) GetFullSync() bool`

GetFullSync returns the FullSync field if non-nil, zero value otherwise.

### GetFullSyncOk

`func (o *AIApplicationsForSync) GetFullSyncOk() (*bool, bool)`

GetFullSyncOk returns a tuple with the FullSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullSync

`func (o *AIApplicationsForSync) SetFullSync(v bool)`

SetFullSync sets FullSync field to given value.

### HasFullSync

`func (o *AIApplicationsForSync) HasFullSync() bool`

HasFullSync returns a boolean if a field has been set.

### GetApplications

`func (o *AIApplicationsForSync) GetApplications() []AIApplicationForSync`

GetApplications returns the Applications field if non-nil, zero value otherwise.

### GetApplicationsOk

`func (o *AIApplicationsForSync) GetApplicationsOk() (*[]AIApplicationForSync, bool)`

GetApplicationsOk returns a tuple with the Applications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplications

`func (o *AIApplicationsForSync) SetApplications(v []AIApplicationForSync)`

SetApplications sets Applications field to given value.

### HasApplications

`func (o *AIApplicationsForSync) HasApplications() bool`

HasApplications returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


