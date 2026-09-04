# ClusterObjectIssue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**Issue** | [**ClusterObjectIssueType**](ClusterObjectIssueType.md) |  | 

## Methods

### NewClusterObjectIssue

`func NewClusterObjectIssue(scopeType string, scopeId string, issue ClusterObjectIssueType, ) *ClusterObjectIssue`

NewClusterObjectIssue instantiates a new ClusterObjectIssue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterObjectIssueWithDefaults

`func NewClusterObjectIssueWithDefaults() *ClusterObjectIssue`

NewClusterObjectIssueWithDefaults instantiates a new ClusterObjectIssue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeType

`func (o *ClusterObjectIssue) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *ClusterObjectIssue) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *ClusterObjectIssue) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *ClusterObjectIssue) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *ClusterObjectIssue) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *ClusterObjectIssue) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetIssue

`func (o *ClusterObjectIssue) GetIssue() ClusterObjectIssueType`

GetIssue returns the Issue field if non-nil, zero value otherwise.

### GetIssueOk

`func (o *ClusterObjectIssue) GetIssueOk() (*ClusterObjectIssueType, bool)`

GetIssueOk returns a tuple with the Issue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssue

`func (o *ClusterObjectIssue) SetIssue(v ClusterObjectIssueType)`

SetIssue sets Issue field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


