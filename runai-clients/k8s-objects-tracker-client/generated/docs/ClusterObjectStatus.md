# ClusterObjectStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the tenant, cluster, department or project, depending on the scope type. | 
**ResourceType** | [**ClusterResourceType**](ClusterResourceType.md) |  | 
**ResourceName** | **string** |  | 
**SecretType** | Pointer to [**SecretType**](SecretType.md) |  | [optional] 
**Issues** | Pointer to [**[]ClusterObjectIssue**](ClusterObjectIssue.md) |  | [optional] 

## Methods

### NewClusterObjectStatus

`func NewClusterObjectStatus(scopeType string, scopeId string, resourceType ClusterResourceType, resourceName string, ) *ClusterObjectStatus`

NewClusterObjectStatus instantiates a new ClusterObjectStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterObjectStatusWithDefaults

`func NewClusterObjectStatusWithDefaults() *ClusterObjectStatus`

NewClusterObjectStatusWithDefaults instantiates a new ClusterObjectStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeType

`func (o *ClusterObjectStatus) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *ClusterObjectStatus) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *ClusterObjectStatus) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *ClusterObjectStatus) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *ClusterObjectStatus) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *ClusterObjectStatus) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetResourceType

`func (o *ClusterObjectStatus) GetResourceType() ClusterResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *ClusterObjectStatus) GetResourceTypeOk() (*ClusterResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *ClusterObjectStatus) SetResourceType(v ClusterResourceType)`

SetResourceType sets ResourceType field to given value.


### GetResourceName

`func (o *ClusterObjectStatus) GetResourceName() string`

GetResourceName returns the ResourceName field if non-nil, zero value otherwise.

### GetResourceNameOk

`func (o *ClusterObjectStatus) GetResourceNameOk() (*string, bool)`

GetResourceNameOk returns a tuple with the ResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceName

`func (o *ClusterObjectStatus) SetResourceName(v string)`

SetResourceName sets ResourceName field to given value.


### GetSecretType

`func (o *ClusterObjectStatus) GetSecretType() SecretType`

GetSecretType returns the SecretType field if non-nil, zero value otherwise.

### GetSecretTypeOk

`func (o *ClusterObjectStatus) GetSecretTypeOk() (*SecretType, bool)`

GetSecretTypeOk returns a tuple with the SecretType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretType

`func (o *ClusterObjectStatus) SetSecretType(v SecretType)`

SetSecretType sets SecretType field to given value.

### HasSecretType

`func (o *ClusterObjectStatus) HasSecretType() bool`

HasSecretType returns a boolean if a field has been set.

### GetIssues

`func (o *ClusterObjectStatus) GetIssues() []ClusterObjectIssue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *ClusterObjectStatus) GetIssuesOk() (*[]ClusterObjectIssue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *ClusterObjectStatus) SetIssues(v []ClusterObjectIssue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *ClusterObjectStatus) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


