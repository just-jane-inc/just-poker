using JustPoker.OpenApi.Api;
using JustPoker.OpenApi.Client;
using JustPoker.OpenApi.Extensions;
using JustPoker.OpenApi.Model;
using JustPoker.Sdk.Models;
using Microsoft.Extensions.DependencyInjection;

// ReSharper disable NullableWarningSuppressionIsUsed

namespace JustPoker.Sdk;

public static class PokerHelpers {
    public static int ChipSum(IDictionary<string, int> stack) {
        return stack.Sum(entry => int.Parse(entry.Key) * entry.Value);
    }

    public static int ChipSum(IEnumerable<Chips> chips) {
        return chips.Sum(c => c.Value);
    }

    public static List<Chips> ConvertStack(IDictionary<string, int> stack) {
        return stack.Select(entry => new Chips(int.Parse(entry.Key), entry.Value)).ToList();
    }

    public static Dictionary<string, int> ConvertChips(IEnumerable<Chips> chips) {
        var stack = new Dictionary<string, int>();
        foreach (var c in chips) {
            var key = c.Denomination.ToString();
            stack[key] = stack.GetValueOrDefault(key, 0) + c.Count;
        }
        return stack;
    }

    public static ServiceProvider CreateApiProvider(string baseUrl, string token = "") {
        if (string.IsNullOrEmpty(baseUrl)) throw new PokerException("base url not provided");

        var services = new ServiceCollection();
        services.AddLogging();
        services.AddApi(config => {
            config.AddApiHttpClients(client => client.BaseAddress = new Uri(baseUrl));

            if (!string.IsNullOrEmpty(token)) config.AddTokens(new BearerToken(token));

            // Disabled as the need is no longer there. But maintained in case swagger changes and we need quick testing.
            // config.ConfigureJsonOptions(options =>
            //     options.Converters.Insert(0, new NullTolerantModelConverterFactory()));
        });

        return services.BuildServiceProvider();
    }

    public static JustErrorDTO? GetError<TResponse>(TResponse response)
        where TResponse : IBadRequest<JustResponseMessageJustErrorDTO?> {
        if (response.IsSuccessStatusCode || !response.TryBadRequest(out var message)) return null;

        return message.Type == "error" ? message.Data : null;
    }

    public static void ThrowError<TResponse>(TResponse response, string action)
        where TResponse : IBadRequest<JustResponseMessageJustErrorDTO?> {
        if (response.IsSuccessStatusCode) return;

        var error = GetError(response);
        throw new PokerException(error is null
            ? $"{action} failed with {(int)response.StatusCode}: {response.RawContent}"
            : $"{action} failed [{error.ErrorCode}]: {error.Error}");
    }

    public static async Task<GameGameDTO?> GetGameStateAsync(IGameApi api, string gameId) {
        var response = await api.GameGameIdStateGetAsync(gameId);
        ThrowError(response, "get game state");
        return response.TryOk(out var message) ? message.Data : null;
    }

    public static async Task<string?> CreateGameFromConfigAsync(string baseUrl, string token,
        GameNewGameConfigDTO config) {
        await using var provider = CreateApiProvider(baseUrl, token);
        var api = provider.GetRequiredService<IGameApi>();

        var response = await api.GamePostAsync(config);
        ThrowError(response, "create game");
        return response.TryOk(out var message) ? message.Data : null;
    }

    public static Task<string?> CreateGameAsync(string baseUrl, string token, int bigBlind = 100,
        int smallBlind = 50, IDictionary<string, int>? startingChips = null, int playerCount = 5,
        bool autoStartHands = true, IList<int>? denominations = null) {
        startingChips ??= new Dictionary<string, int> {
            // Default
            ["10"] = 10,
            ["50"] = 5,
            ["100"] = 2,
            ["500"] = 1
        };

        denominations ??= [10, 50, 100, 500]; // Default

        var config = new GameNewGameConfigDTO(autoStartHands, bigBlind,
            denominations.ToList(), playerCount, smallBlind,
            new Dictionary<string, int>(startingChips));

        return CreateGameFromConfigAsync(baseUrl, token, config);
    }

    public static async Task StartGameAsync(IGameApi api, string gameId) {
        var response = await api.GameGameIdStartedPostAsync(gameId);
        ThrowError(response, "start game");
    }

    public static async Task DeleteUserAsync(string baseUrl, string token) {
        await using var provider = CreateApiProvider(baseUrl, token);
        var api = provider.GetRequiredService<IUserApi>();

        var response = await api.UserMeDeleteAsync();
        ThrowError(response, "delete user");
    }
}