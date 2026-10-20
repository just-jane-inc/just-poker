# JustPoker.OpenApi.Api.AdminApi

All URIs are relative to *https://game.bahms.org/api/poker*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**AdminGameGameIdStatusPost**](AdminApi.md#admingamegameidstatuspost) | **POST** /admin/game/{game_id}/status | Update Game Status |
| [**AdminGameGameIdTablePost**](AdminApi.md#admingamegameidtablepost) | **POST** /admin/game/{game_id}/table | Change Game Table |
| [**GameGameIdDeckGet**](AdminApi.md#gamegameiddeckget) | **GET** /game/{game_id}/deck | Get the deck (order) |

<a id="admingamegameidstatuspost"></a>
# **AdminGameGameIdStatusPost**
> Object AdminGameGameIdStatusPost (string gameId, AdminGameStatusDTO adminGameStatusDTO)

Update Game Status

Changes the status of an active game, for use by admins


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to update status of |  |
| **adminGameStatusDTO** | [**AdminGameStatusDTO**](AdminGameStatusDTO.md) | the status of the game to set |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="admingamegameidtablepost"></a>
# **AdminGameGameIdTablePost**
> Object AdminGameGameIdTablePost (string gameId, GameTableDTO gameTableDTO)

Change Game Table

Changes the state of an active game


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to update status of |  |
| **gameTableDTO** | [**GameTableDTO**](GameTableDTO.md) | the table struct to apply |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gamegameiddeckget"></a>
# **GameGameIdDeckGet**
> JustResponseMessageArrayGameCardDTO GameGameIdDeckGet (string gameId)

Get the deck (order)

Gets the order of the deck used by the current hand


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the game to get deck for |  |

### Return type

[**JustResponseMessageArrayGameCardDTO**](JustResponseMessageArrayGameCardDTO.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

