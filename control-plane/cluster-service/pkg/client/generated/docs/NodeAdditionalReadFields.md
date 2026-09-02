# NodeAdditionalReadFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The unique identifier of the node. | [optional] 
**ClusterUuid** | **string** | The id of the cluster. | 
**UpdatedAt** | **time.Time** |  | 

## Methods

### NewNodeAdditionalReadFields

`func NewNodeAdditionalReadFields(clusterUuid string, updatedAt time.Time, ) *NodeAdditionalReadFields`

NewNodeAdditionalReadFields instantiates a new NodeAdditionalReadFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeAdditionalReadFieldsWithDefaults

`func NewNodeAdditionalReadFieldsWithDefaults() *NodeAdditionalReadFields`

NewNodeAdditionalReadFieldsWithDefaults instantiates a new NodeAdditionalReadFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodeAdditionalReadFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodeAdditionalReadFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodeAdditionalReadFields) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *NodeAdditionalReadFields) HasId() bool`

HasId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *NodeAdditionalReadFields) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *NodeAdditionalReadFields) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *NodeAdditionalReadFields) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetUpdatedAt

`func (o *NodeAdditionalReadFields) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *NodeAdditionalReadFields) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *NodeAdditionalReadFields) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


