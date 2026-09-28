# AssetSubmissionIds

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Environment** | Pointer to **NullableString** | Optional unique identifier of the asset. | [optional] 
**Compute** | Pointer to **NullableString** | Optional unique identifier of the asset. | [optional] 
**Datasources** | Pointer to **[]string** |  | [optional] 

## Methods

### NewAssetSubmissionIds

`func NewAssetSubmissionIds() *AssetSubmissionIds`

NewAssetSubmissionIds instantiates a new AssetSubmissionIds object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetSubmissionIdsWithDefaults

`func NewAssetSubmissionIdsWithDefaults() *AssetSubmissionIds`

NewAssetSubmissionIdsWithDefaults instantiates a new AssetSubmissionIds object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnvironment

`func (o *AssetSubmissionIds) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *AssetSubmissionIds) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *AssetSubmissionIds) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.

### HasEnvironment

`func (o *AssetSubmissionIds) HasEnvironment() bool`

HasEnvironment returns a boolean if a field has been set.

### SetEnvironmentNil

`func (o *AssetSubmissionIds) SetEnvironmentNil(b bool)`

 SetEnvironmentNil sets the value for Environment to be an explicit nil

### UnsetEnvironment
`func (o *AssetSubmissionIds) UnsetEnvironment()`

UnsetEnvironment ensures that no value is present for Environment, not even an explicit nil
### GetCompute

`func (o *AssetSubmissionIds) GetCompute() string`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *AssetSubmissionIds) GetComputeOk() (*string, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *AssetSubmissionIds) SetCompute(v string)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *AssetSubmissionIds) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *AssetSubmissionIds) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *AssetSubmissionIds) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetDatasources

`func (o *AssetSubmissionIds) GetDatasources() []string`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *AssetSubmissionIds) GetDatasourcesOk() (*[]string, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *AssetSubmissionIds) SetDatasources(v []string)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *AssetSubmissionIds) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


