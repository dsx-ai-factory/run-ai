# AssetClusterStatusIssue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**Issue** | [**AssetClusterStatusIssueEnum**](AssetClusterStatusIssueEnum.md) |  | 

## Methods

### NewAssetClusterStatusIssue

`func NewAssetClusterStatusIssue(scopeId string, scopeType ScopeType, issue AssetClusterStatusIssueEnum, ) *AssetClusterStatusIssue`

NewAssetClusterStatusIssue instantiates a new AssetClusterStatusIssue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetClusterStatusIssueWithDefaults

`func NewAssetClusterStatusIssueWithDefaults() *AssetClusterStatusIssue`

NewAssetClusterStatusIssueWithDefaults instantiates a new AssetClusterStatusIssue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeId

`func (o *AssetClusterStatusIssue) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AssetClusterStatusIssue) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AssetClusterStatusIssue) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AssetClusterStatusIssue) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AssetClusterStatusIssue) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AssetClusterStatusIssue) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetIssue

`func (o *AssetClusterStatusIssue) GetIssue() AssetClusterStatusIssueEnum`

GetIssue returns the Issue field if non-nil, zero value otherwise.

### GetIssueOk

`func (o *AssetClusterStatusIssue) GetIssueOk() (*AssetClusterStatusIssueEnum, bool)`

GetIssueOk returns a tuple with the Issue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssue

`func (o *AssetClusterStatusIssue) SetIssue(v AssetClusterStatusIssueEnum)`

SetIssue sets Issue field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


