# SyncConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**AvoidDeletionsOfObsoletePolicies** | Pointer to **bool** | true if a customer requested to avoid deletions of obsolete policies on the cluster. The implication is that the customer is responsible for any inconsistency b/w control plane and cluster policies, that the system attempts to prevent by deleting inconsistent cluster policies. | [optional] 

## Methods

### NewSyncConfig

`func NewSyncConfig() *SyncConfig`

NewSyncConfig instantiates a new SyncConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncConfigWithDefaults

`func NewSyncConfigWithDefaults() *SyncConfig`

NewSyncConfigWithDefaults instantiates a new SyncConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SyncConfig) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SyncConfig) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SyncConfig) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *SyncConfig) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetAvoidDeletionsOfObsoletePolicies

`func (o *SyncConfig) GetAvoidDeletionsOfObsoletePolicies() bool`

GetAvoidDeletionsOfObsoletePolicies returns the AvoidDeletionsOfObsoletePolicies field if non-nil, zero value otherwise.

### GetAvoidDeletionsOfObsoletePoliciesOk

`func (o *SyncConfig) GetAvoidDeletionsOfObsoletePoliciesOk() (*bool, bool)`

GetAvoidDeletionsOfObsoletePoliciesOk returns a tuple with the AvoidDeletionsOfObsoletePolicies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvoidDeletionsOfObsoletePolicies

`func (o *SyncConfig) SetAvoidDeletionsOfObsoletePolicies(v bool)`

SetAvoidDeletionsOfObsoletePolicies sets AvoidDeletionsOfObsoletePolicies field to given value.

### HasAvoidDeletionsOfObsoletePolicies

`func (o *SyncConfig) HasAvoidDeletionsOfObsoletePolicies() bool`

HasAvoidDeletionsOfObsoletePolicies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


