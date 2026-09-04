# RoleAllOfAssignments

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AssignedTo** | **string** |  | 
**Type** | [**SubjectType**](SubjectType.md) |  | 
**ScopeId** | **string** |  | 
**AssignedBy** | **string** |  | 
**LastUpdated** | **time.Time** |  | 

## Methods

### NewRoleAllOfAssignments

`func NewRoleAllOfAssignments(assignedTo string, type_ SubjectType, scopeId string, assignedBy string, lastUpdated time.Time, ) *RoleAllOfAssignments`

NewRoleAllOfAssignments instantiates a new RoleAllOfAssignments object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleAllOfAssignmentsWithDefaults

`func NewRoleAllOfAssignmentsWithDefaults() *RoleAllOfAssignments`

NewRoleAllOfAssignmentsWithDefaults instantiates a new RoleAllOfAssignments object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssignedTo

`func (o *RoleAllOfAssignments) GetAssignedTo() string`

GetAssignedTo returns the AssignedTo field if non-nil, zero value otherwise.

### GetAssignedToOk

`func (o *RoleAllOfAssignments) GetAssignedToOk() (*string, bool)`

GetAssignedToOk returns a tuple with the AssignedTo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssignedTo

`func (o *RoleAllOfAssignments) SetAssignedTo(v string)`

SetAssignedTo sets AssignedTo field to given value.


### GetType

`func (o *RoleAllOfAssignments) GetType() SubjectType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *RoleAllOfAssignments) GetTypeOk() (*SubjectType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *RoleAllOfAssignments) SetType(v SubjectType)`

SetType sets Type field to given value.


### GetScopeId

`func (o *RoleAllOfAssignments) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *RoleAllOfAssignments) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *RoleAllOfAssignments) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetAssignedBy

`func (o *RoleAllOfAssignments) GetAssignedBy() string`

GetAssignedBy returns the AssignedBy field if non-nil, zero value otherwise.

### GetAssignedByOk

`func (o *RoleAllOfAssignments) GetAssignedByOk() (*string, bool)`

GetAssignedByOk returns a tuple with the AssignedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssignedBy

`func (o *RoleAllOfAssignments) SetAssignedBy(v string)`

SetAssignedBy sets AssignedBy field to given value.


### GetLastUpdated

`func (o *RoleAllOfAssignments) GetLastUpdated() time.Time`

GetLastUpdated returns the LastUpdated field if non-nil, zero value otherwise.

### GetLastUpdatedOk

`func (o *RoleAllOfAssignments) GetLastUpdatedOk() (*time.Time, bool)`

GetLastUpdatedOk returns a tuple with the LastUpdated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUpdated

`func (o *RoleAllOfAssignments) SetLastUpdated(v time.Time)`

SetLastUpdated sets LastUpdated field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


