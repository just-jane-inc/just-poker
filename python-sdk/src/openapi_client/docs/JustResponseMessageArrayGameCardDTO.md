# JustResponseMessageArrayGameCardDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**List[GameCardDTO]**](GameCardDTO.md) |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_response_message_array_game_card_dto import JustResponseMessageArrayGameCardDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustResponseMessageArrayGameCardDTO from a JSON string
just_response_message_array_game_card_dto_instance = JustResponseMessageArrayGameCardDTO.from_json(json)
# print the JSON string representation of the object
print(JustResponseMessageArrayGameCardDTO.to_json())

# convert the object into a dict
just_response_message_array_game_card_dto_dict = just_response_message_array_game_card_dto_instance.to_dict()
# create an instance of JustResponseMessageArrayGameCardDTO from a dict
just_response_message_array_game_card_dto_from_dict = JustResponseMessageArrayGameCardDTO.from_dict(just_response_message_array_game_card_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


