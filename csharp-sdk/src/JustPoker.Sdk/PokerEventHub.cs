using System.Runtime.CompilerServices;
using System.Threading.Channels;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;

namespace JustPoker.Sdk;

public delegate Task PokerEventHandler(PokerEvent pokerEvent);

public sealed class EventSubscription : IDisposable {
    private readonly CancellationTokenSource _cancellation = new();
    private readonly Channel<PokerEvent> _channel;
    private PokerEventHub? _hub;

    internal EventSubscription(PokerEventHub hub, IReadOnlyList<PokerEventType> eventTypes,
        PokerEventHandler? handler, bool inline, ILogger logger, Func<PokerEvent, bool>? predicate = null) {
        _hub = hub;
        EventTypes = eventTypes;
        Handler = handler;
        IsInline = inline;
        Predicate = predicate;
        IsActive = true;

        _channel = Channel.CreateUnbounded<PokerEvent>(new UnboundedChannelOptions {
            SingleReader = true,
            SingleWriter = true
        });

        if (handler is not null && !inline) Consumer = ConsumeAsync(handler, logger, _cancellation.Token);
    }

    public IReadOnlyList<PokerEventType> EventTypes { get; }
    public PokerEventHandler? Handler { get; }
    public bool IsInline { get; }
    public Func<PokerEvent, bool>? Predicate { get; }
    public bool IsActive { get; private set; }

    internal ChannelReader<PokerEvent> Reader => _channel.Reader;
    internal Task? Consumer { get; }

    public void Dispose() {
        Unsubscribe();
    }

    public bool Unsubscribe() {
        if (!IsActive)
            return false;

        IsActive = false;
        _channel.Writer.TryComplete();
        _cancellation.Cancel();

        var hub = _hub;
        _hub = null;
        hub?.Remove(this);
        return true;
    }

    public override string ToString() {
        return
            $"EventSubscription({string.Join('|', EventTypes.Select(t => t.GetText()))}, {(IsActive ? "active" : "unsubscribed")})";
    }

    internal bool Accepts(PokerEvent pokerEvent) {
        return Predicate is null || Predicate(pokerEvent);
    }

    internal bool Publish(PokerEvent pokerEvent) {
        return _channel.Writer.TryWrite(pokerEvent);
    }

    private async Task ConsumeAsync(PokerEventHandler handler, ILogger logger, CancellationToken cancellationToken) {
        try {
            await foreach (var pokerEvent in _channel.Reader.ReadAllAsync(cancellationToken))
                try {
                    await handler(pokerEvent);
                }
                catch (OperationCanceledException) {
                    throw;
                }
                catch (Exception e) {
                    logger.LogError(e, "error in handler {Subscription}", this);
                }
        }
        catch (OperationCanceledException) {
        }
    }
}

/// <summary>
///     Hub for consuming events from the poker server, with full subscription modelling for event types
/// </summary>
/// <param name="transport"></param>
/// <param name="name"></param>
/// <param name="logger"></param>
public sealed class PokerEventHub(IEventTransport transport, string name = "event-hub", ILogger? logger = null)
    : IAsyncDisposable {
    private static readonly TimeSpan DrainTime = TimeSpan.FromSeconds(2);
    private readonly Lock _gate = new();
    private readonly ILogger _logger = logger ?? NullLogger.Instance;

    private readonly string _name = name; // TODO for logging help
    private readonly SemaphoreSlim _startGate = new(1);
    private readonly Dictionary<PokerEventType, List<EventSubscription>> _subscriptions = new();

    private TaskCompletionSource _closed = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private Channel<PokerEvent> _ingest = CreateIngestChannel();
    private Task? _pump;
    private CancellationTokenSource? _pumpCancellation;
    private Task? _reader;

    public bool Running => _pump is { IsCompleted: false };

    public CloseInfo? CloseInfo { get; private set; }

    public IEventTransport Transport => transport;

    public async ValueTask DisposeAsync() {
        await StopAsync();

        List<EventSubscription> remaining;
        lock (_gate) {
            remaining = _subscriptions.Values.SelectMany(handlers => handlers).Distinct().ToList();
        }

        foreach (var subscription in remaining) subscription.Unsubscribe();

        await transport.DisposeAsync();
    }

    public static PokerEventHub PokerWebSocket(
        string baseUrl,
        string token,
        string gameId,
        WebSocketStreamOptions? options = null,
        ILogger? logger = null) {
        options ??= new WebSocketStreamOptions();

        var stream = new WebSocketStream(baseUrl, token, gameId, options, logger);
        var hub = new PokerEventHub(stream, $"event-hub[{gameId}]", logger);

        var existing = options.OnClose;
        options.OnClose = info => {
            existing?.Invoke(info);
            hub.NoteClose(info);
        };

        return hub;
    }

    public EventSubscription Subscribe(PokerEventType eventType, PokerEventHandler handler,
        Func<PokerEvent, bool>? predicate = null) {
        return Subscribe([eventType], handler, predicate);
    }

    public EventSubscription Subscribe(IEnumerable<PokerEventType> eventTypes, PokerEventHandler handler,
        Func<PokerEvent, bool>? predicate = null) {
        return Register(eventTypes, handler, false, predicate);
    }

    public EventSubscription Subscribe(PokerEventType eventType, Action<PokerEvent> handler,
        Func<PokerEvent, bool>? predicate = null) {
        return Subscribe(eventType, pokerEvent => {
            handler(pokerEvent);
            return Task.CompletedTask;
        }, predicate);
    }

    internal EventSubscription SubscribeInline(IEnumerable<PokerEventType> eventTypes, PokerEventHandler handler,
        Func<PokerEvent, bool>? predicate = null) {
        return Register(eventTypes, handler, true, predicate);
    }

    internal EventSubscription SubscribeInline(PokerEventType eventType, PokerEventHandler handler,
        Func<PokerEvent, bool>? predicate = null) {
        return Register([eventType], handler, true, predicate);
    }

    public int SubscriberCount(PokerEventType? eventType = null) {
        lock (_gate) {
            if (eventType is null) return _subscriptions.Values.Sum(handlers => handlers.Count);

            return _subscriptions.TryGetValue(eventType.Value, out var found) ? found.Count : 0;
        }
    }

    public async Task<PokerEventHub> StartAsync(CancellationToken cancellationToken = default) {
        if (Running) return this;

        await _startGate.WaitAsync(cancellationToken);

        try {
            if (Running) return this;

            if (_closed.Task.IsCompleted)
                _closed = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);

            await transport.ConnectAsync(cancellationToken);

            _ingest = CreateIngestChannel();
            _pumpCancellation = new CancellationTokenSource();
            _reader = ReadAsync(_ingest.Writer, _pumpCancellation.Token);
            _pump = PumpAsync(_ingest.Reader, _pumpCancellation.Token);
            return this;
        }
        finally {
            _startGate.Release();
        }
    }

    public async Task StopAsync() {
        await transport.CloseAsync();

        var reader = _reader;
        var pump = _pump;
        var cancellation = _pumpCancellation;
        _reader = null;
        _pump = null;
        _pumpCancellation = null;

        var running = Task.WhenAll(new[] { reader, pump }.Where(task => task is not null)!);

        if (await Task.WhenAny(running, Task.Delay(DrainTime)) != running && cancellation is not null)
            await cancellation.CancelAsync();

        try {
            await running;
        }
        catch (OperationCanceledException) {
        }

        cancellation?.Dispose();
        _closed.TrySetResult();
    }

    public async Task<PokerEvent> WaitForAsync(PokerEventType eventType, Func<PokerEvent, bool>? predicate = null,
        TimeSpan? timeout = null, CancellationToken cancellationToken = default) {
        var completion = new TaskCompletionSource<PokerEvent>(TaskCreationOptions.RunContinuationsAsynchronously);

        using var subscription = Subscribe(eventType, pokerEvent => {
            completion.TrySetResult(pokerEvent);
            return Task.CompletedTask;
        }, predicate);

        await StartAsync(cancellationToken);

        return timeout is null
            ? await completion.Task.WaitAsync(cancellationToken)
            : await completion.Task.WaitAsync(timeout.Value, cancellationToken);
    }

    public async Task<CloseInfo?> WaitClosedAsync(TimeSpan? timeout = null,
        CancellationToken cancellationToken = default) {
        if (timeout is null)
            await _closed.Task.WaitAsync(cancellationToken);
        else
            await _closed.Task.WaitAsync(timeout.Value, cancellationToken);

        return CloseInfo;
    }

    public async IAsyncEnumerable<PokerEvent> StreamAsync(PokerEventType eventType = PokerEventType.All,
        Func<PokerEvent, bool>? predicate = null,
        [EnumeratorCancellation] CancellationToken cancellationToken = default) {
        using var subscription = Register([eventType], null, false, predicate);
        using var closedCancellation = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);

        _ = _closed.Task.ContinueWith(
            _ => {
                try {
                    // ReSharper disable once AccessToDisposedClosure
                    closedCancellation.Cancel();
                }
                catch (ObjectDisposedException) {
                }
            },
            CancellationToken.None,
            TaskContinuationOptions.ExecuteSynchronously,
            TaskScheduler.Default);

        await StartAsync(closedCancellation.Token);

        while (true) {
            var draining = false;

            try {
                if (!await subscription.Reader.WaitToReadAsync(closedCancellation.Token)) break;
            }
            catch (OperationCanceledException) {
                draining = true;
            }

            while (subscription.Reader.TryRead(out var pokerEvent)) yield return pokerEvent;

            if (draining) break;
        }
    }

    public async Task RunForeverAsync(CancellationToken cancellationToken = default) {
        await transport.ConnectAsync(cancellationToken);

        _ingest = CreateIngestChannel();
        await Task.WhenAll(ReadAsync(_ingest.Writer, cancellationToken), PumpAsync(_ingest.Reader, cancellationToken));
    }

    internal void Remove(EventSubscription subscription) {
        lock (_gate) {
            foreach (var key in subscription.EventTypes) {
                if (!_subscriptions.TryGetValue(key, out var handlers)) continue;

                handlers.Remove(subscription);

                if (handlers.Count == 0) _subscriptions.Remove(key);
            }
        }
    }

    internal void NoteClose(CloseInfo info) {
        CloseInfo = info;
    }

    private EventSubscription Register(IEnumerable<PokerEventType> eventTypes, PokerEventHandler? handler, bool inline,
        Func<PokerEvent, bool>? predicate = null) {
        var keys = eventTypes.Distinct().ToArray();
        if (keys.Length == 0) throw new PokerException("subscribe needs at least one event type");

        var subscription = new EventSubscription(this, keys, handler, inline, _logger, predicate);

        lock (_gate) {
            foreach (var key in keys) {
                if (!_subscriptions.TryGetValue(key, out var handlers)) {
                    handlers = [];
                    _subscriptions[key] = handlers;
                }

                handlers.Add(subscription);
            }
        }

        return subscription;
    }

    private static Channel<PokerEvent> CreateIngestChannel() {
        return Channel.CreateUnbounded<PokerEvent>(new UnboundedChannelOptions {
            SingleReader = true,
            SingleWriter = true
        });
    }

    private async Task ReadAsync(ChannelWriter<PokerEvent> ingest, CancellationToken cancellationToken) {
        try {
            await foreach (var pokerEvent in transport.EventsAsync(cancellationToken))
                if (!ingest.TryWrite(pokerEvent))
                    _logger.LogWarning("dropped event on {Hub} ingest", _name);
        }
        catch (OperationCanceledException) {
        }
        catch (Exception e) {
            _logger.LogError(e, "event feed failed on {Hub}", _name);
            CloseInfo ??= new CloseInfo(null, e.Message);
        }
        finally {
            ingest.TryComplete();
        }
    }

    private async Task PumpAsync(ChannelReader<PokerEvent> ingest, CancellationToken cancellationToken) {
        try {
            await foreach (var pokerEvent in ingest.ReadAllAsync(cancellationToken)) await DispatchAsync(pokerEvent);
        }
        catch (OperationCanceledException) {
        }
        catch (Exception e) {
            _logger.LogError(e, "event dispatch failed on {Hub}", _name);
            CloseInfo ??= new CloseInfo(null, e.Message);
        }
        finally {
            _closed.TrySetResult();
        }
    }

    private async Task DispatchAsync(PokerEvent pokerEvent) {
        List<EventSubscription> targets;

        lock (_gate) {
            targets = [];

            if (_subscriptions.TryGetValue(PokerEventType.All, out var wildcard)) targets.AddRange(wildcard);

            if (pokerEvent.EventType != PokerEventType.All &&
                _subscriptions.TryGetValue(pokerEvent.EventType, out var handlers))
                targets.AddRange(handlers);
        }

        foreach (var subscription in targets.Where(s =>
                     s is { IsActive: true, IsInline: true, Handler: not null } && s.Accepts(pokerEvent)))
            try {
                await subscription.Handler?.Invoke(pokerEvent)!;
            }
            catch (OperationCanceledException) {
                throw;
            }
            catch (Exception e) {
                _logger.LogError(e, "error in inline handler {Subscription} on {Hub}", subscription, _name);
            }

        foreach (var subscription in targets.Where(s =>
                     s is { IsActive: true, IsInline: false } && s.Accepts(pokerEvent)))
            if (!subscription.Publish(pokerEvent))
                _logger.LogWarning("dropped event for {Subscription} on {Hub}", subscription, _name);
    }
}