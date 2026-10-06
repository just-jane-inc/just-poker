package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/just-jane-inc/just-poker/server/just"
	"github.com/mattlangl/gophe"
)

// deck contains an array of [card]
type deck struct {
	cards []card
}

// card encodes an individual playing card using rank and suit runes
type card struct {
	rank rune
	suit rune
}

// AsDTO converts the [card] model to the [CardDTO] type
func (c card) AsDTO() CardDTO {
	return CardDTO{Rank: c.rank, Suit: c.suit}
}

// Burn removes a [card] from the deck without returning it
func (d *deck) Burn() {
	if len(d.cards) == 0 {
		return
	}

	d.cards = d.cards[1:]
}

// Reset remakes and shuffles the deck
func (d *deck) Reset() {
	deck := make([]card, 0)
	for _, suit := range cardSuits {
		for _, rank := range cardRanks {
			deck = append(deck, card{rank: rank, suit: suit})
		}
	}

	rand.Shuffle(len(deck), func(i int, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	d.cards = deck
}

// Draw removes the top card from the top and returns it
func (d *deck) Draw() *card {
	if len(d.cards) == 0 {
		// TODO: this should never happen, but if it does we should not be
		// returning nil, this can lead to a panic. just make the return
		// not a pointer and return a default joker or something in this case
		return nil
	}

	drawnCard := d.cards[0]
	if len(d.cards) == 1 {
		d.cards = make([]card, 0)
	} else {
		d.cards = d.cards[1:]
	}

	return &drawnCard
}

// ToString converts a [card] to a string
func (c *card) ToString() string {
	return (string(c.rank) + string(c.suit))
}

// Hand encodes an arbitrary collection of cards
type Hand struct {
	Cards []*card
}

// GetHandStrings returns an array of [card.ToString] for each card in a [Hand]
func (h Hand) GetHandStrings() []string {
	thisHand := make([]string, len(h.Cards))
	for i, card := range h.Cards {
		thisHand[i] = card.ToString()
	}

	return thisHand
}

// GetHandScore calculates the int rank for a hand with 1 being the best possible hand.
func (h Hand) GetHandScore() (int, error) {
	if len(h.Cards) < 5 || len(h.Cards) > 7 {
		return 0, just.NewPokerError(fmt.Sprintf("hand of len %d cannot be evaluated", len(h.Cards)), just.Unknown)
	}

	var seen uint64
	cards := make([]gophe.Card, len(h.Cards))
	for i, c := range h.Cards {
		card := gophe.NewCard(c.ToString())
		cards[i] = card

		id := card.ID()
		if id >= 52 {
			return 0, just.NewPokerError(fmt.Sprintf("invalid card: %s", c.ToString()), just.Unknown)
		}

		mask := uint64(1) << id
		if seen&mask != 0 {
			return 0, just.NewPokerError(fmt.Sprintf("hand includes duplicate card: %s", c.ToString()), just.Unknown)
		}

		seen |= mask
	}

	rank := gophe.EvaluateCards(cards...)
	return int(rank.GetValue()), nil
}

// CompareTo compares two poker hands (of 5 or 7 cards) to determine which is better
//
// returns 1 if h is better then other, returns 0 if the hands are equal and -1 if other is better
func (h Hand) CompareTo(other Hand) int {
	thatScore, err := other.GetHandScore()
	if err != nil {
		just.Logger.Errorf("encountered error when getting score for hand: %v", err)
		return 0
	}

	thisScore, err := h.GetHandScore()
	if err != nil {
		just.Logger.Errorf("encountered error when getting score for hand: %v", err)
		return 0
	}

	if thisScore < thatScore {
		return 1
	} else if thisScore > thatScore {
		return -1
	}

	return 0
}
