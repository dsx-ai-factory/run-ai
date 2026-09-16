# Workloadv2Authorization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthorizationType** | [**NullableAuthorizationType**](AuthorizationType.md) |  | 
**AuthorizedUsers** | Pointer to **[]string** | List of users or service accounts that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** | List of groups that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**ForceJsonResponse** | Pointer to **NullableBool** | Force JSON errors instead of HTTP error pages or redirects. | [optional] [default to false]

## Methods

### NewWorkloadv2Authorization

`func NewWorkloadv2Authorization(authorizationType NullableAuthorizationType, ) *Workloadv2Authorization`

NewWorkloadv2Authorization instantiates a new Workloadv2Authorization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadv2AuthorizationWithDefaults

`func NewWorkloadv2AuthorizationWithDefaults() *Workloadv2Authorization`

NewWorkloadv2AuthorizationWithDefaults instantiates a new Workloadv2Authorization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthorizationType

`func (o *Workloadv2Authorization) GetAuthorizationType() AuthorizationType`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *Workloadv2Authorization) GetAuthorizationTypeOk() (*AuthorizationType, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *Workloadv2Authorization) SetAuthorizationType(v AuthorizationType)`

SetAuthorizationType sets AuthorizationType field to given value.


### SetAuthorizationTypeNil

`func (o *Workloadv2Authorization) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *Workloadv2Authorization) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *Workloadv2Authorization) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *Workloadv2Authorization) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *Workloadv2Authorization) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *Workloadv2Authorization) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *Workloadv2Authorization) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *Workloadv2Authorization) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *Workloadv2Authorization) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *Workloadv2Authorization) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *Workloadv2Authorization) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *Workloadv2Authorization) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *Workloadv2Authorization) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *Workloadv2Authorization) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetForceJsonResponse

`func (o *Workloadv2Authorization) GetForceJsonResponse() bool`

GetForceJsonResponse returns the ForceJsonResponse field if non-nil, zero value otherwise.

### GetForceJsonResponseOk

`func (o *Workloadv2Authorization) GetForceJsonResponseOk() (*bool, bool)`

GetForceJsonResponseOk returns a tuple with the ForceJsonResponse field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForceJsonResponse

`func (o *Workloadv2Authorization) SetForceJsonResponse(v bool)`

SetForceJsonResponse sets ForceJsonResponse field to given value.

### HasForceJsonResponse

`func (o *Workloadv2Authorization) HasForceJsonResponse() bool`

HasForceJsonResponse returns a boolean if a field has been set.

### SetForceJsonResponseNil

`func (o *Workloadv2Authorization) SetForceJsonResponseNil(b bool)`

 SetForceJsonResponseNil sets the value for ForceJsonResponse to be an explicit nil

### UnsetForceJsonResponse
`func (o *Workloadv2Authorization) UnsetForceJsonResponse()`

UnsetForceJsonResponse ensures that no value is present for ForceJsonResponse, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


