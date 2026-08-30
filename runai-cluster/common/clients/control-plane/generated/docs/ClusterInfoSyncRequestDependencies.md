# ClusterInfoSyncRequestDependencies

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Required** | [**map[string]ClusterDependencyStatus**](ClusterDependencyStatus.md) |  | 
**Optional** | [**map[string]ClusterDependencyStatus**](ClusterDependencyStatus.md) |  | 

## Methods

### NewClusterInfoSyncRequestDependencies

`func NewClusterInfoSyncRequestDependencies(required map[string]ClusterDependencyStatus, optional map[string]ClusterDependencyStatus, ) *ClusterInfoSyncRequestDependencies`

NewClusterInfoSyncRequestDependencies instantiates a new ClusterInfoSyncRequestDependencies object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInfoSyncRequestDependenciesWithDefaults

`func NewClusterInfoSyncRequestDependenciesWithDefaults() *ClusterInfoSyncRequestDependencies`

NewClusterInfoSyncRequestDependenciesWithDefaults instantiates a new ClusterInfoSyncRequestDependencies object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequired

`func (o *ClusterInfoSyncRequestDependencies) GetRequired() map[string]ClusterDependencyStatus`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *ClusterInfoSyncRequestDependencies) GetRequiredOk() (*map[string]ClusterDependencyStatus, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *ClusterInfoSyncRequestDependencies) SetRequired(v map[string]ClusterDependencyStatus)`

SetRequired sets Required field to given value.


### GetOptional

`func (o *ClusterInfoSyncRequestDependencies) GetOptional() map[string]ClusterDependencyStatus`

GetOptional returns the Optional field if non-nil, zero value otherwise.

### GetOptionalOk

`func (o *ClusterInfoSyncRequestDependencies) GetOptionalOk() (*map[string]ClusterDependencyStatus, bool)`

GetOptionalOk returns a tuple with the Optional field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptional

`func (o *ClusterInfoSyncRequestDependencies) SetOptional(v map[string]ClusterDependencyStatus)`

SetOptional sets Optional field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


