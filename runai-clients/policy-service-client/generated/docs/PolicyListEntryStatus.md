# PolicyListEntryStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Clusters** | Pointer to [**[]PolicySyncStatusOfCluster**](PolicySyncStatusOfCluster.md) |  | [optional] 
**Validation** | Pointer to [**NullablePolicyValidationStatusValidation**](PolicyValidationStatusValidation.md) |  | [optional] 

## Methods

### NewPolicyListEntryStatus

`func NewPolicyListEntryStatus() *PolicyListEntryStatus`

NewPolicyListEntryStatus instantiates a new PolicyListEntryStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyListEntryStatusWithDefaults

`func NewPolicyListEntryStatusWithDefaults() *PolicyListEntryStatus`

NewPolicyListEntryStatusWithDefaults instantiates a new PolicyListEntryStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusters

`func (o *PolicyListEntryStatus) GetClusters() []PolicySyncStatusOfCluster`

GetClusters returns the Clusters field if non-nil, zero value otherwise.

### GetClustersOk

`func (o *PolicyListEntryStatus) GetClustersOk() (*[]PolicySyncStatusOfCluster, bool)`

GetClustersOk returns a tuple with the Clusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusters

`func (o *PolicyListEntryStatus) SetClusters(v []PolicySyncStatusOfCluster)`

SetClusters sets Clusters field to given value.

### HasClusters

`func (o *PolicyListEntryStatus) HasClusters() bool`

HasClusters returns a boolean if a field has been set.

### GetValidation

`func (o *PolicyListEntryStatus) GetValidation() PolicyValidationStatusValidation`

GetValidation returns the Validation field if non-nil, zero value otherwise.

### GetValidationOk

`func (o *PolicyListEntryStatus) GetValidationOk() (*PolicyValidationStatusValidation, bool)`

GetValidationOk returns a tuple with the Validation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidation

`func (o *PolicyListEntryStatus) SetValidation(v PolicyValidationStatusValidation)`

SetValidation sets Validation field to given value.

### HasValidation

`func (o *PolicyListEntryStatus) HasValidation() bool`

HasValidation returns a boolean if a field has been set.

### SetValidationNil

`func (o *PolicyListEntryStatus) SetValidationNil(b bool)`

 SetValidationNil sets the value for Validation to be an explicit nil

### UnsetValidation
`func (o *PolicyListEntryStatus) UnsetValidation()`

UnsetValidation ensures that no value is present for Validation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


