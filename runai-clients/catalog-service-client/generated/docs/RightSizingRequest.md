# RightSizingRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | Cluster whose node-pool hardware metadata is evaluated. | 
**Image** | **string** | Untagged public nvcr.io/nim/organization/model image reference. | 
**Tag** | **string** | Exact image tag. NGC does not automatically refresh latest metadata. | 
**GpuType** | Pointer to **NullableString** | Exact inventory GPU type. Omitted or null evaluates all reported types and capacities. | [optional] 

## Methods

### NewRightSizingRequest

`func NewRightSizingRequest(clusterId string, image string, tag string, ) *RightSizingRequest`

NewRightSizingRequest instantiates a new RightSizingRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRightSizingRequestWithDefaults

`func NewRightSizingRequestWithDefaults() *RightSizingRequest`

NewRightSizingRequestWithDefaults instantiates a new RightSizingRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *RightSizingRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *RightSizingRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *RightSizingRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetImage

`func (o *RightSizingRequest) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *RightSizingRequest) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *RightSizingRequest) SetImage(v string)`

SetImage sets Image field to given value.


### GetTag

`func (o *RightSizingRequest) GetTag() string`

GetTag returns the Tag field if non-nil, zero value otherwise.

### GetTagOk

`func (o *RightSizingRequest) GetTagOk() (*string, bool)`

GetTagOk returns a tuple with the Tag field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTag

`func (o *RightSizingRequest) SetTag(v string)`

SetTag sets Tag field to given value.


### GetGpuType

`func (o *RightSizingRequest) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *RightSizingRequest) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *RightSizingRequest) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.

### HasGpuType

`func (o *RightSizingRequest) HasGpuType() bool`

HasGpuType returns a boolean if a field has been set.

### SetGpuTypeNil

`func (o *RightSizingRequest) SetGpuTypeNil(b bool)`

 SetGpuTypeNil sets the value for GpuType to be an explicit nil

### UnsetGpuType
`func (o *RightSizingRequest) UnsetGpuType()`

UnsetGpuType ensures that no value is present for GpuType, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


