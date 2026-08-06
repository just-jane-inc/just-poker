# AdminGameGameIdTablePostRequest


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
from openapi_client.models.admin_game_game_id_table_post_request import AdminGameGameIdTablePostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AdminGameGameIdTablePostRequest from a JSON string
admin_game_game_id_table_post_request_instance = AdminGameGameIdTablePostRequest.from_json(json)
# print the JSON string representation of the object
print(AdminGameGameIdTablePostRequest.to_json())

# convert the object into a dict
admin_game_game_id_table_post_request_dict = admin_game_game_id_table_post_request_instance.to_dict()
# create an instance of AdminGameGameIdTablePostRequest from a dict
admin_game_game_id_table_post_request_from_dict = AdminGameGameIdTablePostRequest.from_dict(admin_game_game_id_table_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


