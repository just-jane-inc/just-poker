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
    ///     Try to calculate if you can cover a bet for an amount needed. Will recursively do chip exchanges to try.
    ///     If you cannot cover, but have chips, you can treat your bet as an all in
    /// </summary>
    /// <param name="amountNeeded"></param>
    /// <returns></returns>
    public Task<List<Chips>> TryCoverBetAsync(int amountNeeded) {
        _logger.LogDebug("try cover bet for [{Amount}] stack: [{Stack}]", amountNeeded,
            string.Join(" ", _currentStack));
        return TryCoverBetAsync(amountNeeded, []);
    }

    private async Task<List<Chips>> TryCoverBetAsync(int amountNeeded, List<Chips> bet) {
        if (amountNeeded <= 0) return bet;

        var denominations = CurrentState?.GameConfig?.ChipDenominations;
        if (denominations is null || denominations.Count == 0)
            throw new PokerException("try cover bet invoked before a state was received");

        if (ChipTotal() <= amountNeeded) {
            foreach (var chips in _currentStack.Where(chips => chips.Count > 0)) {
                AddToBet(bet, chips.Denomination, chips.Count);
                chips.Count = 0;
            }

            return bet;
        }

        foreach (var chips in _currentStack.OrderByDescending(chips => chips.Denomination)) {
            if (chips.Count == 0) continue;

            var take = Math.Min(chips.Count, amountNeeded / chips.Denomination);
            if (take == 0) continue;

            _logger.LogDebug("taking {Take}x{Denomination} from stack for bet", take, chips.Denomination);
            chips.Count -= take;
            AddToBet(bet, chips.Denomination, take);
            amountNeeded -= take * chips.Denomination;

            if (amountNeeded == 0) return bet;
        }

        foreach (var chips in _currentStack.OrderByDescending(chips => chips.Denomination)) {
            if (chips.Count <= 0) continue;
            if (chips.Denomination <= amountNeeded) break;

            _logger.LogDebug("exchanging 1x{Denomination} for smaller chips", chips.Denomination);
            var broken = BreakChip(chips.Denomination, denominations);
            await ExchangeChipsAsync([new Chips(chips.Denomination, 1)], broken);
            chips.Count -= 1;

            foreach (var brokenChips in broken) MergeStack(brokenChips);
            return await TryCoverBetAsync(amountNeeded, bet);
        }

        throw new InvalidBetException(
            $"could not construct a bet covering {amountNeeded} from [{string.Join(" ", _currentStack)}]");
    }

    private static void AddToBet(List<Chips> bet, int denomination, int count) {
        foreach (var chips in bet)
            if (chips.Denomination == denomination) {
                chips.Count += count;
                return;
            }

        bet.Add(new Chips(denomination, count));
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
        raiseTo -= currentBet;
        var bet = await TryCoverBetAsync(raiseTo);

        var stack = PokerHelpers.ConvertChips(bet);
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
        var bet = await TryCoverBetAsync(amount);
        var stack = PokerHelpers.ConvertChips(bet);
        try {
            return await SendActionAsync(GamePlayerIntent.PlayerIntentAnte, stack);
        }
        catch (PokerException e) {
            _logger.LogError(e.Message);
            // TODO @Jane Game ID 1879, "Wolf" had 1x50 left, and was failing big blind requires 100 chips with the above.
            // Happend in python SDK as well. This is a server problem again I think
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
        var bet = await TryCoverBetAsync(amount);
        var stack = PokerHelpers.ConvertChips(bet);
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

    private List<Chips> BreakChip(int value, IList<int> denominations) {
        List<Chips> result = [];
        var remaining = value;

        foreach (var denomination in denominations.OrderDescending()) {
            if (denomination >= value)
                continue;

            var count = remaining / denomination;
            remaining %= denomination;

            if (count > 0)
                result.Add(new Chips(denomination, count));
            if (remaining == 0)
                break;
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