# GpuPlacementMatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolIds** | **[]string** |  |
**Labels** | **map[string]string** |  |

## Methods

### NewGpuPlacementMatch

`func NewGpuPlacementMatch(nodePoolIds []string, labels map[string]string, ) *GpuPlacementMatch`

NewGpuPlacementMatch instantiates a new GpuPlacementMatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGpuPlacementMatchWithDefaults

`func NewGpuPlacementMatchWithDefaults() *GpuPlacementMatch`

NewGpuPlacementMatchWithDefaults instantiates a new GpuPlacementMatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolIds

`func (o *GpuPlacementMatch) GetNodePoolIds() []string`

GetNodePoolIds returns the NodePoolIds field if non-nil, zero value otherwise.

### GetNodePoolIdsOk

`func (o *GpuPlacementMatch) GetNodePoolIdsOk() (*[]string, bool)`

GetNodePoolIdsOk returns a tuple with the NodePoolIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolIds

`func (o *GpuPlacementMatch) SetNodePoolIds(v []string)`

SetNodePoolIds sets NodePoolIds field to given value.


### GetLabels

`func (o *GpuPlacementMatch) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *GpuPlacementMatch) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *GpuPlacementMatch) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
