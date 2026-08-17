namespace JustPoker.Sdk.Models;

public sealed class Chips(int denomination, int count) {
    public int Denomination { get; } = denomination;
    public int Count { get; set; } = count;
    public int Value => Denomination * Count;

    public override string ToString() {
        return $"{Count}x{Denomination}";
    }
}