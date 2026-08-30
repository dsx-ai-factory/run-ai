# GetDatavolumes200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Datavolumes** | [**[]DatavolumeNoSharedScopes**](DatavolumeNoSharedScopes.md) |  | 

## Methods

### NewGetDatavolumes200Response

`func NewGetDatavolumes200Response(datavolumes []DatavolumeNoSharedScopes, ) *GetDatavolumes200Response`

NewGetDatavolumes200Response instantiates a new GetDatavolumes200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDatavolumes200ResponseWithDefaults

`func NewGetDatavolumes200ResponseWithDefaults() *GetDatavolumes200Response`

NewGetDatavolumes200ResponseWithDefaults instantiates a new GetDatavolumes200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetDatavolumes200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetDatavolumes200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetDatavolumes200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetDatavolumes200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetDatavolumes

`func (o *GetDatavolumes200Response) GetDatavolumes() []DatavolumeNoSharedScopes`

GetDatavolumes returns the Datavolumes field if non-nil, zero value otherwise.

### GetDatavolumesOk

`func (o *GetDatavolumes200Response) GetDatavolumesOk() (*[]DatavolumeNoSharedScopes, bool)`

GetDatavolumesOk returns a tuple with the Datavolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumes

`func (o *GetDatavolumes200Response) SetDatavolumes(v []DatavolumeNoSharedScopes)`

SetDatavolumes sets Datavolumes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


