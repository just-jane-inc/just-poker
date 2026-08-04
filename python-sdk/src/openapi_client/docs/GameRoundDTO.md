# GameRoundDTO

the current round

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**bet** | **int** | the amount of chips required to call (how does this work with split pots?) | [optional] 
**current_aggressor** | **int** | the player who currently ends the round? the last raise? | [optional] 
**current_player_position** | **int** | the index into the play array for the player whose turn it currently is | [optional] 
**current_round_type** | [**GameRoundType**](GameRoundType.md) |  | [optional] 

## Example

```python
from openapi_client.models.game_round_dto import GameRoundDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameRoundDTO from a JSON string
game_round_dto_instance = GameRoundDTO.from_json(json)
# print the JSON string representation of the object
print(GameRoundDTO.to_json())

# convert the object into a dict
game_round_dto_dict = game_round_dto_instance.to_dict()
# create an instance of GameRoundDTO from a dict
game_round_dto_from_dict = GameRoundDTO.from_dict(game_round_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


