# JustWebsocketMessageGameRoundStartEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**GameRoundStartEventDTO**](GameRoundStartEventDTO.md) |  | [optional] 
**event_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**time_sent** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_websocket_message_game_round_start_event_dto import JustWebsocketMessageGameRoundStartEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustWebsocketMessageGameRoundStartEventDTO from a JSON string
just_websocket_message_game_round_start_event_dto_instance = JustWebsocketMessageGameRoundStartEventDTO.from_json(json)
# print the JSON string representation of the object
print(JustWebsocketMessageGameRoundStartEventDTO.to_json())

# convert the object into a dict
just_websocket_message_game_round_start_event_dto_dict = just_websocket_message_game_round_start_event_dto_instance.to_dict()
# create an instance of JustWebsocketMessageGameRoundStartEventDTO from a dict
just_websocket_message_game_round_start_event_dto_from_dict = JustWebsocketMessageGameRoundStartEventDTO.from_dict(just_websocket_message_game_round_start_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


