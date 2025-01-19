package game

import "sort"

func evaluateHand(flop []Card, playerHand []Card, turn Card, river Card) string {
}

func checkPair(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	cards := append(playerHand, append(flop, turn, river)...)

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank {
				count++
			}
		}
		if count == 2 {
			return true, cards[i].Rank
		}
	}
	return false, 0

}

func checkTwoPair(flop []Card, playerHand []Card, turn Card, river Card) (bool, int, int) {

	pairs := []int{}
	usedPairs := map[int]bool{}
	cards := append(playerHand, append(flop, turn, river)...)

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank && !usedPairs[cards[i].Rank] {
				count++
			}
		}
		if count == 2 {
			pairs = append(pairs, cards[i].Rank)
			usedPairs[cards[i].Rank] = true
		}
	}

	if len(pairs) >= 2 {
		return true, pairs[0], pairs[1]
	}
	return false, 0, 0

}

func checkTrips(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	trips := []int{}
	usedTrips := map[int]bool{}
	cards := append(playerHand, append(flop, turn, river)...)

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank && !usedTrips[cards[i].Rank] {
				count++
			}
		}
		if count == 3 {
			trips = append(trips, cards[i].Rank)
			usedTrips[cards[i].Rank] = true
		}
	}

	if len(trips) != 1 {
		return false, 0
	}

	if len(trips) == 1 {
		return true, trips[0]
	}
	return false, 0
}

func checkStraight(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	cards := append(playerHand, append(flop, turn, river)...)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	nonPair := []int{}
	for i := 0; i < len(cards); i++ {
		if i == 0 || cards[i].Rank != cards[i-1].Rank {
			nonPair = append(nonPair, cards[i].Rank)
			if cards[i].Rank == 14 {
				nonPair = append(nonPair, 1)
			}
		}
	}

	sort.Slice(nonPair, func(i, j int) bool {
		return nonPair[i] > nonPair[j]
	})

	count := 1
	highStraight := 0
	for i := 0; i < len(nonPair)-1; i++ {
		if nonPair[i] == nonPair[i+1]+1 {
			count++
			if count == 5 {
				highStraight = nonPair[i]
				return true, highStraight
			}
		} else {
			count = 1
		}

	}
	return false, 0

}

func checkFlush(flop []Card, playerHand []Card, turn Card, river Card) (bool, string, int) {
	cards := append(playerHand, append(flop, turn, river)...)

	spades := 0
	hearts := 0
	diamonds := 0
	clubs := 0

	for i := 0; i < len(cards); i++ {
		suit := cards[i].Suit

		switch suit {
		case "Spades":
			spades++
		case "Hearts":
			hearts++
		case "Diamonds":
			diamonds++
		case "Clubs":
			clubs++
		}
	}

	var flushSuit string
	if spades >= 5 {
		flushSuit = "Spades"
	} else if hearts >= 5 {
		flushSuit = "Hearts"
	} else if diamonds >= 5 {
		flushSuit = "Diamonds"
	} else if clubs >= 5 {
		flushSuit = "Clubs"
	} else {
		return false, "", 0
	}

	flushCards := []int{}
	for i := 0; i < len(cards); i++ {
		if cards[i].Suit == flushSuit {
			flushCards = append(flushCards, cards[i].Rank)
		}
	}

	sort.Slice(flushCards, func(i, j int) bool {
		return flushCards[i] > flushCards[j]
	})

	return true, flushSuit, flushCards[0]

}

func checkBoat(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkQuads(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkSF(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkRoyal(flop []Card, playerHand []Card, turn Card, river Card) bool {

}
