# openapi_client.GameApi

All URIs are relative to *https://game.bahms.org/api/poker*

Method | HTTP request | Description
------------- | ------------- | -------------
[**game_game_id_action_post**](GameApi.md#game_game_id_action_post) | **POST** /game/{game_id}/action | Player Action
[**game_game_id_chip_exchange_post**](GameApi.md#game_game_id_chip_exchange_post) | **POST** /game/{game_id}/chip/exchange | Exchange Chips
[**game_game_id_player_post**](GameApi.md#game_game_id_player_post) | **POST** /game/{game_id}/player | Join a Game
[**game_game_id_started_post**](GameApi.md#game_game_id_started_post) | **POST** /game/{game_id}/started | Start Game
[**game_game_id_state_get**](GameApi.md#game_game_id_state_get) | **GET** /game/{game_id}/state | Game State
[**game_game_id_state_listen_get**](GameApi.md#game_game_id_state_listen_get) | **GET** /game/{game_id}/state/listen | Get Listener
[**game_game_id_state_listen_post**](GameApi.md#game_game_id_state_listen_post) | **POST** /game/{game_id}/state/listen | Register Listener
[**game_game_id_state_ws_get**](GameApi.md#game_game_id_state_ws_get) | **GET** /game/{game_id}/state/ws | Connect Updates
[**game_get**](GameApi.md#game_get) | **GET** /game | Gets Active Games
[**game_post**](GameApi.md#game_post) | **POST** /game | Create Game
[**hand_evaluator_evaluate_post**](GameApi.md#hand_evaluator_evaluate_post) | **POST** /hand-evaluator/evaluate | Evaluate a Hand


# **game_game_id_action_post**
> JustResponseMessageAny game_game_id_action_post(game_id, game_game_id_action_post_request)

Player Action

post the action preformed by a player

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.game_game_id_action_post_request import GameGameIdActionPostRequest
from openapi_client.models.just_response_message_any import JustResponseMessageAny
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | the id of the game
    game_game_id_action_post_request = openapi_client.GameGameIdActionPostRequest() # GameGameIdActionPostRequest | the action the player is preforming

    try:
        # Player Action
        api_response = await api_instance.game_game_id_action_post(game_id, game_game_id_action_post_request)
        print("The response of GameApi->game_game_id_action_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_action_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| the id of the game | 
 **game_game_id_action_post_request** | [**GameGameIdActionPostRequest**](GameGameIdActionPostRequest.md)| the action the player is preforming | 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_chip_exchange_post**
> JustResponseMessageAny game_game_id_chip_exchange_post(game_id, game_game_id_chip_exchange_post_request)

Exchange Chips

exchange chips in the players stack with the tables rack

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.game_game_id_chip_exchange_post_request import GameGameIdChipExchangePostRequest
from openapi_client.models.just_response_message_any import JustResponseMessageAny
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game exchange chips in
    game_game_id_chip_exchange_post_request = openapi_client.GameGameIdChipExchangePostRequest() # GameGameIdChipExchangePostRequest | a specification for the chips to exchange

    try:
        # Exchange Chips
        api_response = await api_instance.game_game_id_chip_exchange_post(game_id, game_game_id_chip_exchange_post_request)
        print("The response of GameApi->game_game_id_chip_exchange_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_chip_exchange_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game exchange chips in | 
 **game_game_id_chip_exchange_post_request** | [**GameGameIdChipExchangePostRequest**](GameGameIdChipExchangePostRequest.md)| a specification for the chips to exchange | 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_player_post**
> JustResponseMessageAny game_game_id_player_post(game_id, body=body)

Join a Game

Join an open game

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.just_response_message_any import JustResponseMessageAny
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to join
    body = None # object |  (optional)

    try:
        # Join a Game
        api_response = await api_instance.game_game_id_player_post(game_id, body=body)
        print("The response of GameApi->game_game_id_player_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_player_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to join | 
 **body** | **object**|  | [optional] 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_started_post**
> JustResponseMessageAny game_game_id_started_post(game_id, body=body)

Start Game

starts a game from a created game lobby, closing it to joins and starting play

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.just_response_message_any import JustResponseMessageAny
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | the id of the game to start
    body = None # object |  (optional)

    try:
        # Start Game
        api_response = await api_instance.game_game_id_started_post(game_id, body=body)
        print("The response of GameApi->game_game_id_started_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_started_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| the id of the game to start | 
 **body** | **object**|  | [optional] 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_state_get**
> JustResponseMessageGameGameDTO game_game_id_state_get(game_id)

Game State

gets the current state of the game from the perspective of the requesting user

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.just_response_message_game_game_dto import JustResponseMessageGameGameDTO
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to get the state of

    try:
        # Game State
        api_response = await api_instance.game_game_id_state_get(game_id)
        print("The response of GameApi->game_game_id_state_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_state_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to get the state of | 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_state_listen_get**
> GameGameDTO game_game_id_state_listen_get(game_id)

Get Listener

creates a listener that will begin buffering game events that can be queried from an endpoint

### Example


```python
import openapi_client
from openapi_client.models.game_game_dto import GameGameDTO
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)


# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to get events from

    try:
        # Get Listener
        api_response = await api_instance.game_game_id_state_listen_get(game_id)
        print("The response of GameApi->game_game_id_state_listen_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_state_listen_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to get events from | 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_state_listen_post**
> game_game_id_state_listen_post(game_id)

Register Listener

creates a listener that will begin buffering game events that can be queried from an endpoint

### Example


```python
import openapi_client
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)


# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to listen to

    try:
        # Register Listener
        await api_instance.game_game_id_state_listen_post(game_id)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_state_listen_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to listen to | 

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
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_game_id_state_ws_get**
> object game_game_id_state_ws_get(game_id)

Connect Updates

gets all game updates

### Example


```python
import openapi_client
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)


# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to get events from

    try:
        # Connect Updates
        api_response = await api_instance.game_game_id_state_ws_get(game_id)
        print("The response of GameApi->game_game_id_state_ws_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_game_id_state_ws_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to get events from | 

### Return type

**object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_get**
> List[GameActiveGameDTO] game_get()

Gets Active Games

gets all games that are currently being played

### Example


```python
import openapi_client
from openapi_client.models.game_active_game_dto import GameActiveGameDTO
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)


# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)

    try:
        # Gets Active Games
        api_response = await api_instance.game_get()
        print("The response of GameApi->game_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[GameActiveGameDTO]**](GameActiveGameDTO.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **game_post**
> JustResponseMessageString game_post(game_post_request)

Create Game

creates a new game from a configuration file

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.game_post_request import GamePostRequest
from openapi_client.models.just_response_message_string import JustResponseMessageString
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    game_post_request = openapi_client.GamePostRequest() # GamePostRequest | an object defining configuration information for the new game

    try:
        # Create Game
        api_response = await api_instance.game_post(game_post_request)
        print("The response of GameApi->game_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->game_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_post_request** | [**GamePostRequest**](GamePostRequest.md)| an object defining configuration information for the new game | 

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
**200** | game created - game id as string |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **hand_evaluator_evaluate_post**
> object hand_evaluator_evaluate_post(hand_evaluator_evaluate_post_request)

Evaluate a Hand

Evaluator? I hardly...

### Example


```python
import openapi_client
from openapi_client.models.hand_evaluator_evaluate_post_request import HandEvaluatorEvaluatePostRequest
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)


# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.GameApi(api_client)
    hand_evaluator_evaluate_post_request = openapi_client.HandEvaluatorEvaluatePostRequest() # HandEvaluatorEvaluatePostRequest | hand to evaluate, either 5 or 7 cards

    try:
        # Evaluate a Hand
        api_response = await api_instance.hand_evaluator_evaluate_post(hand_evaluator_evaluate_post_request)
        print("The response of GameApi->hand_evaluator_evaluate_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling GameApi->hand_evaluator_evaluate_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **hand_evaluator_evaluate_post_request** | [**HandEvaluatorEvaluatePostRequest**](HandEvaluatorEvaluatePostRequest.md)| hand to evaluate, either 5 or 7 cards | 

### Return type

**object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

