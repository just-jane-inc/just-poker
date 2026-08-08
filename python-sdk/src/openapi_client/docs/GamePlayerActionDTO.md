# GamePlayerActionDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**accepted_at** | **str** | a timestamp capturing when a succesful action was accepted by the game | [optional] 
**chips** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**intent** | [**GamePlayerIntent**](GamePlayerIntent.md) |  | [optional] 
**player_id** | **str** | the id of the player preforming the action | [optional] 

## Example

```python
from openapi_client.models.game_player_action_dto import GamePlayerActionDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GamePlayerActionDTO from a JSON string
game_player_action_dto_instance = GamePlayerActionDTO.from_json(json)
# print the JSON string representation of the object
print(GamePlayerActionDTO.to_json())

# convert the object into a dict
game_player_action_dto_dict = game_player_action_dto_instance.to_dict()
# create an instance of GamePlayerActionDTO from a dict
game_player_action_dto_from_dict = GamePlayerActionDTO.from_dict(game_player_action_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


