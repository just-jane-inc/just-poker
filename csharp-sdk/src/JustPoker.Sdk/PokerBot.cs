using JustPoker.OpenApi.Api;
using JustPoker.OpenApi.Model;
using JustPoker.Sdk.Enums;
using JustPoker.Sdk.Models;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;

#pragma warning disable CS1591

namespace JustPoker.Sdk;

public sealed class PokerBot : IAsyncDisposable {
    private const int MaxChipDowns = 16;

    private readonly string _baseUrl;
    private readonly ILogger _logger;

    private readonly ServiceProvider _provider;
    private readonly TimeSpan? _timeout;
    private readonly string _token;
    private List<Chips> _currentStack = [];
    private PokerEventHub? _hub;

    private EventSubscription? _stateSubscription;


    /// <summary>
    ///     A poker bot
    /// </summary>
    /// <param name="baseUrl"></param>
    /// The url of the game server to connect to
    /// <param name="token"></param>
    /// the user authorization token to use for this bot
    /// <param name="userId"></param>
    /// the user id of the requested user for this bot
    /// <param name="gameId"></param>
    /// the id of the game to interact with
    /// <param name="timeout"></param>
    /// the timeout time (as timespan) the bot should wait for its turn when performing an action. Null or 0 to disable.
    /// <param name="logger"></param>
    /// <exception cref="PokerException"></exception>
    public PokerBot(string baseUrl, string token, string userId, string gameId,
        TimeSpan? timeout = null, ILogger? logger = null) {
        if (string.IsNullOrEmpty(token)) throw new PokerException("api token not provided");

        if (string.IsNullOrEmpty(baseUrl)) throw new PokerException("base url not provided");

        if (string.IsNullOrEmpty(gameId)) throw new PokerException("game id not provided");

        if (string.IsNullOrEmpty(userId)) throw new PokerException("user id not provided");

        _baseUrl = baseUrl;
        _token = token;
        UserId = userId;
        GameId = gameId;
        _timeout = timeout;
        _logger = logger ?? NullLogger.Instance;

        _provider = PokerHelpers.CreateApiProvider(baseUrl, token);
        GameApi = _provider.GetRequiredService<IGameApi>();
        UserApi = _provider.GetRequiredService<IUserApi>();
    }

    public string GameId { get; }

    public string UserId { get; }

    public bool Joined { get; private set; }

    public GameGameDTO? CurrentState { get; private set; }

    public GamePlayerDTO? Player { get; private set; }

    public IGameApi GameApi { get; }

    public IUserApi UserApi { get; }

    public IReadOnlyList<Chips> CurrentStack => _currentStack;

    public IReadOnlyList<Card> HoleCards => Player?.Hole?.Select(Card.FromDto).ToList() ?? [];

    /// <summary>
    ///     Get the reference to the PokerEventHub configured for this bot, creates if DNE
    /// </summary>
    public PokerEventHub Events {
        get {
            if (_hub is not null) return _hub;

            _hub = PokerEventHub.PokerWebSocket(_baseUrl, _token, GameId, logger: _logger);
            _stateSubscription?.Unsubscribe();

            // Subscribe in advance to these events to always have our internal model up to date
            _stateSubscription = _hub.SubscribeInline(
                [PokerEventType.Welcome, PokerEventType.GameStateUpdate, PokerEventType.StartingGame],
                IngestEventAsync);

            return _hub;
        }
    }

    public async ValueTask DisposeAsync() {
        _stateSubscription?.Unsubscribe();

        if (_hub is not null) await _hub.StopAsync();

        await _provider.DisposeAsync();
    }

    public int ChipTotal() {
        var sum = 0;
        foreach (var chip in CurrentStack) sum += chip.Value;
        return sum;
    }

    /// <summary>
    ///     Join the configured game for this bot, noops if already joined
    /// </summary>
    public async Task JoinGameAsync() {
        if (Joined)
            return;
        var response = await GameApi.GameGameIdPlayerPostAsync(GameId);
        PokerHelpers.ThrowError(response, "join game");
        Joined = true;
    }

    /// <summary>
    ///     Starts the configured game for this bot
    /// </summary>
    public async Task StartGameAsync() {
        var response = await GameApi.GameGameIdStartedPostAsync(GameId);
        PokerHelpers.ThrowError(response, "start game");
    }

    /// <summary>
    ///     Manually fetch the game state for this bot's game
    /// </summary>
    /// <returns></returns>
    public async Task<GameGameDTO?> GetGameStateAsync() {
        var response = await GameApi.GameGameIdStateGetAsync(GameId);
        PokerHelpers.ThrowError(response, "get game state");

        if (!response.TryOk(out var msg) || msg.Data is null) return null;
        IngestGameDto(msg.Data);
        return msg.Data;
    }


    public Task<PokerEventHub> StartEventsAsync() {
        return Events.StartAsync();
    }


    public Task StopEventsAsync() {
        if (_hub is null) throw new PokerException("no event hub has been started");

        return _hub.StopAsync();
    }

    /// <summary>
    ///     Exchange chips held by the bot with the game's exchange
    /// </summary>
    /// <param name="give"></param>
    /// the chips the bot is giving to the server exchange
    /// <param name="receive"></param>
    /// the chips the bot wants to receive as a result of the exchange
    /// <exception cref="PokerException"></exception>
    /// Raises if the exchange is determined invalid by the server
    public async Task ExchangeChipsAsync(IEnumerable<Chips> give, IEnumerable<Chips> receive) {
        var giveStack = new Dictionary<string, int>();
        var receiveStack = new Dictionary<string, int>();
        foreach (var chip in give) giveStack[chip.Denomination.ToString()] = chip.Count;
        foreach (var chip in receive) receiveStack[chip.Denomination.ToString()] = chip.Count;

        var req = new GameChipExchangeDTO(giveStack, receiveStack);
        var response =
            await GameApi.GameGameIdChipExchangePostAsync(GameId, req);
        PokerHelpers.ThrowError(response, "chip exchange");
    }

    /// <summary>
    ///     Merges the provided chips into the bot's current stack locally
    /// </summary>
    /// <param name="chips"></param>
    public void MergeStack(Chips chips) {
        foreach (var chip in _currentStack)
            if (chip.Denomination == chips.Denomination) {
                chip.Count += chips.Count;
                return;
            }

        _currentStack.Add(chips);
    }

    /// <summary>
    ///     computes a valid bet from a set of available chips and denominations to satisfy a required amount
    /// </summary>
    /// <param name="amount"></param>
    /// the amount we need to bet
    /// <returns></returns>
    /// a mapping of string to integer expressing the bet that the player can make to satisfy the provided amount.
    /// <remarks>
    ///     Constraints:
    ///     - let D = self._current_state.game_config.chip_denominations
    ///     - let C = self._player.stack.keys()
    ///     - ∀d ∈ D, d > 0             (all denominations are positive integers)
    ///     - C ⊆ D                     (denominations appearing in players stack are strictly a subset of those in the game)
    ///     - ∀k,d ∈ D, d lt k gte d∣k      (for all k,d in denominations k less than d implies k divides d evenly)
    ///     - ∃d ∈ D such that d∣amount (for the provided amount to bet their exists some element of denominations which divdes
    ///     it evenly)
    ///     Notes:
    ///     1. The constraints described above are not validated in this code and their violation represents undefined behavior
    ///     2. This funcion will alter the state of the game - it can preform a single chip exchange if one is required
    ///     3. If the provided amount is greater then the sum of all chips the player has we simply return all of their chips -
    ///     "all in"
    /// </remarks>
    /// <exception cref="PokerException"></exception>
    /// <exception cref="InvalidBetException"></exception>
    public async Task<Dictionary<string, int>> ComputeValidBetAsync(int amount) {
        if (CurrentState?.GameConfig?.ChipDenominations is not { Count: > 0 } chipDenominations)
            throw new PokerException("compute valid bet invoked before a state was received");

        if (Player?.Stack is null)
            throw new PokerException("compute valid bet invoked before a player was received");

        var chips = Player.Stack.ToDictionary(entry => int.Parse(entry.Key), entry => entry.Value);

        if (chips.Sum(entry => entry.Key * entry.Value) < amount)
            return chips.ToDictionary(entry => entry.Key.ToString(), entry => entry.Value);

        var denominations = chipDenominations.OrderByDescending(d => d).ToList();

        var validBet = new Dictionary<int, int>();
        foreach (var denomination in denominations) {
            if (denomination > amount) continue;

            var take = amount / denomination;
            amount -= take * denomination;
            validBet[denomination] = take;

            if (amount == 0) break;
        }

        if (amount != 0)
            throw new InvalidBetException("no valid bet can be constructed - constraint violation likely");

        var missingChips = new Dictionary<int, int>();
        foreach (var (denomination, count) in validBet) {
            if (!chips.TryGetValue(denomination, out var available)) {
                missingChips[denomination] = count;
                continue;
            }

            if (available < count) {
                missingChips[denomination] = count - available;
                chips[denomination] = 0;
            }
            else {
                chips[denomination] = available - count;
            }
        }

        var give = denominations.ToDictionary(d => d, _ => 0);
        var receive = denominations.ToDictionary(d => d, _ => 0);

        foreach (var (denomination, missing) in missingChips.OrderBy(entry => entry.Key)) {
            var count = missing;
            if (count == 0) break;

            foreach (var held in chips.Keys.OrderByDescending(d => d).ToList()) {
                if (count <= 0) break;

                if (held == denomination || chips[held] == 0) continue;

                if (held > denomination) {
                    var exchangeRate = held / denomination;
                    while (chips[held] > 0 && count > 0) {
                        chips[held] -= 1;
                        give[held] += 1;
                        receive[denomination] += exchangeRate;
                        count -= exchangeRate;
                    }
                }
                else {
                    var exchangeRate = denomination / held;
                    while (chips[held] >= exchangeRate && count > 0) {
                        chips[held] -= exchangeRate;
                        give[held] += exchangeRate;
                        receive[denomination] += 1;
                        count -= 1;
                    }
                }
            }
        }

        if (give.Sum(entry => entry.Key * entry.Value) > 0)
            try {
                await ExchangeChipsAsync(
                    give.Select(entry => new Chips(entry.Key, entry.Value)).ToList(),
                    receive.Select(entry => new Chips(entry.Key, entry.Value)).ToList());
            }
            catch (Exception) {
                _logger.LogError(
                    "encountered error exchanging chips | receive=[{Receive}] | give=[{Give}] | bet=[{Bet}] | stack=[{Stack}]",
                    ChipDictionaryToString(receive), ChipDictionaryToString(give), ChipDictionaryToString(validBet), string.Join(" ", _currentStack));
                throw;
            }

        return validBet.ToDictionary(entry => entry.Key.ToString(), entry => entry.Value);
    }

    private static string ChipDictionaryToString(IDictionary<int, int> chips) {
        return string.Join(" ", chips.Select(entry => $"{entry.Value}x{entry.Key}"));
    }

    private Dictionary<int, int> Held() {
        var held = new Dictionary<int, int>();
        foreach (var chips in _currentStack)
            held[chips.Denomination] = held.GetValueOrDefault(chips.Denomination, 0) + chips.Count;
        return held;
    }

    /// <summary>
    ///     Sends the check action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> CheckAsync() {
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        return await SendActionAsync(GamePlayerIntent.PlayerIntentCheck, new Dictionary<string, int>());
    }

    /// <summary>
    ///     Sends the all in action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> AllInAsync() {
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        return await SendActionAsync(GamePlayerIntent.PlayerIntentAllIn, new Dictionary<string, int>());
    }

    /// <summary>
    ///     Sends the raise action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> RaiseAsync(int raiseTo) {
        if (Player is null) return false;
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        var currentBet = PokerHelpers.ChipSum(Player.CurrentBet ?? []);
        var amount = raiseTo - currentBet;

        var stack = await ComputeValidBetAsync(amount);
        return await SendActionAsync(GamePlayerIntent.PlayerIntentRaise, stack);
    }

    /// <summary>
    ///     Sends the ante action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> AnteAsync() {
        if (Player is null) return false;
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        var amount = CurrentState?.Table?.CurrentRound?.Bet ?? 0;
        if (amount <= 0) throw new PokerException("erm, no bet?");

        var currentBet = PokerHelpers.ChipSum(Player.CurrentBet ?? []);
        amount -= currentBet;
        var stack = await ComputeValidBetAsync(amount);
        try {
            return await SendActionAsync(GamePlayerIntent.PlayerIntentAnte, stack);
        }
        catch (PokerException e) {
            _logger.LogError(e.Message);
            throw;
        }
    }

    /// <summary>
    ///     Sends the call action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> CallAsync() {
        if (Player is null) return false;
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        var amount = CurrentState?.Table?.CurrentRound?.Bet ?? 0;
        if (amount <= 0) throw new PokerException("erm, no bet?");

        var currentBet = PokerHelpers.ChipSum(Player.CurrentBet ?? []);
        amount -= currentBet;
        var stack = await ComputeValidBetAsync(amount);
        return await SendActionAsync(GamePlayerIntent.PlayerIntentCall, stack);
    }

    /// <summary>
    ///     Sends the fold action after waiting for the bot's turn
    /// </summary>
    /// <returns></returns>
    public async Task<bool> FoldAsync() {
        if (Player is null) return false;
        await WaitForMyTurnAsync();
        if (!IsMyTurn()) return false;

        return await SendActionAsync(GamePlayerIntent.PlayerIntentFold);
    }

    /// <summary>
    ///     Sends the requested action to the bot's joined game
    /// </summary>
    /// <returns></returns>
    public async Task<bool> SendActionAsync(GamePlayerIntent intent, Dictionary<string, int>? bet = null) {
        if (bet is null) bet = new Dictionary<string, int>();
        var dto = new GamePlayerActionDTO {
            Chips = bet,
            Intent = intent
        };

        var response = await GameApi.GameGameIdActionPostAsync(GameId, dto);
        PokerHelpers.ThrowError(response, $"action {intent}");
        return true;
    }

    /// <summary>
    ///     Blocks until either the bot's turn arrives or the timeout (if provided) expires.
    /// </summary>
    /// <param name="cancellationToken"></param>
    /// <exception cref="PokerException"></exception>
    /// Raises if the bot is not in a joined game
    public async Task WaitForMyTurnAsync(CancellationToken cancellationToken = default) {
        if (!Joined) throw new PokerException("you are not in the game - it can never be your turn");

        if (!Events.Running) await StartEventsAsync();

        using var cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);

        if (_timeout is { } timeout && timeout > TimeSpan.Zero)
            cts.CancelAfter(timeout);

        var token = cts.Token;
        while (!token.IsCancellationRequested && !IsMyTurn()) {
            token.ThrowIfCancellationRequested();
            try {
                await Task.Delay(100, token);
            }
            catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested) {
                throw new PokerException("your turn never came after waiting");
            }
        }
    }

    /// <summary>
    ///     Is the current player position of the round equal to your bot's seat position
    /// </summary>
    /// <returns></returns>
    public bool IsMyTurn() {
        return Player?.Position == CurrentPlayerPosition();
    }

    private List<Chips> BreakChip(int value, IList<int> denominations, bool needOffDenom = true) {
        List<Chips> result = [];
        var remaining = value;

        var pool = denominations.Where(d => d < value).OrderByDescending(d => d).ToList();
        if (!needOffDenom) {
            var step = denominations.Min();
            var onDenom = pool.Where(d => d % step == 0).ToList();
            if (onDenom.Count > 0) pool = onDenom;
        }

        foreach (var denomination in pool) {
            var count = remaining / denomination;
            if (count > 0) {
                result.Add(new Chips(denomination, count));
                remaining %= denomination;
            }

            if (remaining == 0) break;
        }

        if (remaining != 0) {
            _logger.LogError($"{value} can not be subdivided with provided denominations: {denominations}");
            throw new PokerException("invalid chip exchange request with current denominations");
        }

        return result;
    }

    private void IngestGameDto(GameGameDTO? state) {
        if (state is null) return;

        foreach (var gamePlayerDto in state.Table?.Players ?? [])
            if (gamePlayerDto.UserId == UserId) {
                Joined = true;
                Player = gamePlayerDto;
                _currentStack = PokerHelpers.ConvertStack(Player?.Stack ?? new Dictionary<string, int>());
                break;
            }

        CurrentState = state;
    }

    private Task IngestEventAsync(PokerEvent pokerEvent) {
        if (pokerEvent.DataAs<GameGameDTO>() is { } state) IngestGameDto(state);

        return Task.CompletedTask;
    }

    private int? CurrentPlayerPosition() {
        return CurrentState?.Table?.CurrentRound?.CurrentPlayerPosition;
    }

    public async Task<PokerBot> InitializeAsync() {
        await StartEventsAsync();
        return this;
    }
}