# GetDatavolumesInternal200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Datavolumes** | [**[]DatavolumeInternal**](DatavolumeInternal.md) |  | 

## Methods

### NewGetDatavolumesInternal200Response

`func NewGetDatavolumesInternal200Response(datavolumes []DatavolumeInternal, ) *GetDatavolumesInternal200Response`

NewGetDatavolumesInternal200Response instantiates a new GetDatavolumesInternal200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDatavolumesInternal200ResponseWithDefaults

`func NewGetDatavolumesInternal200ResponseWithDefaults() *GetDatavolumesInternal200Response`

NewGetDatavolumesInternal200ResponseWithDefaults instantiates a new GetDatavolumesInternal200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetDatavolumesInternal200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetDatavolumesInternal200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetDatavolumesInternal200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetDatavolumesInternal200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetDatavolumes

`func (o *GetDatavolumesInternal200Response) GetDatavolumes() []DatavolumeInternal`

GetDatavolumes returns the Datavolumes field if non-nil, zero value otherwise.

### GetDatavolumesOk

`func (o *GetDatavolumesInternal200Response) GetDatavolumesOk() (*[]DatavolumeInternal, bool)`

GetDatavolumesOk returns a tuple with the Datavolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumes

`func (o *GetDatavolumesInternal200Response) SetDatavolumes(v []DatavolumeInternal)`

SetDatavolumes sets Datavolumes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


