# JustResponseMessageString


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | **str** |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.just_response_message_string import JustResponseMessageString

# TODO update the JSON string below
json = "{}"
# create an instance of JustResponseMessageString from a JSON string
just_response_message_string_instance = JustResponseMessageString.from_json(json)
# print the JSON string representation of the object
print(JustResponseMessageString.to_json())

# convert the object into a dict
just_response_message_string_dict = just_response_message_string_instance.to_dict()
# create an instance of JustResponseMessageString from a dict
just_response_message_string_from_dict = JustResponseMessageString.from_dict(just_response_message_string_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


