package game

import (
	"math/rand/v2"

	"github.com/just-jane-inc/just-poker/server/just"
)

type deck struct {
	cards []card
}

type card struct {
	rank rune
	suit rune
}

func (c card) AsDTO() CardDTO {
	return CardDTO{Rank: c.rank, Suit: c.suit}
}

func (d *deck) Burn() {
	if len(d.cards) == 0 {
		return
	}

	d.cards = d.cards[1:]
}

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

func (d *deck) Set(cards []card) {
	d.cards = cards
}

func (d *deck) Draw() *card {
	if len(d.cards) == 0 {
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

func (c *card) ToString() string {
	return (string(c.rank) + string(c.suit))
}

type Hand struct {
	Cards []*card
}

func (h Hand) GetHandStrings() []string {
	thisHand := make([]string, len(h.Cards))
	for i, card := range h.Cards {
		thisHand[i] = card.ToString()
	}

	return thisHand
}

// CompareTo compares two poker hands (of 5 or 7 cards) to determine which is better
//
// returns 1 if h is better then other, returns 0 if the hands are equal and -1 if other is better
func (h Hand) CompareTo(other Hand) int {
	thisHand := make([]just.Card, len(h.Cards))
	for i, card := range h.Cards {
		thisHand[i] = just.Card{Rank: card.rank, Suit: card.suit}
	}

	thatHand := make([]just.Card, len(other.Cards))
	for i, card := range other.Cards {
		thatHand[i] = just.Card{Rank: card.rank, Suit: card.suit}
	}

	thatScore, err := just.GetHandScore(thatHand)
	if err != nil {
		just.Logger.Errorf("encountered error when getting score for hand: %v", err)
		return 0
	}

	thisScore, err := just.GetHandScore(thisHand)
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
