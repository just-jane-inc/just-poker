# GameTableDTO

the table

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**big_blind_position** | **int** | the position of the big blind | [optional] 
**button_position** | **int** | the position of the button | [optional] 
**current_hand** | [**GameHandDTO**](GameHandDTO.md) |  | [optional] 
**current_round** | [**GameRoundDTO**](GameRoundDTO.md) |  | [optional] 
**players** | [**List[GamePlayerDTO]**](GamePlayerDTO.md) | An array of players at the table | [optional] 
**pot** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**small_blind_position** | **int** | the position of the small blind | [optional] 
**street** | [**List[GameCardDTO]**](GameCardDTO.md) | the cards that are on the street (community cards) | [optional] 

## Example

```python
from openapi_client.models.game_table_dto import GameTableDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameTableDTO from a JSON string
game_table_dto_instance = GameTableDTO.from_json(json)
# print the JSON string representation of the object
print(GameTableDTO.to_json())

# convert the object into a dict
game_table_dto_dict = game_table_dto_instance.to_dict()
# create an instance of GameTableDTO from a dict
game_table_dto_from_dict = GameTableDTO.from_dict(game_table_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


