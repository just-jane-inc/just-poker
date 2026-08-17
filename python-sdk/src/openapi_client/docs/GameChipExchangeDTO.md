# GameChipExchangeDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**give** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**receive** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**user_id** | **str** | the stack to give during the exchange | [optional] 

## Example

```python
from openapi_client.models.game_chip_exchange_dto import GameChipExchangeDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameChipExchangeDTO from a JSON string
game_chip_exchange_dto_instance = GameChipExchangeDTO.from_json(json)
# print the JSON string representation of the object
print(GameChipExchangeDTO.to_json())

# convert the object into a dict
game_chip_exchange_dto_dict = game_chip_exchange_dto_instance.to_dict()
# create an instance of GameChipExchangeDTO from a dict
game_chip_exchange_dto_from_dict = GameChipExchangeDTO.from_dict(game_chip_exchange_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


