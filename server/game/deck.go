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
		d.cards = d.cards[1:]
	}
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

func (this Hand) GetHandStrings() []string {
	thisHand := make([]string, len(this.Cards))
	for i, card := range this.Cards {
		thisHand[i] = card.ToString()
	}

	return thisHand
}

// this > that -> 1
// this == that -> 0
// this < that -> -1
func (this Hand) CompareTo(that Hand) int {
	thisHand := make([]just.Card, len(this.Cards))
	for i, card := range this.Cards {
		thisHand[i] = just.Card{Rank: card.rank, Suit: card.suit}
	}

	thatHand := make([]just.Card, len(that.Cards))
	for i, card := range that.Cards {
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
	} else {
		return 0
	}
}
