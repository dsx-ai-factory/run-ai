# DatasourceRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Name** | **string** | The name of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 

## Methods

### NewDatasourceRef

`func NewDatasourceRef(id string, name string, kind AssetKind, ) *DatasourceRef`

NewDatasourceRef instantiates a new DatasourceRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceRefWithDefaults

`func NewDatasourceRefWithDefaults() *DatasourceRef`

NewDatasourceRefWithDefaults instantiates a new DatasourceRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DatasourceRef) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DatasourceRef) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DatasourceRef) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *DatasourceRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DatasourceRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DatasourceRef) SetName(v string)`

SetName sets Name field to given value.


### GetKind

`func (o *DatasourceRef) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *DatasourceRef) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *DatasourceRef) SetKind(v AssetKind)`

SetKind sets Kind field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


