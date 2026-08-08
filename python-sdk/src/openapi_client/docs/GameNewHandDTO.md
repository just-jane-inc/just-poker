# GameNewHandDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**deck** | [**List[GameCardDTO]**](GameCardDTO.md) | the deck, optionally provided to determine trhe order of cards if not provided the cards will be ordered randomly by the server | [optional] 

## Example

```python
from openapi_client.models.game_new_hand_dto import GameNewHandDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameNewHandDTO from a JSON string
game_new_hand_dto_instance = GameNewHandDTO.from_json(json)
# print the JSON string representation of the object
print(GameNewHandDTO.to_json())

# convert the object into a dict
game_new_hand_dto_dict = game_new_hand_dto_instance.to_dict()
# create an instance of GameNewHandDTO from a dict
game_new_hand_dto_from_dict = GameNewHandDTO.from_dict(game_new_hand_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


