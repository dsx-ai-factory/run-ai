# CatalogParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OrderBy** | Pointer to [**[]CatalogParamsOrderByInner**](CatalogParamsOrderByInner.md) |  | [optional] 
**QueryFields** | Pointer to **[]string** |  | [optional] 
**ScoredSize** | Pointer to **int32** |  | [optional] 
**PageSize** | Pointer to **int32** |  | [optional] 
**Fields** | Pointer to **[]string** |  | [optional] 
**Page** | Pointer to **int32** |  | [optional] 
**Filters** | Pointer to [**[]CatalogParamsOrderByInner**](CatalogParamsOrderByInner.md) |  | [optional] 
**Query** | Pointer to **string** |  | [optional] 

## Methods

### NewCatalogParams

`func NewCatalogParams() *CatalogParams`

NewCatalogParams instantiates a new CatalogParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCatalogParamsWithDefaults

`func NewCatalogParamsWithDefaults() *CatalogParams`

NewCatalogParamsWithDefaults instantiates a new CatalogParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOrderBy

`func (o *CatalogParams) GetOrderBy() []CatalogParamsOrderByInner`

GetOrderBy returns the OrderBy field if non-nil, zero value otherwise.

### GetOrderByOk

`func (o *CatalogParams) GetOrderByOk() (*[]CatalogParamsOrderByInner, bool)`

GetOrderByOk returns a tuple with the OrderBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrderBy

`func (o *CatalogParams) SetOrderBy(v []CatalogParamsOrderByInner)`

SetOrderBy sets OrderBy field to given value.

### HasOrderBy

`func (o *CatalogParams) HasOrderBy() bool`

HasOrderBy returns a boolean if a field has been set.

### GetQueryFields

`func (o *CatalogParams) GetQueryFields() []string`

GetQueryFields returns the QueryFields field if non-nil, zero value otherwise.

### GetQueryFieldsOk

`func (o *CatalogParams) GetQueryFieldsOk() (*[]string, bool)`

GetQueryFieldsOk returns a tuple with the QueryFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueryFields

`func (o *CatalogParams) SetQueryFields(v []string)`

SetQueryFields sets QueryFields field to given value.

### HasQueryFields

`func (o *CatalogParams) HasQueryFields() bool`

HasQueryFields returns a boolean if a field has been set.

### GetScoredSize

`func (o *CatalogParams) GetScoredSize() int32`

GetScoredSize returns the ScoredSize field if non-nil, zero value otherwise.

### GetScoredSizeOk

`func (o *CatalogParams) GetScoredSizeOk() (*int32, bool)`

GetScoredSizeOk returns a tuple with the ScoredSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScoredSize

`func (o *CatalogParams) SetScoredSize(v int32)`

SetScoredSize sets ScoredSize field to given value.

### HasScoredSize

`func (o *CatalogParams) HasScoredSize() bool`

HasScoredSize returns a boolean if a field has been set.

### GetPageSize

`func (o *CatalogParams) GetPageSize() int32`

GetPageSize returns the PageSize field if non-nil, zero value otherwise.

### GetPageSizeOk

`func (o *CatalogParams) GetPageSizeOk() (*int32, bool)`

GetPageSizeOk returns a tuple with the PageSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageSize

`func (o *CatalogParams) SetPageSize(v int32)`

SetPageSize sets PageSize field to given value.

### HasPageSize

`func (o *CatalogParams) HasPageSize() bool`

HasPageSize returns a boolean if a field has been set.

### GetFields

`func (o *CatalogParams) GetFields() []string`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *CatalogParams) GetFieldsOk() (*[]string, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *CatalogParams) SetFields(v []string)`

SetFields sets Fields field to given value.

### HasFields

`func (o *CatalogParams) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetPage

`func (o *CatalogParams) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *CatalogParams) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *CatalogParams) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *CatalogParams) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetFilters

`func (o *CatalogParams) GetFilters() []CatalogParamsOrderByInner`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *CatalogParams) GetFiltersOk() (*[]CatalogParamsOrderByInner, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *CatalogParams) SetFilters(v []CatalogParamsOrderByInner)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *CatalogParams) HasFilters() bool`

HasFilters returns a boolean if a field has been set.

### GetQuery

`func (o *CatalogParams) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *CatalogParams) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *CatalogParams) SetQuery(v string)`

SetQuery sets Query field to given value.

### HasQuery

`func (o *CatalogParams) HasQuery() bool`

HasQuery returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


