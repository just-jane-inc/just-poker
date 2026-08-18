# JustPoker.OpenApi.Model.GamePlayerDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentBet** | **Dictionary&lt;string, int&gt;** | an optional mapping of chips that is required by some action types. | [optional] 
**DisplayName** | **string** | the users display name | [optional] 
**Hole** | [**List&lt;GameCardDTO&gt;**](GameCardDTO.md) | the cards current held by this player - only visible for authorized users during a game. | [optional] 
**Position** | **int** | the players position at the table, starting with 0 being the first player sitting clockwise from the dealer | [optional] 
**PotContribution** | **int** | the sum total the player has contributed to the pot, note that this does not include chips currently in CurrentBet | [optional] 
**Stack** | **Dictionary&lt;string, int&gt;** | an optional mapping of chips that is required by some action types. | [optional] 
**State** | **string** | the players state | [optional] 
**UserId** | **string** | the id of the user | [optional] 
**UserType** | **JustUserType** |  | [optional] 

[[Back to Model list]](../../README.md#documentation-for-models) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to README]](../../README.md)

