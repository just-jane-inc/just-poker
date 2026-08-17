using System.Runtime.CompilerServices;
using System.Text.Json;
using JustPoker.OpenApi.Model;
using JustPoker.Sdk;
using Microsoft.Extensions.Logging;

namespace JustPoker.Listener;

internal static class Program
{
    public static string GetSourceFilePathName( [CallerFilePath] string? callerFilePath = null )
        => callerFilePath ?? "";
    
    private static async Task<int> Main(string[] args)
    {
        var options = ExampleOptions.Parse(args);
        if (options is null)
        {
            await Console.Error.WriteLineAsync("usage: Example --url <base-url> --game-id <id> --user <jill>");
            return 1;
        }

        using var loggerFactory =
            LoggerFactory.Create(builder => builder.AddConsole().SetMinimumLevel(LogLevel.Information));
        var logger = loggerFactory.CreateLogger("listener");

        var configFile = Path.Combine(Path.GetDirectoryName(GetSourceFilePathName())!, "../Example/users.json");
        var users = JsonSerializer.Deserialize<Dictionary<string, UserConfig>>(await File.ReadAllTextAsync(configFile));
        
        if (users is null)
        {
            logger.LogError("users config is missing");
            return 1;
        }

        if (!users.TryGetValue(options.User, out var user))
        {
            logger.LogError("Invalid user '{user}'", options.User);
            return 1;
        }
        
        if (string.IsNullOrEmpty(user.Token) || string.IsNullOrEmpty(user.UserId))
        {
            logger.LogError("Invalid user '{user}' missing token or user_id", options.User);
            return 1;
        }

        await using var bot = new PokerBot(
            baseUrl: options.BaseUrl,
            token: user.Token,
            userId: user.UserId,
            gameId: options.GameId,
            logger: logger);
        
        var gameOver = new TaskCompletionSource();
        
        using var stateSubscription = bot.Events.Subscribe(PokerEventType.All, e => OnEventAsync(e, gameOver, logger));

        // start up
        await bot.InitializeAsync();
        
        // wait until game ends
        await gameOver.Task;

        // should unsubscribe everything
        await bot.Events.DisposeAsync();
        return 0;
    }
    
    private static void OnEventAsync(PokerEvent pokerEvent, TaskCompletionSource gameOverTask, ILogger logger)
    {

        logger.LogInformation($"[{pokerEvent.EventType}]:  {pokerEvent.RawData}");
        
        if (pokerEvent.EventType == PokerEventType.GameOver)
        {
            gameOverTask.SetResult();
        }
    }

    private sealed record UserConfig
    {
        [System.Text.Json.Serialization.JsonPropertyName("token")]
        public string Token { get; init; } = string.Empty;

        [System.Text.Json.Serialization.JsonPropertyName("user_id")]
        public string UserId { get; init; } = string.Empty;
    }
    
    private sealed record ExampleOptions(string BaseUrl, string GameId, string User)
    {
        public static ExampleOptions? Parse(string[] args)
        {
            string? url = null;
            string? gameId = null;
            string? user = null;

            for (var i = 0; i + 1 < args.Length; i += 2)
            {
                switch (args[i])
                {
                    case "--url":
                        url = args[i + 1];
                        break;
                    case "--game-id":
                    case "--id":
                        gameId = args[i + 1];
                        break;
                    case "--user":
                        user = args[i + 1];
                        break;
                }
            }

            if (url is null || gameId is null || user is null)
            {
                return null;
            }

            return new ExampleOptions(url, gameId, user);
        }
    }
}
