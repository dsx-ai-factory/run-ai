# ClusterCurrent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | [**[]ClusterCurrentResourcesInner**](ClusterCurrentResourcesInner.md) |  | 
**ProjectResources** | [**[]ClusterCurrentProjectResourcesInner**](ClusterCurrentProjectResourcesInner.md) |  | 

## Methods

### NewClusterCurrent

`func NewClusterCurrent(resources []ClusterCurrentResourcesInner, projectResources []ClusterCurrentProjectResourcesInner, ) *ClusterCurrent`

NewClusterCurrent instantiates a new ClusterCurrent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCurrentWithDefaults

`func NewClusterCurrentWithDefaults() *ClusterCurrent`

NewClusterCurrentWithDefaults instantiates a new ClusterCurrent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *ClusterCurrent) GetResources() []ClusterCurrentResourcesInner`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ClusterCurrent) GetResourcesOk() (*[]ClusterCurrentResourcesInner, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ClusterCurrent) SetResources(v []ClusterCurrentResourcesInner)`

SetResources sets Resources field to given value.


### GetProjectResources

`func (o *ClusterCurrent) GetProjectResources() []ClusterCurrentProjectResourcesInner`

GetProjectResources returns the ProjectResources field if non-nil, zero value otherwise.

### GetProjectResourcesOk

`func (o *ClusterCurrent) GetProjectResourcesOk() (*[]ClusterCurrentProjectResourcesInner, bool)`

GetProjectResourcesOk returns a tuple with the ProjectResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectResources

`func (o *ClusterCurrent) SetProjectResources(v []ClusterCurrentProjectResourcesInner)`

SetProjectResources sets ProjectResources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


