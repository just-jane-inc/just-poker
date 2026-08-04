# GamePlayerDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**current_bet** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**display_name** | **str** | the users display name | [optional] 
**hole** | [**List[GameCardDTO]**](GameCardDTO.md) | the cards current held by this player - only visible for authorized users during a game. | [optional] 
**position** | **int** | the players position at the table, starting with 0 being the first player sitting clockwise from the dealer | [optional] 
**pot_contribution** | **int** | the sum total the player has contributed to the pot, note that this does not include chips currently in CurrentBet | [optional] 
**stack** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**state** | **str** | the players state | [optional] 
**user_id** | **str** | the is of the user | [optional] 
**user_type** | **str** | the type of user TODO: make this an enum | [optional] 

## Example

```python
from openapi_client.models.game_player_dto import GamePlayerDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GamePlayerDTO from a JSON string
game_player_dto_instance = GamePlayerDTO.from_json(json)
# print the JSON string representation of the object
print(GamePlayerDTO.to_json())

# convert the object into a dict
game_player_dto_dict = game_player_dto_instance.to_dict()
# create an instance of GamePlayerDTO from a dict
game_player_dto_from_dict = GamePlayerDTO.from_dict(game_player_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


