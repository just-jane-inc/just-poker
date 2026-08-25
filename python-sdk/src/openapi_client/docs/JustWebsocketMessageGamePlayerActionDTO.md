# JustWebsocketMessageGamePlayerActionDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**GamePlayerActionDTO**](GamePlayerActionDTO.md) |  | [optional] 
**event_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**time_sent** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_websocket_message_game_player_action_dto import JustWebsocketMessageGamePlayerActionDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustWebsocketMessageGamePlayerActionDTO from a JSON string
just_websocket_message_game_player_action_dto_instance = JustWebsocketMessageGamePlayerActionDTO.from_json(json)
# print the JSON string representation of the object
print(JustWebsocketMessageGamePlayerActionDTO.to_json())

# convert the object into a dict
just_websocket_message_game_player_action_dto_dict = just_websocket_message_game_player_action_dto_instance.to_dict()
# create an instance of JustWebsocketMessageGamePlayerActionDTO from a dict
just_websocket_message_game_player_action_dto_from_dict = JustWebsocketMessageGamePlayerActionDTO.from_dict(just_websocket_message_game_player_action_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


