# TrialState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TrialId** | **NullableString** |  | 
**TenantName** | **NullableString** |  | 
**TenantDomain** | **NullableString** |  | 
**Email** | **NullableString** |  | 
**Company** | **NullableString** |  | 
**TenantId** | **NullableString** |  | 
**ClusterId** | **NullableString** | The id of the cluster. | 
**ClusterStatus** | **NullableString** |  | 
**FirstLogin** | **NullableTime** |  | 

## Methods

### NewTrialState

`func NewTrialState(trialId NullableString, tenantName NullableString, tenantDomain NullableString, email NullableString, company NullableString, tenantId NullableString, clusterId NullableString, clusterStatus NullableString, firstLogin NullableTime, ) *TrialState`

NewTrialState instantiates a new TrialState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrialStateWithDefaults

`func NewTrialStateWithDefaults() *TrialState`

NewTrialStateWithDefaults instantiates a new TrialState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTrialId

`func (o *TrialState) GetTrialId() string`

GetTrialId returns the TrialId field if non-nil, zero value otherwise.

### GetTrialIdOk

`func (o *TrialState) GetTrialIdOk() (*string, bool)`

GetTrialIdOk returns a tuple with the TrialId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrialId

`func (o *TrialState) SetTrialId(v string)`

SetTrialId sets TrialId field to given value.


### SetTrialIdNil

`func (o *TrialState) SetTrialIdNil(b bool)`

 SetTrialIdNil sets the value for TrialId to be an explicit nil

### UnsetTrialId
`func (o *TrialState) UnsetTrialId()`

UnsetTrialId ensures that no value is present for TrialId, not even an explicit nil
### GetTenantName

`func (o *TrialState) GetTenantName() string`

GetTenantName returns the TenantName field if non-nil, zero value otherwise.

### GetTenantNameOk

`func (o *TrialState) GetTenantNameOk() (*string, bool)`

GetTenantNameOk returns a tuple with the TenantName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantName

`func (o *TrialState) SetTenantName(v string)`

SetTenantName sets TenantName field to given value.


### SetTenantNameNil

`func (o *TrialState) SetTenantNameNil(b bool)`

 SetTenantNameNil sets the value for TenantName to be an explicit nil

### UnsetTenantName
`func (o *TrialState) UnsetTenantName()`

UnsetTenantName ensures that no value is present for TenantName, not even an explicit nil
### GetTenantDomain

`func (o *TrialState) GetTenantDomain() string`

GetTenantDomain returns the TenantDomain field if non-nil, zero value otherwise.

### GetTenantDomainOk

`func (o *TrialState) GetTenantDomainOk() (*string, bool)`

GetTenantDomainOk returns a tuple with the TenantDomain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantDomain

`func (o *TrialState) SetTenantDomain(v string)`

SetTenantDomain sets TenantDomain field to given value.


### SetTenantDomainNil

`func (o *TrialState) SetTenantDomainNil(b bool)`

 SetTenantDomainNil sets the value for TenantDomain to be an explicit nil

### UnsetTenantDomain
`func (o *TrialState) UnsetTenantDomain()`

UnsetTenantDomain ensures that no value is present for TenantDomain, not even an explicit nil
### GetEmail

`func (o *TrialState) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *TrialState) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *TrialState) SetEmail(v string)`

SetEmail sets Email field to given value.


### SetEmailNil

`func (o *TrialState) SetEmailNil(b bool)`

 SetEmailNil sets the value for Email to be an explicit nil

### UnsetEmail
`func (o *TrialState) UnsetEmail()`

UnsetEmail ensures that no value is present for Email, not even an explicit nil
### GetCompany

`func (o *TrialState) GetCompany() string`

GetCompany returns the Company field if non-nil, zero value otherwise.

### GetCompanyOk

`func (o *TrialState) GetCompanyOk() (*string, bool)`

GetCompanyOk returns a tuple with the Company field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompany

`func (o *TrialState) SetCompany(v string)`

SetCompany sets Company field to given value.


### SetCompanyNil

`func (o *TrialState) SetCompanyNil(b bool)`

 SetCompanyNil sets the value for Company to be an explicit nil

### UnsetCompany
`func (o *TrialState) UnsetCompany()`

UnsetCompany ensures that no value is present for Company, not even an explicit nil
### GetTenantId

`func (o *TrialState) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *TrialState) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *TrialState) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.


### SetTenantIdNil

`func (o *TrialState) SetTenantIdNil(b bool)`

 SetTenantIdNil sets the value for TenantId to be an explicit nil

### UnsetTenantId
`func (o *TrialState) UnsetTenantId()`

UnsetTenantId ensures that no value is present for TenantId, not even an explicit nil
### GetClusterId

`func (o *TrialState) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *TrialState) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *TrialState) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### SetClusterIdNil

`func (o *TrialState) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *TrialState) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetClusterStatus

`func (o *TrialState) GetClusterStatus() string`

GetClusterStatus returns the ClusterStatus field if non-nil, zero value otherwise.

### GetClusterStatusOk

`func (o *TrialState) GetClusterStatusOk() (*string, bool)`

GetClusterStatusOk returns a tuple with the ClusterStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterStatus

`func (o *TrialState) SetClusterStatus(v string)`

SetClusterStatus sets ClusterStatus field to given value.


### SetClusterStatusNil

`func (o *TrialState) SetClusterStatusNil(b bool)`

 SetClusterStatusNil sets the value for ClusterStatus to be an explicit nil

### UnsetClusterStatus
`func (o *TrialState) UnsetClusterStatus()`

UnsetClusterStatus ensures that no value is present for ClusterStatus, not even an explicit nil
### GetFirstLogin

`func (o *TrialState) GetFirstLogin() time.Time`

GetFirstLogin returns the FirstLogin field if non-nil, zero value otherwise.

### GetFirstLoginOk

`func (o *TrialState) GetFirstLoginOk() (*time.Time, bool)`

GetFirstLoginOk returns a tuple with the FirstLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFirstLogin

`func (o *TrialState) SetFirstLogin(v time.Time)`

SetFirstLogin sets FirstLogin field to given value.


### SetFirstLoginNil

`func (o *TrialState) SetFirstLoginNil(b bool)`

 SetFirstLoginNil sets the value for FirstLogin to be an explicit nil

### UnsetFirstLogin
`func (o *TrialState) UnsetFirstLogin()`

UnsetFirstLogin ensures that no value is present for FirstLogin, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


