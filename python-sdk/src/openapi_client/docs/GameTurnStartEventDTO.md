# GameTurnStartEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**bet_amount** | **int** |  | [optional] 
**player_id** | **str** |  | [optional] 
**round_type** | [**GameRoundType**](GameRoundType.md) |  | [optional] 

## Example

```python
from openapi_client.models.game_turn_start_event_dto import GameTurnStartEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameTurnStartEventDTO from a JSON string
game_turn_start_event_dto_instance = GameTurnStartEventDTO.from_json(json)
# print the JSON string representation of the object
print(GameTurnStartEventDTO.to_json())

# convert the object into a dict
game_turn_start_event_dto_dict = game_turn_start_event_dto_instance.to_dict()
# create an instance of GameTurnStartEventDTO from a dict
game_turn_start_event_dto_from_dict = GameTurnStartEventDTO.from_dict(game_turn_start_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


