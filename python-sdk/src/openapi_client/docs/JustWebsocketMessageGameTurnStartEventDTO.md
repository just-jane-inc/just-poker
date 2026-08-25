# JustWebsocketMessageGameTurnStartEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**GameTurnStartEventDTO**](GameTurnStartEventDTO.md) |  | [optional] 
**event_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**time_sent** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_websocket_message_game_turn_start_event_dto import JustWebsocketMessageGameTurnStartEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustWebsocketMessageGameTurnStartEventDTO from a JSON string
just_websocket_message_game_turn_start_event_dto_instance = JustWebsocketMessageGameTurnStartEventDTO.from_json(json)
# print the JSON string representation of the object
print(JustWebsocketMessageGameTurnStartEventDTO.to_json())

# convert the object into a dict
just_websocket_message_game_turn_start_event_dto_dict = just_websocket_message_game_turn_start_event_dto_instance.to_dict()
# create an instance of JustWebsocketMessageGameTurnStartEventDTO from a dict
just_websocket_message_game_turn_start_event_dto_from_dict = JustWebsocketMessageGameTurnStartEventDTO.from_dict(just_websocket_message_game_turn_start_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


