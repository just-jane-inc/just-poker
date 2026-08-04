# GameCardDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**rank** | **int** | the rank of a card as a rune - int32 ASCII encoding | [optional] 
**suit** | **int** | the suit of a card as a rune - int32 ASCII encoding | [optional] 

## Example

```python
from openapi_client.models.game_card_dto import GameCardDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameCardDTO from a JSON string
game_card_dto_instance = GameCardDTO.from_json(json)
# print the JSON string representation of the object
print(GameCardDTO.to_json())

# convert the object into a dict
game_card_dto_dict = game_card_dto_instance.to_dict()
# create an instance of GameCardDTO from a dict
game_card_dto_from_dict = GameCardDTO.from_dict(game_card_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


