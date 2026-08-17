using System.Threading.Channels;
using JustPoker.Sdk.Enums;
using Microsoft.Extensions.Logging;

namespace JustPoker.Sdk.Models;

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