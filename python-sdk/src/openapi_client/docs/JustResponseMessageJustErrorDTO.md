# JustResponseMessageJustErrorDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**JustErrorDTO**](JustErrorDTO.md) |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_response_message_just_error_dto import JustResponseMessageJustErrorDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustResponseMessageJustErrorDTO from a JSON string
just_response_message_just_error_dto_instance = JustResponseMessageJustErrorDTO.from_json(json)
# print the JSON string representation of the object
print(JustResponseMessageJustErrorDTO.to_json())

# convert the object into a dict
just_response_message_just_error_dto_dict = just_response_message_just_error_dto_instance.to_dict()
# create an instance of JustResponseMessageJustErrorDTO from a dict
just_response_message_just_error_dto_from_dict = JustResponseMessageJustErrorDTO.from_dict(just_response_message_just_error_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


