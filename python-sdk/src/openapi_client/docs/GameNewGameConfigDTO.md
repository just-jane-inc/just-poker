# GameNewGameConfigDTO

the configuration used to setup the game

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**auto_starts_hands** | **bool** | a flag which indicates true if the game server should wait for a signal to start hands or if it should do so automatically | [optional] 
**big_blind** | **int** | the big blind | [optional] 
**bot_turn_timeout** | **int** | the number of milliseconds that a bot has to take a turn | [optional] 
**chip_denominations** | **List[int]** | a collection of denominations that are available for chips at the table | [optional] 
**player_count** | **int** | the number of players (max) the game supports | [optional] 
**small_blind** | **int** | the small blind | [optional] 
**starting_chips** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 

## Example

```python
from openapi_client.models.game_new_game_config_dto import GameNewGameConfigDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameNewGameConfigDTO from a JSON string
game_new_game_config_dto_instance = GameNewGameConfigDTO.from_json(json)
# print the JSON string representation of the object
print(GameNewGameConfigDTO.to_json())

# convert the object into a dict
game_new_game_config_dto_dict = game_new_game_config_dto_instance.to_dict()
# create an instance of GameNewGameConfigDTO from a dict
game_new_game_config_dto_from_dict = GameNewGameConfigDTO.from_dict(game_new_game_config_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


