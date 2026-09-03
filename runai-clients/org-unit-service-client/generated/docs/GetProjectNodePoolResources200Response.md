# GetProjectNodePoolResources200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolResources** | [**[]ProjectNodePoolResources**](ProjectNodePoolResources.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetProjectNodePoolResources200Response

`func NewGetProjectNodePoolResources200Response(nodePoolResources []ProjectNodePoolResources, ) *GetProjectNodePoolResources200Response`

NewGetProjectNodePoolResources200Response instantiates a new GetProjectNodePoolResources200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetProjectNodePoolResources200ResponseWithDefaults

`func NewGetProjectNodePoolResources200ResponseWithDefaults() *GetProjectNodePoolResources200Response`

NewGetProjectNodePoolResources200ResponseWithDefaults instantiates a new GetProjectNodePoolResources200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolResources

`func (o *GetProjectNodePoolResources200Response) GetNodePoolResources() []ProjectNodePoolResources`

GetNodePoolResources returns the NodePoolResources field if non-nil, zero value otherwise.

### GetNodePoolResourcesOk

`func (o *GetProjectNodePoolResources200Response) GetNodePoolResourcesOk() (*[]ProjectNodePoolResources, bool)`

GetNodePoolResourcesOk returns a tuple with the NodePoolResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolResources

`func (o *GetProjectNodePoolResources200Response) SetNodePoolResources(v []ProjectNodePoolResources)`

SetNodePoolResources sets NodePoolResources field to given value.


### GetNext

`func (o *GetProjectNodePoolResources200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetProjectNodePoolResources200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetProjectNodePoolResources200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetProjectNodePoolResources200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


