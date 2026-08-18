# Created with Openapi Generator

<a id="cli"></a>
## Creating the library
Create a config.yaml file similar to what is below, then run the following powershell command to generate the library `java -jar "<path>/openapi-generator/modules/openapi-generator-cli/target/openapi-generator-cli.jar" generate -c config.yaml`

```yaml
generatorName: csharp
inputSpec: ..\server\docs\swagger.json
outputDir: out

# https://openapi-generator.tech/docs/generators/csharp
additionalProperties:
  packageGuid: '{2D10A01B-476C-4376-BDAE-D11C7F7C7499}'

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
- appDescription: # Error Format &amp; Meaning  All response messages are in the format of: &#x60;&#x60;&#x60;json {   \&quot;message_type\&quot;: \&quot;string\&quot;,   \&quot;data\&quot;: {} } &#x60;&#x60;&#x60;  All errors use the \&quot;error\&quot; message_type and have the data format of: &#x60;&#x60;&#x60;json {   \&quot;error_code\&quot;: int,   \&quot;error\&quot;: \&quot;string\&quot; } &#x60;&#x60;&#x60;   The \&quot;error\&quot; field is intended only as a human readable message and should not be depended on for branching or parsing in code. This field should be considered volatile within the api. The error_code field is stable and intended for client code to use within recovery operations. The error codes and their meanings are enumerated in the table below   |   Code | Name                  | Meaning                                       | | - -- --: | - -- -- -- -- -- -- -- -- -- -- | - -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- | | &#x60;1000&#x60; | &#x60;Unknown&#x60; | A novel error...congrats.                | | &#x60;1001&#x60; | &#x60;TurnOrderViolation&#x60;     | Player action received for a player out of turn. | | &#x60;1002&#x60; | &#x60;InvalidActionType&#x60;   | Player action received with an intent that is not valid for the game state.                        | | &#x60;1003&#x60; | &#x60;NotEnoughChips&#x60;   | Player action received with a chip amount that cannot be covered by the players stack            | | &#x60;1004&#x60; | &#x60;InvalidBetAmount&#x60;    | Player action received with a bet amount that violates requirements i.e. incorrect amount for a blind, call, or a raise which is too low.                 |   

## Build
This C# SDK is automatically generated by the [OpenAPI Generator](https://openapi-generator.tech) project.

- SDK version: 1.0.0
- Generator version: 7.24.0
- Build package: org.openapitools.codegen.languages.CSharpClientCodegen
