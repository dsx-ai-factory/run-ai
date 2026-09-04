# RevisionsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectId** | **string** |  | 
**ClusterId** | Pointer to **string** |  | [optional] 
**Revisions** | [**[]RevisionSyncFields**](RevisionSyncFields.md) |  | 
**From** | Pointer to **string** | the first revision id to sync. if empty, all revisions before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last revision to sync. if empty, all revisions after the last one sent will be deleted. | [optional] 
**IdsToIgnore** | Pointer to **[]string** | revision ids to ignore when deleting. | [optional] 

## Methods

### NewRevisionsSyncRequest

`func NewRevisionsSyncRequest(projectId string, revisions []RevisionSyncFields, ) *RevisionsSyncRequest`

NewRevisionsSyncRequest instantiates a new RevisionsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRevisionsSyncRequestWithDefaults

`func NewRevisionsSyncRequestWithDefaults() *RevisionsSyncRequest`

NewRevisionsSyncRequestWithDefaults instantiates a new RevisionsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectId

`func (o *RevisionsSyncRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *RevisionsSyncRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *RevisionsSyncRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *RevisionsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *RevisionsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *RevisionsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *RevisionsSyncRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRevisions

`func (o *RevisionsSyncRequest) GetRevisions() []RevisionSyncFields`

GetRevisions returns the Revisions field if non-nil, zero value otherwise.

### GetRevisionsOk

`func (o *RevisionsSyncRequest) GetRevisionsOk() (*[]RevisionSyncFields, bool)`

GetRevisionsOk returns a tuple with the Revisions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevisions

`func (o *RevisionsSyncRequest) SetRevisions(v []RevisionSyncFields)`

SetRevisions sets Revisions field to given value.


### GetFrom

`func (o *RevisionsSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *RevisionsSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *RevisionsSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *RevisionsSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *RevisionsSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *RevisionsSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *RevisionsSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *RevisionsSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.

### GetIdsToIgnore

`func (o *RevisionsSyncRequest) GetIdsToIgnore() []string`

GetIdsToIgnore returns the IdsToIgnore field if non-nil, zero value otherwise.

### GetIdsToIgnoreOk

`func (o *RevisionsSyncRequest) GetIdsToIgnoreOk() (*[]string, bool)`

GetIdsToIgnoreOk returns a tuple with the IdsToIgnore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdsToIgnore

`func (o *RevisionsSyncRequest) SetIdsToIgnore(v []string)`

SetIdsToIgnore sets IdsToIgnore field to given value.

### HasIdsToIgnore

`func (o *RevisionsSyncRequest) HasIdsToIgnore() bool`

HasIdsToIgnore returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


