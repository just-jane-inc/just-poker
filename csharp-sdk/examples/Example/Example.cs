using System.Runtime.CompilerServices;
using System.Text.Json;
using System.Text.Json.Serialization;
using JustPoker.OpenApi.Model;
using JustPoker.Sdk;
using Microsoft.Extensions.Logging;

namespace JustPoker.Examples;

internal static class Program {
    public static string GetSourceFilePathName([CallerFilePath] string? callerFilePath = null) {
        return callerFilePath ?? "";
    }

    private static async Task<int> Main(string[] args) {
        var options = ExampleOptions.Parse(args);
        if (options is null) {
            await Console.Error.WriteLineAsync("usage: Example --url <base-url> --game-id <id> --user <wolf>");
            return 1;
        }

        using var loggerFactory =
            LoggerFactory.Create(builder => builder.AddConsole().SetMinimumLevel(LogLevel.Information));
        var logger = loggerFactory.CreateLogger("example");

        var configFile = Path.Combine(Path.GetDirectoryName(GetSourceFilePathName())!, "users.json");
        var users = JsonSerializer.Deserialize<Dictionary<string, UserConfig>>(await File.ReadAllTextAsync(configFile));

        if (users is null) {
            logger.LogError("users config is missing");
            return 1;
        }

        if (!users.TryGetValue(options.User, out var user)) {
            logger.LogError("Invalid user '{user}'", options.User);
            return 1;
        }

        if (string.IsNullOrEmpty(user.Token) || string.IsNullOrEmpty(user.UserId)) {
            logger.LogError("Invalid user '{user}' missing token or user_id", options.User);
            return 1;
        }

        await using var bot = new PokerBot(
            options.BaseUrl,
            user.Token,
            user.UserId,
            options.GameId,
            logger: logger);


        var gameStarted = new TaskCompletionSource();
        var gameOver = new TaskCompletionSource();

        // subscribe before events roll in and before joining game to ensure we do consume them as intended
        using var startedSubscription = bot.Events.Subscribe(PokerEventType.StartingGame, _ => {
            logger.LogInformation("game started");
            gameStarted.TrySetResult();
            return Task.CompletedTask;
        });

        using var overSubscription = bot.Events.Subscribe(PokerEventType.GameOver, _ => {
            logger.LogInformation("game over");
            gameOver.TrySetResult();
            return Task.CompletedTask;
        });

        using var stateSubscription = bot.Events.Subscribe(
            [PokerEventType.GameStateUpdate, PokerEventType.Welcome],
            e => OnGameStateAsync(bot, e, logger));

        // start up
        await bot.InitializeAsync();
        if (!bot.Joined)
            await bot.JoinGameAsync();
        logger.LogInformation("Joined Game");

        if (bot.CurrentState?.StartedAt is null) {
            logger.LogInformation("waiting for the game to start");
            // blocks until game start has been received from a state where it was not set prior
            await gameStarted.Task;
        }

        // wait until game ends
        await gameOver.Task;

        // should unsubscribe everything
        await bot.Events.DisposeAsync();

        logger.LogInformation("yippie");
        return 0;
    }

    private static async Task OnGameStateAsync(PokerBot bot, PokerEvent pokerEvent, ILogger logger) {
        if (pokerEvent.DataAs<GameGameDTO>() is not { } state)
            return;

        if (!bot.IsMyTurn())
            return;

        try {
            var round = state.Table?.CurrentRound;
            if (round is null)
                return;

            var myCurrentBet = PokerHelpers.ChipSum(bot.Player?.CurrentBet ?? new Dictionary<string, int>());
            logger.LogInformation(
                $"You have: {PokerHelpers.ChipSum(bot.Player?.Stack ?? new Dictionary<string, int>())}," +
                $" bet {myCurrentBet}." +
                $" Cards: {string.Join(" ", bot.HoleCards.Select(c => c.ToUnicode()))}");

            logger.LogInformation($"Round Type: {round.CurrentRoundType}");
            switch (round.CurrentRoundType) {
                case GameRoundType.RoundTypeCompleted:
                    break;
                case GameRoundType.RoundTypeAnte:
                    logger.LogInformation("You Ante");
                    await bot.AnteAsync();
                    break;

                case GameRoundType.RoundTypePreFlop:
                    if (myCurrentBet == round.Bet) {
                        logger.LogInformation("You Check");

                        await bot.CheckAsync();
                    }
                    else if (bot.ChipTotal() > round.Bet * 4) {
                        logger.LogInformation("You Call");
                        await bot.CallAsync();
                    }
                    else {
                        logger.LogInformation("You Fold");
                        await bot.FoldAsync();
                    }

                    break;

                default:
                    if (myCurrentBet == round.Bet) {
                        logger.LogInformation("You Check");

                        await bot.CheckAsync();
                    }
                    else if (bot.ChipTotal() > round.Bet * 3) {
                        logger.LogInformation("You Call");
                        await bot.CallAsync();
                    }
                    else if (bot.ChipTotal() < 400) {
                        logger.LogInformation("You AllIn");
                        await bot.AllInAsync();
                    }
                    else {
                        logger.LogInformation("You Fold");
                        await bot.FoldAsync();
                    }

                    break;
            }
        }
        catch (PokerException ex) {
            logger.LogError(ex, "could not act on this state");
        }
    }

    private sealed record UserConfig {
        [JsonPropertyName("token")] public string Token { get; } = string.Empty;

        [JsonPropertyName("user_id")] public string UserId { get; } = string.Empty;
    }

    private sealed record ExampleOptions(string BaseUrl, string GameId, string User) {
        public static ExampleOptions? Parse(string[] args) {
            string? url = null;
            string? gameId = null;
            string? user = null;

            for (var i = 0; i + 1 < args.Length; i += 2)
                switch (args[i]) {
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

            if (url is null || gameId is null || user is null) return null;

            return new ExampleOptions(url, gameId, user);
        }
    }
}