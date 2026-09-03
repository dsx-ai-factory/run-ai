# AgentSyncDataDepricated

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | **string** |  | 
**PodGroups** | Pointer to [**[]PodGroup**](PodGroup.md) |  | [optional] 
**Pods** | Pointer to [**[]PodGroup**](PodGroup.md) |  | [optional] 

## Methods

### NewAgentSyncDataDepricated

`func NewAgentSyncDataDepricated(version string, ) *AgentSyncDataDepricated`

NewAgentSyncDataDepricated instantiates a new AgentSyncDataDepricated object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAgentSyncDataDepricatedWithDefaults

`func NewAgentSyncDataDepricatedWithDefaults() *AgentSyncDataDepricated`

NewAgentSyncDataDepricatedWithDefaults instantiates a new AgentSyncDataDepricated object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *AgentSyncDataDepricated) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *AgentSyncDataDepricated) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *AgentSyncDataDepricated) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetPodGroups

`func (o *AgentSyncDataDepricated) GetPodGroups() []PodGroup`

GetPodGroups returns the PodGroups field if non-nil, zero value otherwise.

### GetPodGroupsOk

`func (o *AgentSyncDataDepricated) GetPodGroupsOk() (*[]PodGroup, bool)`

GetPodGroupsOk returns a tuple with the PodGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroups

`func (o *AgentSyncDataDepricated) SetPodGroups(v []PodGroup)`

SetPodGroups sets PodGroups field to given value.

### HasPodGroups

`func (o *AgentSyncDataDepricated) HasPodGroups() bool`

HasPodGroups returns a boolean if a field has been set.

### GetPods

`func (o *AgentSyncDataDepricated) GetPods() []PodGroup`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *AgentSyncDataDepricated) GetPodsOk() (*[]PodGroup, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *AgentSyncDataDepricated) SetPods(v []PodGroup)`

SetPods sets Pods field to given value.

### HasPods

`func (o *AgentSyncDataDepricated) HasPods() bool`

HasPods returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


