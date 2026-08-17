# JustPoker.OpenApi.Api.GameApi

All URIs are relative to *https://game.bahms.org/api/poker*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GameGameIdActionPost**](GameApi.md#gamegameidactionpost) | **POST** /game/{game_id}/action | Player Action |
| [**GameGameIdChipExchangePost**](GameApi.md#gamegameidchipexchangepost) | **POST** /game/{game_id}/chip/exchange | Exchange Chips |
| [**GameGameIdDelete**](GameApi.md#gamegameiddelete) | **DELETE** /game/{game_id} | Delete a Game |
| [**GameGameIdHandPost**](GameApi.md#gamegameidhandpost) | **POST** /game/{game_id}/hand | Start Next Hand |
| [**GameGameIdPlayerPost**](GameApi.md#gamegameidplayerpost) | **POST** /game/{game_id}/player | Join a Game |
| [**GameGameIdStartedPost**](GameApi.md#gamegameidstartedpost) | **POST** /game/{game_id}/started | Start Game |
| [**GameGameIdStateGet**](GameApi.md#gamegameidstateget) | **GET** /game/{game_id}/state | Game State |
| [**GameGameIdStateListenGet**](GameApi.md#gamegameidstatelistenget) | **GET** /game/{game_id}/state/listen | Get Listener |
| [**GameGameIdStateListenPost**](GameApi.md#gamegameidstatelistenpost) | **POST** /game/{game_id}/state/listen | Register Listener |
| [**GameGameIdStatePost**](GameApi.md#gamegameidstatepost) | **POST** /game/{game_id}/state | start game from state |
| [**GameGameIdStateWsGet**](GameApi.md#gamegameidstatewsget) | **GET** /game/{game_id}/state/ws | Connect Updates |
| [**GameGet**](GameApi.md#gameget) | **GET** /game | Gets Active Games |
| [**GamePost**](GameApi.md#gamepost) | **POST** /game | Create Game |
| [**HandEvaluatorEvaluatePost**](GameApi.md#handevaluatorevaluatepost) | **POST** /hand-evaluator/evaluate/ | Evaluate a Hand |

<a id="gamegameidactionpost"></a>
# **GameGameIdActionPost**
> JustResponseMessageAny GameGameIdActionPost (string gameId, GamePlayerActionDTO gamePlayerActionDTO)

Player Action

post the action preformed by a player


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | the id of the game |  |
| **gamePlayerActionDTO** | [**GamePlayerActionDTO**](GamePlayerActionDTO.md) | the action the player is preforming |  |

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

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

<a id="gamegameidchipexchangepost"></a>
# **GameGameIdChipExchangePost**
> JustResponseMessageAny GameGameIdChipExchangePost (string gameId, GameChipExchangeDTO gameChipExchangeDTO)

Exchange Chips

exchange chips in the players stack with the tables rack


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game exchange chips in |  |
| **gameChipExchangeDTO** | [**GameChipExchangeDTO**](GameChipExchangeDTO.md) | a specification for the chips to exchange |  |

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

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

<a id="gamegameiddelete"></a>
# **GameGameIdDelete**
> JustResponseMessageAny GameGameIdDelete (string gameId, Object body = null)

Delete a Game

Delete a game


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to delete |  |
| **body** | **Object** |  | [optional]  |

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

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

<a id="gamegameidhandpost"></a>
# **GameGameIdHandPost**
> Object GameGameIdHandPost (string gameId, GameNewHandDTO gameNewHandDTO)

Start Next Hand

Starts the next poker hand


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to join |  |
| **gameNewHandDTO** | [**GameNewHandDTO**](GameNewHandDTO.md) | a dto containing new hand information |  |

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

<a id="gamegameidplayerpost"></a>
# **GameGameIdPlayerPost**
> JustResponseMessageAny GameGameIdPlayerPost (string gameId, Object body = null)

Join a Game

Join an open game


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to join |  |
| **body** | **Object** |  | [optional]  |

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

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

<a id="gamegameidstartedpost"></a>
# **GameGameIdStartedPost**
> JustResponseMessageAny GameGameIdStartedPost (string gameId, Object body = null)

Start Game

starts a game from a created game lobby, closing it to joins and starting play


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | the id of the game to start |  |
| **body** | **Object** |  | [optional]  |

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

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

<a id="gamegameidstateget"></a>
# **GameGameIdStateGet**
> JustResponseMessageGameGameDTO GameGameIdStateGet (string gameId)

Game State

gets the current state of the game from the perspective of the requesting user


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to get the state of |  |

### Return type

[**JustResponseMessageGameGameDTO**](JustResponseMessageGameGameDTO.md)

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

<a id="gamegameidstatelistenget"></a>
# **GameGameIdStateListenGet**
> GameGameDTO GameGameIdStateListenGet (string gameId)

Get Listener

creates a listener that will begin buffering game events that can be queried from an endpoint


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to get events from |  |

### Return type

[**GameGameDTO**](GameGameDTO.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gamegameidstatelistenpost"></a>
# **GameGameIdStateListenPost**
> void GameGameIdStateListenPost (string gameId)

Register Listener

creates a listener that will begin buffering game events that can be queried from an endpoint


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to listen to |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gamegameidstatepost"></a>
# **GameGameIdStatePost**
> JustResponseMessageString GameGameIdStatePost (string gameId, GameTableDTO gameTableDTO)

start game from state

starts a game from a specific state


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | the id of the game to start |  |
| **gameTableDTO** | [**GameTableDTO**](GameTableDTO.md) | the game state object to start game from |  |

### Return type

[**JustResponseMessageString**](JustResponseMessageString.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | game created - game id as string |  -  |
| **400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gamegameidstatewsget"></a>
# **GameGameIdStateWsGet**
> Object GameGameIdStateWsGet (string gameId)

Connect Updates

gets all game updates


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameId** | **string** | ID of the Game to get events from |  |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gameget"></a>
# **GameGet**
> List&lt;GameActiveGameDTO&gt; GameGet ()

Gets Active Games

gets all games that are currently being played


### Parameters
This endpoint does not need any parameter.
### Return type

[**List&lt;GameActiveGameDTO&gt;**](GameActiveGameDTO.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gamepost"></a>
# **GamePost**
> JustResponseMessageString GamePost (GameNewGameConfigDTO gameNewGameConfigDTO)

Create Game

creates a new game from a configuration file


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameNewGameConfigDTO** | [**GameNewGameConfigDTO**](GameNewGameConfigDTO.md) | an object defining configuration information for the new game |  |

### Return type

[**JustResponseMessageString**](JustResponseMessageString.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | game created - game id as string |  -  |
| **400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="handevaluatorevaluatepost"></a>
# **HandEvaluatorEvaluatePost**
> JustHandEvaluationDTO HandEvaluatorEvaluatePost (List<GameCardDTO> gameCardDTO)

Evaluate a Hand

Evaluator? I hardly...


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **gameCardDTO** | [**List&lt;GameCardDTO&gt;**](GameCardDTO.md) | hand to evaluate, either 5 or 7 cards |  |

### Return type

[**JustHandEvaluationDTO**](JustHandEvaluationDTO.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

