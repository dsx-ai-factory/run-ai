# PodGroupsSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectName** | **string** | The project name we are syncing objects in | 
**PodGroups** | [**[]PodGroup**](PodGroup.md) |  | 
**From** | Pointer to **string** | the first podgroup to sync. if empty, all podgroups before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last podgroup to sync. if empty, all podgroups after the last one sent will be deleted. | [optional] 

## Methods

### NewPodGroupsSync

`func NewPodGroupsSync(projectName string, podGroups []PodGroup, ) *PodGroupsSync`

NewPodGroupsSync instantiates a new PodGroupsSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodGroupsSyncWithDefaults

`func NewPodGroupsSyncWithDefaults() *PodGroupsSync`

NewPodGroupsSyncWithDefaults instantiates a new PodGroupsSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectName

`func (o *PodGroupsSync) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *PodGroupsSync) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *PodGroupsSync) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetPodGroups

`func (o *PodGroupsSync) GetPodGroups() []PodGroup`

GetPodGroups returns the PodGroups field if non-nil, zero value otherwise.

### GetPodGroupsOk

`func (o *PodGroupsSync) GetPodGroupsOk() (*[]PodGroup, bool)`

GetPodGroupsOk returns a tuple with the PodGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroups

`func (o *PodGroupsSync) SetPodGroups(v []PodGroup)`

SetPodGroups sets PodGroups field to given value.


### GetFrom

`func (o *PodGroupsSync) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *PodGroupsSync) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *PodGroupsSync) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *PodGroupsSync) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *PodGroupsSync) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *PodGroupsSync) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *PodGroupsSync) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *PodGroupsSync) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


