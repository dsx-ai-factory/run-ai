# Cluster1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**ClusterMetadata**](ClusterMetadata.md) |  | 
**Current** | [**ClusterCurrent**](ClusterCurrent.md) |  | 
**TimeRange** | Pointer to [**TimeRange**](TimeRange.md) |  | [optional] 

## Methods

### NewCluster1

`func NewCluster1(metadata ClusterMetadata, current ClusterCurrent, ) *Cluster1`

NewCluster1 instantiates a new Cluster1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCluster1WithDefaults

`func NewCluster1WithDefaults() *Cluster1`

NewCluster1WithDefaults instantiates a new Cluster1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *Cluster1) GetMetadata() ClusterMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Cluster1) GetMetadataOk() (*ClusterMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Cluster1) SetMetadata(v ClusterMetadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *Cluster1) GetCurrent() ClusterCurrent`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *Cluster1) GetCurrentOk() (*ClusterCurrent, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *Cluster1) SetCurrent(v ClusterCurrent)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *Cluster1) GetTimeRange() TimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Cluster1) GetTimeRangeOk() (*TimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Cluster1) SetTimeRange(v TimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Cluster1) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


