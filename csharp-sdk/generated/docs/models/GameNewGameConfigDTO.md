# JustPoker.OpenApi.Model.GameNewGameConfigDTO
the configuration used to setup the game

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AutoStartsHands** | **bool** | a flag which indicates true if the game server should wait for a signal to start hands or if it should do so automatically | [optional] 
**BigBlind** | **int** | the big blind | [optional] 
**ChipDenominations** | **List&lt;int&gt;** | a collection of denominations that are available for chips at the table | [optional] 
**PlayerCount** | **int** | the number of players (max) the game supports | [optional] 
**SmallBlind** | **int** | the small blind | [optional] 
**StartingChips** | **Dictionary&lt;string, int&gt;** | an optional mapping of chips that is required by some action types. | [optional] 

[[Back to Model list]](../../README.md#documentation-for-models) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to README]](../../README.md)

