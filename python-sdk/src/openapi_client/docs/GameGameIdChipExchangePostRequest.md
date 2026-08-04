# GameGameIdChipExchangePostRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**give** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**receive** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 

## Example

```python
from openapi_client.models.game_game_id_chip_exchange_post_request import GameGameIdChipExchangePostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameIdChipExchangePostRequest from a JSON string
game_game_id_chip_exchange_post_request_instance = GameGameIdChipExchangePostRequest.from_json(json)
# print the JSON string representation of the object
print(GameGameIdChipExchangePostRequest.to_json())

# convert the object into a dict
game_game_id_chip_exchange_post_request_dict = game_game_id_chip_exchange_post_request_instance.to_dict()
# create an instance of GameGameIdChipExchangePostRequest from a dict
game_game_id_chip_exchange_post_request_from_dict = GameGameIdChipExchangePostRequest.from_dict(game_game_id_chip_exchange_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


