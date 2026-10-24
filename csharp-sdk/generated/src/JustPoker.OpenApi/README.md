# Created with Openapi Generator

<a id="cli"></a>
## Creating the library
Create a config.yaml file similar to what is below, then run the following powershell command to generate the library `java -jar "<path>/openapi-generator/modules/openapi-generator-cli/target/openapi-generator-cli.jar" generate -c config.yaml`

```yaml
generatorName: csharp
inputSpec: ../server/docs/swagger.json
outputDir: out

# https://openapi-generator.tech/docs/generators/csharp
additionalProperties:
  packageGuid: '{B79287C9-8E85-4D8D-A51C-89E1E7F5189C}'

# https://openapi-generator.tech/docs/integrations/#github-integration
# gitHost:
# gitUserId:
# gitRepoId:

# https://openapi-generator.tech/docs/globals
# globalProperties:

# https://openapi-generator.tech/docs/customization/#inline-schema-naming
# inlineSchemaOptions:

# https://openapi-generator.tech/docs/customization/#name-mapping
# modelNameMappings:
# nameMappings:

# https://openapi-generator.tech/docs/customization/#openapi-normalizer
# openapiNormalizer:

# templateDir: https://openapi-generator.tech/docs/templating/#modifying-templates

# releaseNote:
```

<a id="usage"></a>
## Using the library in your project

```cs
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.DependencyInjection;
using JustPoker.OpenApi.Api;
using JustPoker.OpenApi.Client;
using JustPoker.OpenApi.Model;
using Org.OpenAPITools.Extensions;

namespace YourProject
{
    public class Program
    {
        public static async Task Main(string[] args)
        {
            var host = CreateHostBuilder(args).Build();
            var api = host.Services.GetRequiredService<IAdminApi>();
            IAdminGameGameIdStatusPostApiResponse apiResponse = await api.AdminGameGameIdStatusPostAsync("todo");
            Object? model = apiResponse.Ok();
        }

        public static IHostBuilder CreateHostBuilder(string[] args) => Host.CreateDefaultBuilder(args)
          .ConfigureApi((context, options) =>
          {
              // The type of token here depends on the api security specifications
              // Available token types are ApiKeyToken, BasicToken, BearerToken, HttpSigningToken, and OAuthToken.
              BearerToken token = new("<your token>");
              options.AddTokens(token);

              // optionally choose the method the tokens will be provided with, default is RateLimitProvider
              options.UseProvider<RateLimitProvider<BearerToken>, BearerToken>();

              options.ConfigureJsonOptions((jsonOptions) =>
              {
                  // your custom converters if any
              });

              options.AddApiHttpClients(client =>
              {
                  // client configuration
              }, builder =>
              {
                  builder
                      .AddRetryPolicy(2)
                      .AddTimeoutPolicy(TimeSpan.FromSeconds(5))
                      .AddCircuitBreakerPolicy(10, TimeSpan.FromSeconds(30));
                      // add whatever middleware you prefer
                  }
              );
          });
    }
}
```
<a id="questions"></a>
## Questions

- What about HttpRequest failures and retries?
  Configure Polly in the IHttpClientBuilder
- How are tokens used?
  Tokens are provided by a TokenProvider class. The default is RateLimitProvider which will perform client side rate limiting.
  Other providers can be used with the UseProvider method.
- Does an HttpRequest throw an error when the server response is not Ok?
  It depends how you made the request. If the return type is ApiResponse<T> no error will be thrown, though the Content property will be null.
  StatusCode and ReasonPhrase will contain information about the error.
  If the return type is T, then it will throw. If the return type is TOrDefault, it will return null.
- How do I validate requests and process responses?
  Use the provided On and After partial methods in the api classes.

## Api Information
- appName: BAHMS Poker Tournament
- appVersion: 1.0
- appDescription: # Error Format &amp; Meaning  All response messages are in the format of: &#x60;&#x60;&#x60;json {   \&quot;message_type\&quot;: \&quot;string\&quot;,   \&quot;data\&quot;: {} } &#x60;&#x60;&#x60;  All errors use the \&quot;error\&quot; message_type and have the data format of: &#x60;&#x60;&#x60;json {   \&quot;error_code\&quot;: int,   \&quot;error\&quot;: \&quot;string\&quot; } &#x60;&#x60;&#x60;   The \&quot;error\&quot; field is intended only as a human readable message and should not be depended on for branching or parsing in code. This field should be considered volatile within the api. The error_code field is stable and intended for client code to use within recovery operations. The error codes and their meanings are enumerated bellow   |   Code | Name                  | Meaning                                       | | - -- --: | - -- -- -- -- -- -- -- -- -- -- | - -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- | | 67 | SkillIssue | you have a skill issue | 1000 | Unknown | an unknown error occurred, likely an issue in the server | 1001 | UserNotFound | a user referenced by id did not exist | 1002 | GameNotFound | a game referenced by id did not exist | 1003 | TokenMissing | a user request requiring authorization did not include a token | 1004 | MalformedRequestBody | request parse error | 2000 | TurnOrderViolation | player attempted to move out of order | 2001 | InvalidActionType | action produced by player was an invalid type | 2002 | NotEnoughChips | bet amount did not have sufficient chips | 2003 | InvalidBetAmount | bet amount is not valid | 2004 | InvalidChipExchange | a chip exchange was imbalanced or otherwise invalid | 2020 | GameAlreadyStarted | attempt to join after game started | 2021 | GameIsFull | attempt to join a full game | 2022 | PlayerAlreadyJoined | player is already joined to this game | 2023 | GameIsPaused | action attempted while the game is paused | 2024 | InvalidGameConfiguration | game configuration is not valid | 2025 | HandAlreadyInProgress | attempt to start a hand while one is already in progress | 2026 | InvalidChipDenomination | a chip denomination is invalid e.g. negative | 2027 | InvalidChipCount | a chip count is invalid e.g. negative  

## Build
This C# SDK is automatically generated by the [OpenAPI Generator](https://openapi-generator.tech) project.

- SDK version: 1.0.0
- Generator version: 7.24.0
- Build package: org.openapitools.codegen.languages.CSharpClientCodegen
