# ClusterInfoDependencies

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Required** | [**map[string]DependencyStatus**](DependencyStatus.md) |  | 
**Optional** | [**map[string]DependencyStatus**](DependencyStatus.md) |  | 

## Methods

### NewClusterInfoDependencies

`func NewClusterInfoDependencies(required map[string]DependencyStatus, optional map[string]DependencyStatus, ) *ClusterInfoDependencies`

NewClusterInfoDependencies instantiates a new ClusterInfoDependencies object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInfoDependenciesWithDefaults

`func NewClusterInfoDependenciesWithDefaults() *ClusterInfoDependencies`

NewClusterInfoDependenciesWithDefaults instantiates a new ClusterInfoDependencies object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequired

`func (o *ClusterInfoDependencies) GetRequired() map[string]DependencyStatus`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *ClusterInfoDependencies) GetRequiredOk() (*map[string]DependencyStatus, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *ClusterInfoDependencies) SetRequired(v map[string]DependencyStatus)`

SetRequired sets Required field to given value.


### GetOptional

`func (o *ClusterInfoDependencies) GetOptional() map[string]DependencyStatus`

GetOptional returns the Optional field if non-nil, zero value otherwise.

### GetOptionalOk

`func (o *ClusterInfoDependencies) GetOptionalOk() (*map[string]DependencyStatus, bool)`

GetOptionalOk returns a tuple with the Optional field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptional

`func (o *ClusterInfoDependencies) SetOptional(v map[string]DependencyStatus)`

SetOptional sets Optional field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


