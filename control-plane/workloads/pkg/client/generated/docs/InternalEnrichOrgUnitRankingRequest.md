# InternalEnrichOrgUnitRankingRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The cluster ID | 
**OrgUnitType** | **string** | Whether these are projects or departments (all IDs must be same type) | 
**NodePoolName** | Pointer to **NullableString** | Optional node pool filter for scoped counting | [optional] 
**Ids** | **[]string** | Array of project or department IDs to enrich | 

## Methods

### NewInternalEnrichOrgUnitRankingRequest

`func NewInternalEnrichOrgUnitRankingRequest(clusterId string, orgUnitType string, ids []string, ) *InternalEnrichOrgUnitRankingRequest`

NewInternalEnrichOrgUnitRankingRequest instantiates a new InternalEnrichOrgUnitRankingRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalEnrichOrgUnitRankingRequestWithDefaults

`func NewInternalEnrichOrgUnitRankingRequestWithDefaults() *InternalEnrichOrgUnitRankingRequest`

NewInternalEnrichOrgUnitRankingRequestWithDefaults instantiates a new InternalEnrichOrgUnitRankingRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *InternalEnrichOrgUnitRankingRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InternalEnrichOrgUnitRankingRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InternalEnrichOrgUnitRankingRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOrgUnitType

`func (o *InternalEnrichOrgUnitRankingRequest) GetOrgUnitType() string`

GetOrgUnitType returns the OrgUnitType field if non-nil, zero value otherwise.

### GetOrgUnitTypeOk

`func (o *InternalEnrichOrgUnitRankingRequest) GetOrgUnitTypeOk() (*string, bool)`

GetOrgUnitTypeOk returns a tuple with the OrgUnitType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitType

`func (o *InternalEnrichOrgUnitRankingRequest) SetOrgUnitType(v string)`

SetOrgUnitType sets OrgUnitType field to given value.


### GetNodePoolName

`func (o *InternalEnrichOrgUnitRankingRequest) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *InternalEnrichOrgUnitRankingRequest) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *InternalEnrichOrgUnitRankingRequest) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.

### HasNodePoolName

`func (o *InternalEnrichOrgUnitRankingRequest) HasNodePoolName() bool`

HasNodePoolName returns a boolean if a field has been set.

### SetNodePoolNameNil

`func (o *InternalEnrichOrgUnitRankingRequest) SetNodePoolNameNil(b bool)`

 SetNodePoolNameNil sets the value for NodePoolName to be an explicit nil

### UnsetNodePoolName
`func (o *InternalEnrichOrgUnitRankingRequest) UnsetNodePoolName()`

UnsetNodePoolName ensures that no value is present for NodePoolName, not even an explicit nil
### GetIds

`func (o *InternalEnrichOrgUnitRankingRequest) GetIds() []string`

GetIds returns the Ids field if non-nil, zero value otherwise.

### GetIdsOk

`func (o *InternalEnrichOrgUnitRankingRequest) GetIdsOk() (*[]string, bool)`

GetIdsOk returns a tuple with the Ids field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIds

`func (o *InternalEnrichOrgUnitRankingRequest) SetIds(v []string)`

SetIds sets Ids field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


